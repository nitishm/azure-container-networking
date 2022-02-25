#!/bin/bash

REPO_ROOT=$(git rev-parse --show-toplevel)
echo "REPO_ROOT: $REPO_ROOT"
KUSTOMIZE_ROOT=$REPO_ROOT/npm/deploy/kustomize
echo "KUSTOMIZE_ROOT: $KUSTOMIZE_ROOT"
CERTS_STAGING_DIR=$REPO_ROOT/npm/deploy/certs
echo "CERTS_STAGING_DIR: $CERTS_STAGING_DIR"
SAN_CNF_FILE=$CERTS_STAGING_DIR/san.cnf
echo "SAN_CNF_FILE: $SAN_CNF_FILE"
CERTIFICATE_VALIDITY_DAYS=3650

# Check if openssl is installed
if ! command -v openssl &> /dev/null
then
    echo "openssl could not be found"
    exit
fi

# Check if SAN_CNF_FILE exists
if [ ! -f "$SAN_CNF_FILE" ]
then
		echo "SAN_CNF_FILE does not exist"
		exit
fi

# Generate the ca certificate and key
openssl req -x509 -newkey rsa:4096 -days $CERTIFICATE_VALIDITY_DAYS -nodes -keyout $CERTS_STAGING_DIR/ca.key -out $CERTS_STAGING_DIR/ca.crt -subj "/C=US/ST=Washington/L=Redmond/O=Microsoft/OU=Azure/CN=azure-npm.kube-system.svc.cluster.local"

# Create a certificate signing request for the server
openssl req -newkey rsa:4096 -nodes -keyout $CERTS_STAGING_DIR/tls.key -out $CERTS_STAGING_DIR/server-req.pem -config $SAN_CNF_FILE -extensions v3_req -subj "/C=US/ST=Washington/L=Redmond/O=Microsoft/OU=Azure/CN=azure-npm.kube-system.svc.cluster.local" 

# Sign the server certificate with the CA
openssl x509 -req -in server-req.pem -CA $CERTS_STAGING_DIR/ca.crt -CAkey $CERTS_STAGING_DIR/ca.key -CAcreateserial -out $CERTS_STAGING_DIR/tls.crt --days $CERTIFICATE_VALIDITY_DAYS --extfile $SAN_CNF_FILE --extensions v3_req

# Move the generated files to the correct location
mv $CERTS_STAGING_DIR/ca.crt $KUSTOMIZE_ROOT/base

mv $CERTS_STAGING_DIR/tls.crt $KUSTOMIZE_ROOT/overlays/controller
mv $CERTS_STAGING_DIR/tls.key $KUSTOMIZE_ROOT/overlays/controller

echo "Certificates generated and moved to kustomize/base and kustomize/overlays/controller"
