#!/bin/bash

go build witness
openssl genpkey -algorithm ed25519 -outform PEM -out testkey.pem
openssl pkey -in testkey.pem -pubout > testpub.pem

witness sign -k testkey.pem policy.json > policy.signed.json
witness run -s build -r https://log.testifysec.io -k testkey.pem --trace -o attestation1.json -- go build

./witness verify -k testpub.pem -a attestation1.json -p policy.signed.json -f witness #local test
./witness verify -k testpub.pem -r https://log.testifysec.io -p policy.signed.json -f witness #remote test
