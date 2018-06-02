export PROJECT=$(shell gcloud config get-value project -q)
export KEYRING=test
export KEYNAME=testkey
export BUCKET=morix-go-cloudkms-test

build:
	go build -o cloudkms main.go kms.go

test:
	go test

test-init:
	./test.sh
