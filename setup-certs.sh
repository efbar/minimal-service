#! /usr/bin/env bash

PASSWORD="password"
DOMAIN_NAME="minimal-service.org"
rm -rf certs || true

mkdir certs

# create new CA, out: a ca key (private key of the ca)
openssl req -new -newkey rsa:2048 -days 365 -x509 -sha256 -subj "/CN=minimal-service,OU=Cloud,O=MinimalServices,L=Milan,ST=Italy,C=IT" -keyout "certs/ca.key" -out "certs/ca.pem" -nodes

cat > "certs/openssl.cnf" << EOF
[req]
default_bits = 2048
encrypt_key  = no # Change to encrypt the private key using des3 or similar
default_md   = sha256
prompt       = no
utf8         = yes
####
[ req_distinguished_name ]
countryName                     = IT
countryName_min                 = 2
countryName_max                 = 2
countryName_default             = IT
stateOrProvinceName             = Italy
stateOrProvinceName_default     = Italy
localityName                    = Milan
localityName_default            = Milan
organizationalUnitName          = Cloud
organizationalUnitName_default  = Cloud
organizationName                = MinimalServices
organizationName_default        = MinimalServices
commonName                      = minimal-service.org
commonName_max                  = 64
commonName_default              = minimal-service.org
emailAddress                    = admin@minimal-service.org
emailAddress_max                = 64
emailAddress_default            = admin@minimal-service.org
####
# Extensions for SAN IP and SAN DNS
req_extensions = v3_req
# Allow client and server auth. You may want to only allow server auth.
# Link to SAN names.
[v3_req]
basicConstraints     = CA:FALSE
subjectKeyIdentifier = hash
keyUsage             = digitalSignature, keyEncipherment
extendedKeyUsage     = clientAuth, serverAuth
subjectAltName       = @alt_names
# Alternative names are specified as IP.# and DNS.# for IP addresses and
# DNS accordingly. 
[alt_names]
DNS.1 = 127.0.0.1
DNS.2 = *.${DOMAIN_NAME}
EOF

openssl genrsa -out certs/minimalservice.key 2048

openssl req -new -key certs/minimalservice.key -out certs/minimalservice.csr -subj "/CN=minimal-service,OU=Cloud,O=MinimalServices,L=Milan,ST=Italy,C=IT"

openssl x509 -req -in certs/minimalservice.csr -CA certs/ca.pem -CAkey certs/ca.key -CAcreateserial \
-out certs/minimalservice.crt -days 3650 -sha256 -extfile certs/openssl.cnf

