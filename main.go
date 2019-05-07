package main

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/vault/helper/pluginutil"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/hashicorp/vault/logical/plugin"
)

const (
	pathLogin  = "login"
	pathConfig = "config"
)

type jwt_record struct {
	cert     *x509.Certificate
	policies []string
}

func main() {
	apiClientMeta := &pluginutil.APIClientMeta{}

	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:]) // Ignore command, strictly parse flags

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := pluginutil.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, err
	}
	return b, nil
}

type backend struct {
	*framework.Backend

	orgMap   *framework.PolicyMap
	spaceMap *framework.PolicyMap
}

func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.orgMap = &framework.PolicyMap{
		PathMap: framework.PathMap{
			Name: "organizations",
		},
	}

	b.spaceMap = &framework.PolicyMap{
		PathMap: framework.PathMap{
			Name: "spaces",
		},
	}

	b.Backend = &framework.Backend{
		BackendType: logical.TypeCredential,
		AuthRenew:   b.pathAuthRenew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{pathLogin},
			SealWrapStorage: []string{pathConfig},
		},
		Paths: []*framework.Path{
			&framework.Path{
				Pattern: pathLogin,
				Fields: map[string]*framework.FieldSchema{
					"jwt": &framework.FieldSchema{
						Type: framework.TypeString,
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.UpdateOperation: b.pathAuthLogin,
				},
			},
			frameworkPathConfig(&b),
		},
	}

	b.Paths = append(b.Paths, b.orgMap.Paths()...)
	b.Paths = append(b.Paths, b.spaceMap.Paths()...)

	return &b
}

func (b *backend) pathAuthLogin(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	raw := d.Get("jwt").(string)
	if raw == "" {
		return logical.ErrorResponse("missing jwt"), nil
	}
	//cert, err := parseCertificateFromJWT(raw)

	jwtParsedRecord, err := parseCertificateFromJWT2(raw)

	cert := jwtParsedRecord.cert
	//policies := jwtParsedRecord.policies

	if err != nil {
		return logical.ErrorResponse(err.Error()), nil
	}

	var cf struct {
		org   string
		space string
		app   string
	}

	for _, ou := range cert.Subject.OrganizationalUnit {
		kv := strings.SplitN(ou, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "organization":
			cf.org = kv[1]
		case "space":
			cf.space = kv[1]
		case "app":
			cf.app = kv[1]
		}
	}

	if cf.org == "" || cf.space == "" || cf.app == "" {
		return logical.ErrorResponse("missing CF cert organizational units"), nil
	}

	b.Logger().Info("certificate", "org", cf.org, "space", cf.space, "app", cf.app)

	cfg, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if cfg.CertPool == nil {
		return nil, errors.New("no CA configured")
	}

	opts := x509.VerifyOptions{
		Roots: cfg.CertPool,
	}

	if _, err = cert.Verify(opts); err != nil {
		return nil, err
	}

	orgPolicies, err := b.orgMap.Policies(ctx, req.Storage, cf.org)

	/*
	       for _, policy := range policies {
	            containsOrgPolicy := contains(orgPolicies, policy)
	   	}

	*/

	if err != nil {
		return nil, err
	}
	spacePolicies, err := b.spaceMap.Policies(ctx, req.Storage, cf.space)
	if err != nil {
		return nil, err
	}

	if len(orgPolicies) == 0 && len(spacePolicies) == 0 {
		return logical.ErrorResponse(fmt.Sprintf("no policies mapped to org:%s space:%s", cf.org, cf.space)), nil
	}

	certTTL := time.Until(cert.NotAfter)

	resp := logical.Response{
		Auth: &logical.Auth{
			Policies: append(orgPolicies, spacePolicies...),
			Metadata: map[string]string{
				"cf_org":   cf.org,
				"cf_space": cf.space,
				"cf_app":   cf.app,
			},
			LeaseOptions: logical.LeaseOptions{
				Renewable: true,
				TTL:       certTTL,
			},
		},
	}

	return &resp, nil
}

func (b *backend) pathAuthRenew(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	// TODO
	return nil, nil
}

// The JWT logic

func parseCertificateFromJWT(raw string) (*x509.Certificate, error) {
	tok, err := jwt.ParseSigned(raw)

	if err != nil {
		return nil, err
	}

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

	return cert, nil
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
		return nil, errors.New("No roles defined")
	}

	jwtr.cert = cert
	jwtr.policies = policyClaims.Policies

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
