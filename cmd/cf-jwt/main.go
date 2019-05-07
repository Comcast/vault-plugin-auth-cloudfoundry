package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

/*
export CF_INSTANCE_KEY=/Users/raymondharrison/IdeaProjects/vault-plugin-auth-cloudfoundry/cmd/vault-plugin-auth-cloudfoundry/instance/instance.key
export CF_INSTANCE_CERT=/Users/raymondharrison/IdeaProjects/vault-plugin-auth-cloudfoundry/cmd/vault-plugin-auth-cloudfoundry/instance/instance.crt
*/

type jwt_record struct {
	cert     *x509.Certificate
	policies []string
}

func main() {

	roles := os.Args[1:]

	instanceKey := os.Getenv("CF_INSTANCE_KEY")
	instanceCert := os.Getenv("CF_INSTANCE_CERT")

	if len(instanceKey) == 0 {
		log.Fatal("The environment variable CF_INSTANCE_KEY is not set. This contains the cloud foundry instance private key.")
	}

	if len(instanceKey) == 0 {
		log.Fatal("The environment variable CF_INSTANCE_CERT is not set. This contains the cloud foundry instance public key.")
	}

	pkey, _ := getPrivateKeyFromLocation(instanceKey)

	certs, _ := getPublicKeyFromLocation(instanceCert)

	// Build the JWT

	key := jose.SigningKey{Algorithm: jose.RS256, Key: pkey}

	jwk := jose.JSONWebKey{
		Key:          pkey,
		KeyID:        "cf",
		Algorithm:    "RS256",
		Certificates: certs,
	}

	var signerOpts = jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signerOpts.WithHeader("JWK", jwk)

	rsaSigner, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		log.Fatal("Jose RSA Signing Key Error")
	}

	c2 := struct {
		Policies []string
	}{
		roles,
	}

	cl := jwt.Claims{
		NotBefore: jwt.NewNumericDate(certs[0].NotBefore),
	}

	raw, err := jwt.Signed(rsaSigner).Claims(cl).Claims(c2).CompactSerialize()
	if err != nil {
		log.Fatal(err.Error())
	}

	errWrite := ioutil.WriteFile("/tmp/jwt", []byte(raw), 0644)

	if errWrite != nil {
		log.Fatal(errWrite.Error())
	}

	jwtPolicyRecord, err2 := parseCertificateFromJWT2(raw)
	if err2 != nil {
		fmt.Printf("\nErr2 %s\n", err2.Error())
	}

	fmt.Printf("Policy Record %s", contains(jwtPolicyRecord.policies, "read-only"))

}

func getPublicKeyFromLocation(s string) ([]*x509.Certificate, error) {
	keyFile, err := os.Open(s)
	if err != nil {
		return nil, err
	}

	defer keyFile.Close()

	contentBytes, err := ioutil.ReadAll(keyFile)
	if err != nil {
		return nil, err
	}

	// This will be a file with multiple certificates
	sdata := string(contentBytes)

	var blocks []byte

	rest := []byte(sdata)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, errors.New("Error: PEM not parsed")

		}
		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}

	cert, err := x509.ParseCertificates(blocks)

	if err != nil {
		return nil, errors.New(err.Error())
	}

	if len(cert) == 0 {
		return nil, errors.New("Certificate Array Length 0")
	}
	return cert, nil

}

func getPrivateKeyFromLocation(location string) (*rsa.PrivateKey, error) {

	keyFile, err := os.Open(location)
	if err != nil {
		return nil, err
	}

	defer keyFile.Close()

	contentBytes, err := ioutil.ReadAll(keyFile)

	if err != nil {
		return nil, err
	}

	keyText := string(contentBytes)

	block, _ := pem.Decode([]byte(keyText))

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic("failed to parse RSA key: " + err.Error())
	}
	return key, nil
}

func parseCertificateFromJWT2(raw string) (*jwt_record, error) {
	tok, err := jwt.ParseSigned(raw)

	jwtr := new(jwt_record)

	if err != nil {
		return nil, err
	}

	policyClaims := struct {
		Policies []string
	}{}

	hdrs := tok.Headers
	if len(hdrs) != 1 {
		return nil, errors.New("Incorrect header length")
	}

	jwk := hdrs[0].ExtraHeaders["JWK"]

	jwkMap := jwk.(map[string]interface{})

	x5c := jwkMap["x5c"]

	x5cArray := x5c.([]interface{})

	x5cSigningKey := x5cArray[0].(string)

	b64bytes, errb64 := fromBase64Bytes(x5cSigningKey)
	if errb64 != nil {
		return nil, errb64
	}

	cert, errpc := x509.ParseCertificate(b64bytes)

	if errpc != nil {
		return nil, errpc
	}

	if cert.KeyUsage&x509.KeyUsageKeyAgreement == 0 {
		return nil, errors.New("Invalid cert key usage")
	}

	out := jwt.Claims{}
	if err := tok.Claims(cert.PublicKey, &out, &policyClaims); err != nil {
		return nil, err
	}

	if len(policyClaims.Policies) == 0 {
		return nil, errors.New("No roles defines")
	}

	jwtr.cert = cert
	jwtr.policies = policyClaims.Policies

	fmt.Printf("\nInner Record %s\n", jwtr.policies)

	return jwtr, nil
}

func fromBase64Bytes(b64 string) ([]byte, error) {
	re := regexp.MustCompile(`\s+`)
	val, err := base64.StdEncoding.DecodeString(re.ReplaceAllString(b64, ""))
	if err != nil {
		return nil, errors.New("Invalid certificate data or incorrect format")
	}
	return val, nil
}

func contains(slice []string, str string) bool {

	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}
	_, ok := set[str]
	return ok

}
