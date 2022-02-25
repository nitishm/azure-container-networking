# Generating a self-signed certificate with SANs

This section explains how to generate a self-signed certificate with SANs.

> To generate the certs please use `generate_certs.sh` script.

## Generating a local CA

```bash
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca.key -out ca.crt -subj "/C=US/ST=Washington/L=Redmond/O=Microsoft/OU=Azure/CN=azure-npm.kube-system.svc.cluster.local"
```

## Generate a certificate request for the service

```bash
openssl req -newkey rsa:4096 -nodes -keyout tls.key -out server-req.pem -config san.cnf -extensions v3_req -subj "/C=US/ST=Washington/L=Redmond/O=Microsoft/OU=Azure/CN=azure-npm.kube-system.svc.cluster.local"         
```

## Sign the certificate request with the CA

```bash
openssl x509 -req -in server-req.pem -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt --days 365 --extfile san.cnf --extensions v3_req
```

## Move certs/key files to the kustomize

The base includes the CA certificate which will be added to both the controller and daemon pods.

```bash
mv ca.crt kustomize/base
```

The controller is the only component that requires the server certificate and key.

```bash
mv tls.crt kustomize/controller
mv tls.key kustomize/controller
```
