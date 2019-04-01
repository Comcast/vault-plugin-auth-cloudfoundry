.PHONY: certs
certs: clean certs/rootCA.pem certs/instance.pem certs/instance-invalid.pem

clean:
	rm -rf \
		certs && \
	mkdir certs

##
## Root CA
##

certs/rootCA.key:
	openssl genrsa \
		-out $@ \
		1028

certs/rootCA.pem: certs/rootCA.key
	openssl req \
		-config certs.cfg \
		-x509 \
		-new \
		-nodes \
		-key certs/rootCA.key \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:ca' \
		-extensions root \
		-sha256 \
		-days 1024 \
		-out $@

##
## Instance cert
##

certs/instance.key.pem:
	openssl genrsa \
		-out certs/instance.key \
		1028
	openssl pkcs8 \
		-topk8 \
		-nocrypt \
		-inform PEM \
		-outform PEM \
		-in certs/instance.key \
		-out $@

certs/instance.csr: certs/instance.key.pem
	openssl req \
		-newkey rsa:2048 \
		-key certs/instance.key \
		-sha256 \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:instance/OU=organization:org-uuid/OU=space:space-uuid/OU=app:app-uuid' \
		-nodes \
		-in certs/rootCA.key \
		-out $@ \
		-outform PEM

certs/instance.pem: certs/instance.csr
	openssl x509 \
		-req \
		-extfile certs.cfg \
		-in certs/instance.csr \
		-CA certs/rootCA.pem \
		-CAkey certs/rootCA.key \
		-CAcreateserial \
		-days 1024 \
		-out $@ \
		-extensions instance \
		-sha256

##
## Invalid instance cert
## See the certs.cfg file for more details
##

certs/instance-invalid.key.pem:
	openssl genrsa \
		-out certs/instance-invalid.key \
		1028
	openssl pkcs8 \
		-topk8 \
		-nocrypt \
		-inform PEM \
		-outform PEM \
		-in certs/instance-invalid.key \
		-out $@

certs/instance-invalid.csr: certs/instance-invalid.key.pem
	openssl req \
		-newkey rsa:2048 \
		-key certs/instance-invalid.key \
		-sha256 \
		-subj '/CN=vault-plugin-auth-cloudfoundry:DEV:instance/OU=organization:org-uuid/OU=space:space-uuid/OU=app:app-uuid' \
		-nodes \
		-in certs/rootCA.key \
		-out $@ \
		-outform PEM

certs/instance-invalid.pem: certs/instance-invalid.csr
	openssl x509 \
		-req \
		-extfile certs.cfg \
		-in certs/instance-invalid.csr \
		-CA certs/rootCA.pem \
		-CAkey certs/rootCA.key \
		-CAcreateserial \
		-days 1024 \
		-out $@ \
		-extensions missing-keyusage \
		-sha256
