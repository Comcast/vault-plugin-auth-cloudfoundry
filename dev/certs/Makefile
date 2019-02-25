.PHONY: all
all: rootCA.pem instance.pem instance-invalid.pem

clean:
	rm -f \
		instance.csr \
		instance.key \
		instance.key.pem \
		instance.pem \
		rootCA.pem \
		rootCA.key \
		rootCA.srl

##
## Root CA
##

rootCA.key:
	openssl genrsa \
		-out $@ \
		1028

rootCA.pem: rootCA.key
	openssl req \
		-config config/ca.cfg \
		-x509 \
		-new \
		-nodes \
		-key rootCA.key \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:ca' \
		-extensions root \
		-sha256 \
		-days 1024 \
		-out $@

##
## Instance cert
##

instance.key.pem:
	openssl genrsa \
		-out instance.key \
		1028
	openssl pkcs8 \
		-topk8 \
		-nocrypt \
		-inform PEM \
		-outform PEM \
		-in instance.key \
		-out $@

instance.csr: instance.key.pem
	openssl req \
		-newkey rsa:2048 \
		-key instance.key \
		-sha256 \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:instance/OU=organization:org-uuid/OU=space:space-uuid/OU=app:app-uuid' \
		-nodes \
		-in rootCA.key \
		-out $@ \
		-outform PEM

instance.pem: instance.csr
	openssl x509 \
		-req \
		-extfile config/ca.cfg \
		-in instance.csr \
		-CA rootCA.pem \
		-CAkey rootCA.key \
		-CAcreateserial \
		-days 1024 \
		-out $@ \
		-extensions instance \
		-sha256

##
## Invalid instance cert
##

instance-invalid.key.pem:
	openssl genrsa \
		-out instance-invalid.key \
		1028
	openssl pkcs8 \
		-topk8 \
		-nocrypt \
		-inform PEM \
		-outform PEM \
		-in instance-invalid.key \
		-out $@

instance-invalid.csr: instance-invalid.key.pem
	openssl req \
		-newkey rsa:2048 \
		-key instance-invalid.key \
		-sha256 \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:instance/OU=organization:org-uuid/OU=space:space-uuid/OU=app:app-uuid' \
		-nodes \
		-in rootCA.key \
		-out $@ \
		-outform PEM

instance-invalid.pem: instance-invalid.csr
	openssl x509 \
		-req \
		-extfile config/ca.cfg \
		-in instance-invalid.csr \
		-CA rootCA.pem \
		-CAkey rootCA.key \
		-CAcreateserial \
		-days 1024 \
		-out $@ \
		-extensions missing-keyusage \
		-sha256
