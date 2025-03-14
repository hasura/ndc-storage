# Storage Connector

Storage Connector allows you to connect cloud storage services giving you an instant GraphQL API on top of your storage data.

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

## Features

### Supported storage services

| Service                  | Type     | List Buckets | Create Bucket | Update Bucket | Delete Bucket | List Objects | Upload | Download | Delete Object | Soft-Delete | Presigned-URL |
| ------------------------ | -------- | ------------ | ------------- | ------------- | ------------- | ------------ | ------ | -------- | ------------- | ----------- | ------------- |
| AWS S3 (\*)              | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| Google Cloud Storage     | `gcs`    | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ✅          | ✅            |
| Azure Blob Storage       | `azblob` | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ✅          | ✅            |
| File System              | `fs`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ❌            |
| MinIO (\*)               | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| Cloudflare R2 (\*)       | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |
| DigitalOcean Spaces (\*) | `s3`     | ✅           | ✅            | ✅            | ✅            | ✅           | ✅     | ✅       | ✅            | ❌          | ✅            |

(\*): Support Amazon S3 Compatible Cloud Storage providers. The connector uses [MinIO Go Client SDK](https://github.com/minio/minio-go) behind the scenes.

## Get Started

Follow the [Quick Start Guide](https://hasura.io/docs/3.0/getting-started/overview/) in Hasura DDN docs. At the `Connect to data` step, choose the `hasura/storage` data connector from the dropdown and follow the interactive prompts to set the required environment variables.

AWS S3 environment variables are the default settings in the interactive prompt. If you want to use other storage providers you need to manually configure the `configuration.yaml` file and add the required environment variable mappings to the subgraph definition.

## Documentation

- [Configuration](./docs/configuration.md)
- [Manage Objects](./docs/objects.md)

## License

Storage Connector is available under the [Apache License 2.0](./LICENSE).
