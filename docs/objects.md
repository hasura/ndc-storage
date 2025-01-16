# Manage Objects

## Upload Objects

You can upload object files directly by encoding the file content to base64-encoded string or generate a pre-signed URL and let the client upload the file to that URL. Presigned URLs are recommended especially with large files because GraphQL doesn't support file streaming. You have to encode the entire file to a string that consumes a lot of memory.

### Generate pre-signed URL (recommended)

Input the object path and the optional expiry of the pre-signed URL (if the `defaultPresignedExpiry` setting is configured).

```gql
query PresignedUploadUrl {
  storagePresignedUploadUrl(object: "hello.txt", expiry: "1h") {
    url
    expiredAt
  }
}
```

### Direct Upload

The object data must be encoded as a base64 string.

```gql
mutation UploadObject {
  uploadStorageObject(object: "hello.txt", data: "SGVsbG8gd29ybGQK") {
    bucket
    name
    size
    etag
  }
}
```

### Upload Text Objects

Use the `uploadStorageObjectText` mutation if you are confident that the object content is plain text. The request size is less than the base64-encoded string by 30%.

```gql
mutation UploadObjectText {
  uploadStorageObjectText(object: "hello2.txt", data: "Hello World") {
    bucket
    name
    size
    etag
  }
}
```

## Download Objects

Similar to upload. You can download object files directly by encoding the file content to base64-encoded string or generating a pre-signed URL. Presigned URLs are also recommended to avoid memory leaks.

### Generate pre-signed URL (recommended)

Input the object path and the optional expiry of the pre-signed URL (if the `defaultPresignedExpiry` setting is configured).

```gql
query GetSignedDownloadURL {
  storagePresignedDownloadUrl(object: "hello.txt", expiry: "1h") {
    url
    expiredAt
  }
}
```

### Direct Download

The response is a base64-encode string. The client must decode the string to get the raw content.

```gql
query DownloadObject {
  downloadStorageObject(object: "hello.txt")
}

# {
#   "data": {
#     "downloadStorageObject": "SGVsbG8gd29ybGQK"
#   }
# }
```

### Download Text Objects

Use the `downloadStorageObjectText` query if you are confident that the object content is plain text.

```gql
query DownloadObjectText {
  downloadStorageObjectText(object: "hello.txt")
}

# {
#   "data": {
#     "downloadStorageObjectText": "Hello world\n"
#   }
# }
```

### List Objects

> [!NOTE]
> The pagination information is optional. It depends on whether the storage provider's API supports this feature. The pagination method is cursor-based.

| Service              | Pagination |
| -------------------- | ---------- |
| AWS S3               | ❌        |
| MinIO                | ❌        |
| Google Cloud Storage | ❌        |
| Cloudflare R2        | ❌        |
| DigitalOcean Spaces  | ❌        |
| Azure Blob Storage   | ✅         |

```graphql
query ListObjects {
  storageObjects(where: { object: { _starts_with: "hello" } }) {
    pageInfo {
      cursor
      hasNextPage
      nextCursor
    }
    objects {
      clientId
      bucket
      name
      blobType
      serverEncrypted
      size
      storageClass
      userMetadata
      userTagCount
      userTags
      cacheControl
      checksumCrc32
      checksumCrc64Nvme
      checksumCrc32C
      checksumSha256
      checksumSha1
      contentEncoding
      contentDisposition
      contentLanguage
      contentMd5
      contentType
      etag
      expires
      isLatest
      metadata
      owner {
        id
        name
      }
    }
  }
}
```

## Multiple clients and buckets

You can upload to other buckets or services by specifying `clientId` and `bucket` arguments.

```gql
mutation UploadObject {
  uploadStorageObject(
    clientId: "gs"
    bucket: "other-bucket"
    object: "hello.txt"
    data: "SGVsbG8gd29ybGQK"
  ) {
    bucket
    name
    size
    etag
    checksumSha256
    lastModified
  }
}
```
