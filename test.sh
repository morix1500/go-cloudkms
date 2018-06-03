#!/bin/bash

TEST_BUCKET=morix-go-cloudkms-test
PROJECT=$(gcloud config get-value project -q)
LOCATION=asia-northeast1
KEYRING=test
KEYNAME=testkey
GCS_KEY_PREFIX=kms-keys
API_KEY=$(gcloud auth print-access-token)

function encrypt() {
  filename=$1

  res=$(curl -s -X POST "https://cloudkms.googleapis.com/v1/projects/${PROJECT}/locations/${LOCATION}/keyRings/${KEYRING}/cryptoKeys/${KEYNAME}:encrypt" \
    -d "{\"plaintext\":\"$(cat ${filename})\"}" \
    -H "Authorization:Bearer ${API_KEY}" \
    -H "Content-Type:application/json")

  ciphertext=$(echo ${res} | jq -r ".ciphertext")
  echo -n ${ciphertext} > ${filename}.encrypted
}

function init() {
  gsutil mb gs://${TEST_BUCKET}

  gcloud kms keyrings create ${KEYRING} --location ${LOCATION}
  gcloud kms keys create ${KEYNAME} --keyring ${KEYRING} --location ${LOCATION} --purpose encryption
}

#init 2>&1 /dev/null

echo "hoge" | base64 > hoge.txt
echo "fuga" | base64 > fuga.txt
echo "piyo" | base64 > piyo.txt

encrypt hoge.txt 2>&1 /dev/null
encrypt fuga.txt 2>&1 /dev/null
encrypt piyo.txt 2>&1 /dev/null

gsutil cp *.encrypted gs://${TEST_BUCKET}/${GCS_KEY_PREFIX}
