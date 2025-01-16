# Storage Connector

Storage Connector allows you to connect cloud storage services giving you an instant GraphQL API on top of your storage data.

This connector is built using the [Go Data Connector SDK](https://github.com/hasura/ndc-sdk-go) and implements the [Data Connector Spec](https://github.com/hasura/ndc-spec).

## Features

### Supported storage services

At this moment, the connector supports S3 Compatible Storage services.

| Service              | Supported |
| -------------------- | --------- |
| AWS S3               | ✅        |
| MinIO                | ✅        |
| Google Cloud Storage | ✅        |
| Cloudflare R2        | ✅        |
| DigitalOcean Spaces  | ✅        |
| Azure Blob Storage   | ✅        |

## Get Started

Follow the [Quick Start Guide](https://hasura.io/docs/3.0/getting-started/overview/) in Hasura DDN docs. At the `Connect to data` step, choose the `hasura/storage` data connector from the dropdown and follow the interactive prompts to set required environment variables.

The connector is built upon the MinIO Go Client SDK so it supports most of methods in the [API interface](https://min.io/docs/minio/linux/developers/go/API.html)

## Documentation

- [Configuration](./docs/configuration.md)
- [Manage Objects](./docs/objects.md)

## License

Storage Connector is available under the [Apache License 2.0](./LICENSE).
