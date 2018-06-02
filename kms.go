package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	kms "google.golang.org/api/cloudkms/v1"
	"google.golang.org/api/iterator"
	"io"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/alecthomas/kingpin.v2"
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

func (c *CLI) Put(encryptPath string) error {
	file, err := os.Open(encryptPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		return err
	}

	plaintext := buf.Bytes()

	kmsService, err := getKMSService(c.context)
	if err != nil {
		return err
	}

	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", c.keyInfo.ProjectId, c.keyInfo.Location, c.keyInfo.KeyRing, c.keyInfo.KeyName)
	req := &kms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(plaintext),
	}
	resp, err := kmsService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(parentName, req).Do()
	if err != nil {
		return err
	}

	filename := filepath.Base(encryptPath)

	w := c.bucket.Object(filename + ".encrypted").NewWriter(c.context)
	defer w.Close()

	_, err = w.Write([]byte(resp.Ciphertext))
	if err != nil {
		return err
	}
	fmt.Fprintf(c.outStream, "Upload %s\n", filename)

	return nil
}

func (c *CLI) setup(bucket string, keyInfo KeyInfo) error {
	ctx := context.Background()
	gcsBucket, err := getGCSBucket(ctx, bucket)
	if err != nil {
		return err
	}
	c.context = ctx
	c.bucket = gcsBucket
	c.keyInfo = keyInfo

	return nil
}

func (c *CLI) Run(args []string) int {
	app := kingpin.New("cloudkms", "GCP Cloud KMS Get/Put Command")

	versionCmd := app.Command("version", "Print version")
	// list
	listCmd := app.Command("list", "Output key files")
	listBucket := listCmd.Flag("bucket", "GCS Bucket").String()
	// get
	getCmd    := app.Command("get", "Get key file")
	getBucket    := getCmd.Flag("bucket", "GCS Bucket").String()
	getProjectId := getCmd.Flag("project_id", "GCS Project").String()
	getLocation  := getCmd.Flag("location", "GCS KMS Location").Default("asia-northeast1").String()
	getKeyring   := getCmd.Flag("keyring", "GCS KMS Keyring").String()
	getKeyname   := getCmd.Flag("keyname", "GCS KMS Keyname").String()
	getPath      := getCmd.Flag("path", "key file path").String()
	// put
	putCmd    := app.Command("put", "Put key file")
	putBucket     := putCmd.Flag("bucket", "GCS Bucket").String()
	putProjectId  := putCmd.Flag("project_id", "GCS Project").String()
	putLocation   := putCmd.Flag("location", "GCS KMS Location").Default("asia-northeast1").String()
	putKeyring    := putCmd.Flag("keyring", "GCS KMS Keyring").String()
	putKeyname    := putCmd.Flag("keyname", "GCS KMS Keyname").String()
	putPath       := putCmd.Flag("path", "key file path").String()

	switch kingpin.MustParse(app.Parse(args[1:])) {
	case versionCmd.FullCommand():
		fmt.Fprintf(c.errStream, "cloudkms %s\n", Version)
	case listCmd.FullCommand():
		err := c.setup(*listBucket, KeyInfo{})
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}

		err = c.List()
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
	case getCmd.FullCommand():
		keyInfo := KeyInfo{
			ProjectId: *getProjectId,
			Location:  *getLocation,
			KeyRing:   *getKeyring,
			KeyName:   *getKeyname,
		}
		err := c.setup(*getBucket, keyInfo)
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
		err = c.Get(*getPath)
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
	case putCmd.FullCommand():
		keyInfo := KeyInfo{
			ProjectId: *putProjectId,
			Location:  *putLocation,
			KeyRing:   *putKeyring,
			KeyName:   *putKeyname,
		}
		err := c.setup(*putBucket, keyInfo)
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
		err = c.Put(*putPath)
		if err != nil {
			fmt.Fprintf(c.errStream, err.Error())
			return ExitCodeError
		}
	}

	return ExitCodeOK
}
