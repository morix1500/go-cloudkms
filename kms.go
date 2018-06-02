package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	kms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/iterator"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ExitCodeOK = iota
	ExitCodeError
)

type CLI struct {
	outStream, errStream io.Writer
	context              context.Context
	bucket               *storage.BucketHandle
	keyInfo              KeyInfo
}

type KeyInfo struct {
	ProjectId string
	Location  string
	KeyRing   string
	KeyName   string
}

func getGCSBucket(ctx context.Context, bucketName string) (*storage.BucketHandle, error) {
	cli, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return cli.Bucket(bucketName), nil
}

func (c *CLI) List() error {
	objects := c.bucket.Objects(c.context, nil)
	cnt := 0

	for {
		objAttrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		filename := strings.Replace(objAttrs.Name, ".encrypted", "", -1)
		fmt.Fprintf(c.outStream, "%s\n", filename)
		cnt++
	}

	if cnt == 0 {
		return errors.New("The key does not exist")
	}

	return nil
}

func getKMSService(ctx context.Context) (*kms.Service, error) {
	client, err := google.DefaultClient(ctx, kms.CloudPlatformScope)
	if err != nil {
		return nil, err
	}

	return kms.New(client)
}

func (c *CLI) Get(decryptPath string) error {
	obj, err := c.bucket.Object(decryptPath + ".encrypted").NewReader(c.context)
	if err != nil {
		return err
	}
	defer obj.Close()

	filename := filepath.Base(decryptPath)
	w, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(obj); err != nil {
		return err
	}
	ciphertext := buf.String()

	kmsService, err := getKMSService(c.context)
	if err != nil {
		return err
	}

	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", c.keyInfo.ProjectId, c.keyInfo.Location, c.keyInfo.KeyRing, c.keyInfo.KeyName)
	req := &kms.DecryptRequest{
		Ciphertext: ciphertext,
	}
	resp, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Decrypt(parentName, req).Do()
	if err != nil {
		return err
	}
	plain, err := base64.StdEncoding.DecodeString(resp.Plaintext)
	if err != nil {
		return err
	}
	io.Copy(w, bytes.NewReader(plain))
	fmt.Fprintf(c.outStream, "Download %s\n", filename)
	return nil
}

func (c *CLI) Run(args []string) int {
	var version bool
	var list bool
	var get bool

	var projectId string
	var location string
	var keyRing string
	var keyName string
	var keyBucket string
	var decryptPath string

	flags := flag.NewFlagSet("cloudkms", flag.ContinueOnError)
	flags.SetOutput(c.errStream)
	flags.BoolVar(&version, "version", false, "Print version information and quit")
	flags.BoolVar(&list, "list", false, "Show key file list")
	flags.BoolVar(&get, "get", false, "")
	flags.StringVar(&projectId, "project_id", "", "Specific GCP Project ID")
	flags.StringVar(&location, "location", "asia-northeast1", "")
	flags.StringVar(&keyRing, "keyring", "", "")
	flags.StringVar(&keyName, "keyname", "", "")
	flags.StringVar(&keyBucket, "key_bucket", "", "")
	flags.StringVar(&decryptPath, "decrypt_path", "", "")

	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	if version {
		fmt.Fprintf(c.errStream, "cloudkms %s\n", Version)
		return ExitCodeOK
	}

	ctx := context.Background()
	bucket, err := getGCSBucket(ctx, keyBucket)
	if err != nil {
		fmt.Fprintf(c.errStream, err.Error())
		return ExitCodeError
	}
	keyInfo := KeyInfo{
		ProjectId: projectId,
		Location:  location,
		KeyRing:   keyRing,
		KeyName:   keyName,
	}
	c.context = ctx
	c.bucket = bucket
	c.keyInfo = keyInfo

	if list {
		err := c.List()
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
	}

	if get {
		err := c.Get(decryptPath)
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
	}

	return ExitCodeOK
}
