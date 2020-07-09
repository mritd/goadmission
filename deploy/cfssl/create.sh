#!/usr/bin/env bash

set -ex

rm -f *.pem *.csr

cfssl gencert --initca=true dac-ca-csr.json | cfssljson --bare dac-ca
cfssl gencert --ca dac-ca.pem --ca-key dac-ca-key.pem --config dac-gencert.json dac-csr.json | cfssljson --bare dac

kubectl create secret generic \
        dynamic-admission-control-certs \
        --namespace kube-addons \
        --from-file=dac.pem \
        --from-file=dac-key.pem

CA_BUNDLE=$(cat dac-ca.pem | base64 | tr -d '\n')
sed -i "s@\${CA_BUNDLE}@${CA_BUNDLE}@g" ../*/*.yaml