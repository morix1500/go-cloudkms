package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

var (
	projectId string = os.Getenv("PROJECT")
	keyring   string = os.Getenv("KEYRING")
	keyname   string = os.Getenv("KEYNAME")
	bucket    string = os.Getenv("BUCKET")
)

func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{
		outStream: outStream,
		errStream: errStream,
	}
	args := strings.Split("cloudkms --version", " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := fmt.Sprintf("cloudkms %s\n", Version)

	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, errStream.String(), "wrong message")
}

func TestRun_List(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)

	cli := &CLI{
		outStream: outStream,
		errStream: errStream,
	}
	cmd := fmt.Sprintf("cloudkms --list --key_bucket %s", bucket)
	args := strings.Split(cmd, " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := `fuga.txt
hoge.txt
piyo.txt
`

	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, outStream.String(), "wrong message")
}

func TestRun_Get(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)

	cli := &CLI{
		outStream: outStream,
		errStream: errStream,
	}
	cmd := fmt.Sprintf("cloudkms --get --key_bucket %s --project_id %s --keyring %s --keyname %s --decrypt_path hoge.txt", bucket, projectId, keyring, keyname)
	args := strings.Split(cmd, " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := `Download hoge.txt
`
	// check file exists
	_, err := os.Stat("hoge.txt")

	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, outStream.String(), "wrong message")
	assert.Equal(t, nil, err, "not exists hoge.txt")
}
