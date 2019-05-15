package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func frameworkPathConfig(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: pathConfig + "$",
		Fields: map[string]*framework.FieldSchema{
			"ca": {
				Type:        framework.TypeString,
				Description: "PEM encoded CA cert that CF will use to sign instance certificates",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathConfigWrite,
			logical.CreateOperation: b.pathConfigWrite,
			logical.ReadOperation:   b.pathConfigRead,
		},

		HelpSynopsis: `Configures the CF public key and other plugin information.`,
		HelpDescription: `The Cloud Foundry Auth backend uses a root certificate authority
		to validate authentication requests against instance certificates signed by the same CA.`,
	}
}

type config struct {
	CA       string         `json:"ca"`
	CertPool *x509.CertPool `json:"-"`
}

// pathConfigWrite handles create and update commands to the config
func (b *backend) pathConfigWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	certData := data.Get("ca").(string)
	if certData == "" {
		return nil, errors.New("no CA provided")
	}

	block, _ := pem.Decode([]byte(certData))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid certificate block type %q", block.Type)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate PEM: %v", err)
	}
	if !cert.IsCA {
		return nil, errors.New("certificate is not CA")
	}

	cfg := config{
		CA: certData,
	}

	entry, err := logical.StorageEntryJSON(pathConfig, cfg)
	if err != nil {
		return nil, err
	}
	if err = req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}

// pathConfigRead handles read commands to the config
func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := b.config(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	resp := logical.Response{
		Data: map[string]interface{}{
			"ca": cfg.CA,
		},
	}

	return &resp, nil
}

// config pulls and unmarshals the config from the storage backend.
func (b *backend) config(ctx context.Context, s logical.Storage) (*config, error) {
	raw, err := s.Get(ctx, pathConfig)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("config does not exist")
	}

	var cfg config
	if err = raw.DecodeJSON(&cfg); err != nil {
		return nil, err
	}

	if cfg.CA == "" {
		return &cfg, nil
	}

	// reconstitute certificate pool

	block, _ := pem.Decode([]byte(cfg.CA))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate PEM: %v", err)
	}
	if !cert.IsCA {
		return nil, errors.New("certificate is not CA")
	}
	cfg.CertPool = x509.NewCertPool()
	cfg.CertPool.AddCert(cert)

	return &cfg, nil
}
