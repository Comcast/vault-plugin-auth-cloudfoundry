package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	"log"
	"os"
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
