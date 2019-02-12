package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
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
					"cert": &framework.FieldSchema{
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
	certData := d.Get("cert").(string)
	if certData == "" {
		return logical.ErrorResponse("missing cert"), nil
	}

	block, _ := pem.Decode([]byte(certData))
	if block == nil {
		return logical.ErrorResponse("invalid cert"), nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf("invalid cert: %s", err)), nil
	}

	if cert.KeyUsage&x509.KeyUsageKeyAgreement == 0 {
		return logical.ErrorResponse("invalid cert key usage"), nil
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
