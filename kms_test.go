package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
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
	args := strings.Split("cloudkms version", " ")

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
	cmd := fmt.Sprintf("cloudkms list --bucket %s", bucket)
	args := strings.Split(cmd, " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := `fuga.txt
hoge.txt
piyo.txt
`

	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, outStream.String(), "wrong message: " + errStream.String())
}

func TestRun_Get(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)

	cli := &CLI{
		outStream: outStream,
		errStream: errStream,
	}
	//cmd := fmt.Sprintf("cloudkms get --bucket %s --project_id %s --keyring %s --keyname %s --path hoge.txt", bucket, projectId, keyring, keyname)
	cmd := fmt.Sprintf("cloudkms get hoge.txt --bucket %s --location asia-northeast1 --project_id %s --keyring %s --keyname %s", bucket, projectId, keyring, keyname)
	args := strings.Split(cmd, " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := `Download hoge.txt
`
	// check file exists
	_, err := os.Stat("hoge.txt")

	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, outStream.String(), "wrong message: " + errStream.String())
	assert.Equal(t, nil, err, "not exists hoge.txt")
}

func TestRun_Put(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)

	cli := &CLI{
		outStream: outStream,
		errStream: errStream,
	}

	// create test file
	createCmd := exec.Command("bash", "-c", "echo test > test.txt")
	createCmd.Start()

	cmd := fmt.Sprintf("cloudkms put hoge.txt --bucket %s --location asia-northeast1 --project_id %s --keyring %s --keyname %s", bucket, projectId, keyring, keyname)
	args := strings.Split(cmd, " ")

	status := cli.Run(args)
	expectStatus := 0
	expectMsg := `Upload hoge.txt
`
	assert.Equal(t, expectStatus, status, "wrong status")
	assert.Equal(t, expectMsg, outStream.String(), "wrong message: " + errStream.String())
}
