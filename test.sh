#!/bin/bash
set -x

go build -o bin/witness
openssl genpkey -algorithm ed25519 -outform PEM -out testkey.pem
openssl pkey -in testkey.pem -pubout > testpub.pem

./bin/witness sign -k testkey.pem policy.json > policy.signed.json
./bin/witness run -s build -r https://log.testifysec.io -k testkey.pem -o attestation1.json -- go build

./bin/witness verify -k testpub.pem -a attestation1.json -p policy.signed.json -f witness #local test
./bin/witness verify -k testpub.pem -r https://log.testifysec.io -p policy.signed.json -f witness #remote test
