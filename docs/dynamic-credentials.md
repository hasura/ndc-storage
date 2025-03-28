# Dynamic Credentials Mode

## How it works

You can allow users to input storage client credentials directly while sending GraphQL requests by enabling `generator.dynamicCredentials`.

```yaml
# configuration.yaml
runtime:
  maxDownloadSizeMBs: 2
generator:
  promptqlCompatible: true
  dynamicCredentials: true
clients: []
```

The following credential arguments will be presented in all operations:

- `clientType`: The storage provider type. Accept `s3`, `gcs`, and `azblob`. If not set the default provide type will be `s3`.
- `accessKeyId`: Access key ID of S3, GCS or account name of Azure Blob Storage.
- `secretAccessKey`: Secret access key of S3, GCS or the account key of Azure Blob Storage.
- `endpoint`: Endpoint of the storage service. Required for other S3-compatible services such as MinIO, R2, etc... and Azure Blob Storage.

## GraphQL Examples

### S3-compatible Storage

Add the `s3` client type, `accessKeyId` and `secretAccessKey` to request arguments. The endpoint is required if the storage service isn't S3.

```graphql
query DownloadStorageObjectAsText {
  downloadStorageObjectAsText(
    clientType: "s3"
    endpoint: "http://minio:9000"
    accessKeyId: "test-key"
    secretAccessKey: "randomsecret"
    name: "people-1000.csv"
    bucket: "default"
  ) {
    data
  }
}
```

### Google Cloud Storage

Add the `gcs` client type, `accessKeyId` and `secretAccessKey` to request arguments. The Access Key ID and Secret Access Key are [generated HMAC key](https://cloud.google.com/storage/docs/authentication/hmackeys).

```graphql
query DownloadStorageObjectAsText {
  downloadStorageObjectAsText(
    clientType: "gcs"
    accessKeyId: "test-key"
    secretAccessKey: "randomsecret"
    name: "people-1000.csv"
    bucket: "default"
  ) {
    data
  }
}
```

### Azure Blob Storage

Support shared key and connection string credentials.

#### Connection String

Add the `azblob` client type and `endpoint` to request arguments.

```graphql
query DownloadStorageObjectAsText {
  downloadStorageObjectAsText(
    clientType: "azblob"
    endpoint: "AccountName=local;AccountKey=xxx;BlobEndpoint=default"
    name: "people-1000.csv"
    bucket: "default"
  ) {
    data
  }
}
```

#### Shared Key

Add the `azblob` client type, `endpoint`, `accessKeyId` and `secretAccessKey` to request arguments. Access Key ID and Secret Access Key are account name and account key.

```graphql
query DownloadStorageObjectAsText {
  downloadStorageObjectAsText(
    clientType: "azblob"
    endpoint: "https://local.hasura.dev:10000"
    accessKeyId: "local"
    secretAccessKey: "xxxx"
    name: "people-1000.csv"
    bucket: "default"
  ) {
    data
  }
}
```

## PromptQL Examples

See [Dynamic Credentials example in PromptQL](./promptql.md).
