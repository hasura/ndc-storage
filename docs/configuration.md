# Configuration

## Clients

### General Settings

The configuration file `configuration.yaml` contains a list of storage clients. Every client has common settings:

- `id`: the unque identity name of the client. This setting is optional unless there are many configured clients.
- `type`: type of the storage provider. Accept one of `s3`, `gs`.
- `defaultBucket`: the default bucket name.
- `authenticaiton`: the authentication setting.
- `endpoint`: the base endpoint of the storage server. Required for other S3 compatible services such as MinIO, Cloudflare R2, DigitalOcean Spaces, etc...
- `publicHost`: is used to configure the public host for presigned URL generation if the connector communicates privately with the storage server through an internal DNS. If this setting isn't set the host of the generated URL will be a private DNS that isn't accessible from the internet.
- `region`: (optional) region of the bucket going to be created.
- `maxRetries`: maximum number of retry times. Default 10.
- `defaultPresignedExpiry`: the default expiry for presigned URL generation in duration format. The maximum expiry is 7 days \(`168h`\) and minimum is 1 second \(`1s`\).
- `trailingHeaders`: indicates server support of trailing headers. Only supported for v4 signatures.
- `allowedBuckets`: the list of allowed bucket names. This setting prevents users to get buckets and objects outside the list. However, it's recommended to restrict the permissions for the IAM credentials. This setting is useful to let the connector know which buckets belong to this client. The empty value means all buckets are allowed. The storage server will handle the validation.

### Authentication

#### Static Credentials

Configure the authentication type `static` with `accessKeyId` and `secretAccessKey`. `sessionToken` is also supported for [temporary access](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_use-resources.html) but for testing only.

```yaml
clients:
  - type: s3
    authentication:
      type: static
      accessKeyId:
        env: ACCESS_KEY_ID
      secretAccessKey:
        env: SECRET_ACCESS_KEY
```

#### IAM

The IAM authentication retrieves credentials from the AWS EC2, ECS or EKS service, and keeps track if those credentials are expired. This authentication method can be used only if the connector is hosted in the AWS ecosystem.

The following settings are supported:

- `iamAuthEndpoint` : the optional custom endpoint to fetch IAM role credentials. The client can automatically identify the endpoint if not set.

### Examples

#### AWS S3

Create [a user access key](https://docs.aws.amazon.com/IAM/latest/UserGuide/access-keys-admin-managed.html) with S3 permissions to configure the Access Key ID and Secret Access Key.

```yaml
clients:
  - id: s3
    type: s3
    defaultBucket:
      env: DEFAULT_BUCKET
    authentication:
      type: static
      accessKeyId:
        env: ACCESS_KEY_ID
      secretAccessKey:
        env: SECRET_ACCESS_KEY
```

#### Google Cloud Storage

You need to [generate HMAC key](https://cloud.google.com/storage/docs/authentication/hmackeys) to configure the Access Key ID and Secret Access Key.

```yaml
clients:
  - id: gs
    type: gs
    defaultBucket:
      env: DEFAULT_BUCKET
    authentication:
      type: static
      accessKeyId:
        env: ACCESS_KEY_ID
      secretAccessKey:
        env: SECRET_ACCESS_KEY
```

#### Other S3 compatible services

You must configure the endpoint URL alongs with Access Key ID and Secret Access Key.

```yaml
clients:
  - id: minio
    type: s3
    endpoint:
      env: STORAGE_ENDPOINT
    defaultBucket:
      env: DEFAULT_BUCKET
    authentication:
      type: static
      accessKeyId:
        env: ACCESS_KEY_ID
      secretAccessKey:
        env: SECRET_ACCESS_KEY
```

#### Cloudflare R2

You must configure the endpoint URL alongs with [Access Key ID and Secret Access Key](https://developers.cloudflare.com/r2/api/s3/tokens/#get-s3-api-credentials-from-an-api-token). See [Cloudflare docs](https://developers.cloudflare.com/r2/api/s3/api/) for more context.

#### DigitalOcean Spaces

See [Spaces API Reference Documentation](https://docs.digitalocean.com/reference/api/spaces-api/).