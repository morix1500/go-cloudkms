export PROJECT=$(shell gcloud config get-value project -q)
export KEYRING=test
export KEYNAME=testkey
export BUCKET=morix-go-cloudkms-test

test:
	go test

test-init:
	./test.sh
