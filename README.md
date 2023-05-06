# Vault kubernetes dotenv manager

A tool for populating secrets from Hashicorp Vault in .env format and keeping them alive for Kubernetes pods.

## Usage

The manager should be run as either an init container or a sidecar in a Kubernetes pod. The manager will authenticate to
Vault using the Kubernetes authentication method with the service account token. For it to work, the service account
token must be mounted into the filesystem. The manager is configured via a YAML configuration file.

### Probes

The manager exposes an HTTP port usable for Kubernetes probes at the `/liveness` path. The port will only be opened
while the manager is in the keep-alive state, so only after the secrets have been populated. This way it is usable with
for example [kubexit](https://github.com/karlkfi/kubexit) to manage the main container's lifecycle from the manager
sidecar.

### Operating modes

The manager lifecycle has 2 stages:

* Populate: in this phase the manager will authenticate to Vault, retrieve the secrets and write them to the volumes.
  The probe HTTP endpoint will not return success while in this phase.
* Keep-alive: in this phase the manager will expect the population to be complete, and it will just keep any leases it
  acquired alive.

It allows setting one of 3 operating modes using these phases using the optional `-mode` flag:

* `populate`: In this mode only the populate phase is executed, after which the manager will exit. In this mode the
  `revokeAuthLeaseOnQuit` configuration option is ignored, and the leases will not be revoked when the manager exits.
  This mode can be used in an init container to ensure the secrets are populated before the containers start.
* `keep-alive`: In this mode only the keep-alive phase is executed and the manager will keep running until terminated
  (or fails to renew the leases). The `revokeAuthLeaseOnQuit` configuration option is respected, so if it is set to
  true, the manager will revoke all leases before exiting. This mode will not populate the secrets, if the secrets and
  the data directory are not populated, the manager will exit with an error. This mode is meant to be used as a
  sidecar container if there are no sidecar lifecycle management tools are used and the `populate` mode has already
  been run as an init container.
* default mode: If no `-mode` flag has been set, then both the populate and keep-alive phases are executed and the
  manager will keep running until terminated. The `revokeAuthLeaseOnQuit` configuration is respected in this mode.
  If you use a container lifecycle management tool like kubexit or if your main container is not sensitive to the
  secrets not being fully populated at startup, then this is the recommended mode.

### Command line flags

| name                  | description                                                                                                         | default       |
|-----------------------|---------------------------------------------------------------------------------------------------------------------|---------------|
| config                | The path to the configuration file                                                                                  | `config.yaml` |
| mode                  | The operating mode as described in the [operating modes](#Operating modes) section                                  | default mode  |
| http-port             | The port to listen on for the HTTP probe endpoint                                                                   | 8000          |
| wait-after-population | The number of seconds to wait after the population phase before either exiting or moving on to the keep-alive phase | 0             |
| logtostderr           | Whether to send the logs to stderr or to stdout                                                                     | true          |
| stderrthreshold       | The log level threshold for the messages to send to stderr                                                          | Info          |
| help                  | Shows the usage                                                                                                     |               |

### Configuration file

The configuration file is in YAML format, by default the manager will look for a file called `config.yaml` in the
working directory unless overridden with the `-config` flag. The JSON schema for the configuration file is available in
the `config.schema.yaml` file. An example file is in the repository called `config.example.yaml`.

#### Base configuration

| name                | type                                   | required | description                                                                                                                                                                            |
|---------------------|----------------------------------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| dataDir             | string                                 | **yes**  | The directory to store the authentication token and secret lease data in. This directory will store the authentication token, so care should be taken that nothing else can access it. |
| vaultUrl            | string                                 | **yes**  | The URL to the Vault instance.                                                                                                                                                         |
| tokenPath           | string                                 | no       | The path to the Kubernetes service account token. Defaults to `/var/run/secrets/kubernetes.io/serviceaccount/token`, which should be fine for most setups.                             |
| namespace           | string                                 | no       | The Kubernetes namespace to use during authentication. Defaults to `default`                                                                                                           |
| role                | string                                 | **yes**  | The Vault role to use during authentication.                                                                                                                                           |
| vaultAuthMethodPath | string                                 | **yes**  | The path to the authentication method to use in Vault                                                                                                                                  |
| secrets             | array of [secret](#Secret definitions) | **yes**  | The definitions of the secrets                                                                                                                                                         |

#### Secret definitions

| name          | type                    | required                               | description                                                                                                                                                                                                                                                                                                                          |
|---------------|-------------------------|----------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| name          | string                  | **yes**                                | Human readable name for the secret. Will be used in error messages                                                                                                                                                                                                                                                                   |
| origin        | enum (file,token,vault) | **yes**                                | The source for this secret. "token" for the vault authentication token, "file" if a file in the filesystem, or "vault" if the source is a vault secret. NOTE that the token and dynamic vault secrets will expire if the manager is not keeping them alive                                                                           |
| format        | enum (dotenv, file)     | **yes**                                | The output format for the secret. Can be either "dotenv" to put the values into a file in .env format or "file" to place the secret values into individual files, where the file name will be the key of the secret.                                                                                                                 |
| directoryMode | int                     | no                                     | The filesystem mode (unix permissions) of the directory to place the secrets in. Only applies if the directory will be created by the manager. Must be in octal notation (0755, 0644, etc). Note the umask will be still applied on top of this permission. Defaults to 0755                                                         |
| fileMode      | int                     | no                                     | The filesystem mode (unix permissions) of any created files. Only applies to files created while populating this secret. Must be in octal notation (0755, 0644, etc). Note the umask will be still applied on top of this permission. Defaults to 0644                                                                               |
| source        | string                  | **yes** for `file` and `vault` origins | Source path for the secret. For vault source secrets this is the path for the secret in vault, for file source secrets it's the path to the source file. Token source secrets don't use it. Required for file and vault secrets.                                                                                                     |
| destination   | string                  | **yes*                                 | The path to where to populate the secret. For dotenv format secrets, it's the path to the .env file, for file format secrets it's the path to the directory where to crate the files.                                                                                                                                                |
| secretBaseKey | string                  | no                                     | If the secret source stores the secret in a sub object, then the key for the sub object is set here. Typically used with a vault kv type secret, which responds with an object, where the actual secret data is stored under a base key called "data". See [secretBaseKey](#secretBaseKey) for details.                              |
| mapping       | object                  | no                                     | If the secret keys need to be mapped to something else in the target, this object should store the mappings in an object with the key being the destination/mapped key, and the value the source key in the secret. If mapping is used, only the mapped keys from the secret will be populated. See [mapping](#mapping) for details. |
| decoders      | array of enum (base64)  | no                                     | If the secret values are encoded, and need to be decoded before population, the decoders can be set here. Multiple decoders are supported, and the decoders will be used in the order they are listed here. By default no decoders are used.                                                                                         |

## Example

Given a version2 key value secret in Vault at the path `/kv2/api-key` with the following contents:
```json
{
  "API_KEY": "test"
}
```

Given a version2 key value secret in Vault at the path `/kv2/tls` with the following contents. Both the certificate and the key are in DER format base64 encoded before stored in Vault:
```json
{
  "certificate": "...",
  "key": "..."
}
```

Given a database engine set at `/database` and a database role in the engine called `prod`

Given a base dotenv file that does not contain any secrets (stored in a configmap for example):
```dotenv
APP_ENV=production
DB_HOST=db
```

With the following configuration file:
```yaml
dataDir: /data-dir # A volumeMount for an in memory emptyDir volume for example
vaultUrl: https://vault.example.com
role: kubernetes # The name of the role used during authentication
vaultAuthMethodPath: kubernetes # The path to the kubernetes authentication method set up in Vault

secrets:
# First secret that copies in the base .env file. In this example it is stored in a configmap that is mounted to /dotenv-base
- name: dotenv                  # Just a descriptive human-readable name used in errrors to help identify which secret caused the problem
  origin: file                  # The origin is a file on the filesystem
  format: file                  # We just treat it as a file and effectively copy it
  source: /dotenv-base/.env     # The path where the configmap is mounted as a volumeMount, including the file name
  destination: /dotenv          # The path where to create the secret. Since we are in `file` format, the filename is not set here
- name: api-key                 
  format: dotenv                
  source: /kv2/data/api-key     # The KV engine exposes the data under the `data` subpath
  destination: /dotenv/.env     # We need to give the full path including the filename here, since we are not in file mode
  secretBaseKey: data           # Required for kv version 2 secrets
- name: db                      
  format: dotenv                
  source: /database/creds/prod  # The standard path for retrieving credentials from Vault
  destination: /dotenv/.env     
  mapping:                       
    DB_RO_USER: username        # The application needs both an RO and an RW user, but we happen to use the same for both here
    DB_RO_PASSWORD: password     
    DB_RW_USER: username        
    DB_RW_PASSWORD: password    
- name: tls                     
  format: file                  # Place each value as a separate file in the destination dir
  source: /kv2/tls              
  destination: /tls             # The destination directory
  secretBaseKey: data           # Required for kv version 2 secrets
  decoders:                      
  - base64                      # The base64 decoder will decode the secrets before saving them as files
  mapping:                      
    tls.crt: certificate        # We want to rename the files to tls.crt and tls.key
    tls.key: key
```

Then a `.env` file in the `/dotenv` directory should be created with a content similar to below (order of values in the secret values is undefined)
```dotenv
APP_ENV=production
DB_HOST=db

##########################
# Secret source: api-key #
##########################
API_KEY=test

#####################
# Secret source: db #
#####################
DB_RO_USER=dynamic-test-user
DB_RO_PASSWORD=dynamic-random-password
DB_RW_USER=dynamic-test-user
DB_RW_PASSWORD=dynamic-random-password
```

And the `/tls` directory should have a file called `tls.crt` created with the certificate in DER format and a `tls.key` 
file with the private key in DER format.

### secretBaseKey

For any secret engine, that returns the secret values not on the top level (for example kv and kv version 2), the 
`secretBaseKey` configuration value can be used to specify the key in the top level object that contains the actual 
secrets (`data` for the kv engines).

### mapping

Mapping allows renaming secrets to different dotenv keys or different filenames before populating them. When using 
mapping, the new name must be set as the key of the object, and the original name must be set as the value.

Given the following mapping:
```yaml
mapping:
  DB_RO_USER: username
  DB_RO_PASSWORD: password
```

The mapping will add the key `DB_RO_USER` in the .env file with the value of the `username` key in the secret. The same 
original keys can be used multiple times in the mapping.

When using mapping, only mapped values will be populated from the given secret, all other secret values will be 
discarded.

### decoders

Decoders allow storing data in various encodings in secrets, and the manager will decode each value of a secret before 
populating. This allows for example storing binary data in Vault secrets in base 64 encoding.

If multiple decoders are specified, they will be used in the order they were defined in. This functionality can be 
useful if some secret values are encoded multiple times.

Currently, the only supported decoder is the `base64` decoder.

### Permission handling

Filesystem permissions for a secret can be set using the `directoryMode` and `fileMode` values. The permissions should 
be set in octal notation (starting with the `0` digit to designate the notation). Defaults for files are `0644` and for 
directories `0755`. The permissions are only applied when a file or directory is created. If they file or directory 
already exists, the permissions are not changed. The umask will be applied on top of these permissions, so if you set 
`0666` for a file, but the created file has `0644` permissions, verify the umask settings. 

## Gotchas

### Termination

When terminating a pod, Kubernetes will send the terminate signal to all containers at the same time (unless some delay
configuration has been used). The manager will terminate immediately when it receives a terminate signal. If the
`revokeAuthLeaseOnQuit` config option is used, this means the leases will be revoked immediately. This can result in
dynamic secrets, like database credentials or the vault token get invalidated in Vault before the main container
actually shuts down, potentially causing errors in the main application. If using this option, the usage of a container
lifecycle management tool is recommended until Kubernetes implements the sidecar feature:
https://github.com/kubernetes/enhancements/issues/753

