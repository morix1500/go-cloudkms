# cloudkms
It is a command to safely manage secret key and credential file with GCS using CloudKMS.

## Example
```shell
# Login is required at gcloud
$ gcloud auth application-default login

# create sample key file
echo "xxxxxxxxxxxxxxxxx" > api_key.txt

# put key file
$ cloudkms put api_key.txt \
               --bucket keyfiles-gcs-bucket \
               --project sample-111 \
               --keyring sample-keyring \
               --keyname sample-keyring-key
Upload api_key.txt

# get key list
$ cloudkms list --bucket keyfiles-gcs-bucket
service-account-key.json
api_key.txt

# Confirm the contents of the file encrypted by Cloud KMS
$ gsutil cat gs://keyfiles-gcs-bucket/kms-keys/api_key.txt.encrypted
CiQAPX9xtlnCmxixrQipWt2XixqCrMGUaW3caVkEe1QIdRg2Fj0SOwBYHqWMJ0orj3JXWu6203bHHu3cfXPW+dve3zIPlDzzbDrdMv70Q6cRorwAZrY8TY0VdZcXpt3BW6qY%

# get key file
$ export KMS_GCS_BUCKET=keyfiles-gcs-bucket
$ export KMS_PROJECT=sample-111
$ export KMS_KEYRING=sample-keyring
$ export KMS_KEYNAME=sample-keyring-key

$ cloudkms get api_key.txt
Download api_key.txt

$ cat api_key.txt
xxxxxxxxxxxxxxxxx
```

## Usage
```shell
$ cloudkms --help
usage: cloudkms [<flags>] <command> [<args> ...]

GCP Cloud KMS Get/Put Command

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Commands:
  help [<command>...]
    Show help.

  version
    Print version

  list [<flags>]
    Output encryption key files

  get [<flags>] <path>
    Get encryption key file

  put [<flags>] <path>
    Put encryption key file

------------------------------------------

$ cloudkms list --help
usage: cloudkms list [<flags>]

Output encryption key files

Flags:
  --help       Show context-sensitive help (also try --help-long and --help-man).
  --bucket=""  Specify the GCS bucket that stores the encryption key. Configurable with environment
               variable: KMS_GCS_BUCKET

------------------------------------------

$ cloudkms get --help
usage: cloudkms get [<flags>] <path>

Get encryption key file

Flags:
  --help               Show context-sensitive help (also try --help-long and --help-man).
  --bucket=""          Specify the GCS bucket that stores the encryption key. Configurable with
                       environment variable: KMS_GCS_BUCKET
  --project_id=""      GCP Project ID. Configurable with environment variable: KMS_PROJECT
  --location="global"  Region that stored KMS Keyring. Configurable with environment variable:
                       KMS_LOCATION
  --keyring=""         KMS Keyring. Configurable with environment variable: KMS_KEYRING
  --keyname=""         KMS keyring Keyname. Configurable with environment variable: KMS_KEYNAME

Args:
  <path>  Name of the saved encryption key

------------------------------------------

$ cloudkms put --help
usage: cloudkms put [<flags>] <path>

Put encryption key file

Flags:
  --help               Show context-sensitive help (also try --help-long and --help-man).
  --bucket=""          Specify the GCS bucket that stores the encryption key. Configurable with
                       environment variable: KMS_GCS_BUCKET
  --project_id=""      GCP Project ID. Configurable with environment variable: KMS_PROJECT
  --location="global"  Region that stored KMS Keyring. Configurable with environment variable:
                       KMS_LOCATION
  --keyring=""         KMS Keyring. Configurable with environment variable: KMS_KEYRING
  --keyname=""         KMS keyring Keyname. Configurable with environment variable: KMS_KEYNAME

Args:
  <path>  Name of the saved encryption key
```
