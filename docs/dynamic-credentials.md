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

**Example**:

```graphql
query DownloadStorageObjectText {
  downloadStorageObjectText(
    clientType: "s3"
    endpoint: "http://minio:9000"
    accessKeyId: "test-key"
    secretAccessKey: "randomsecret"
    object: "people-1000.csv"
    bucket: "default"
  ) {
    data
  }
}
```
