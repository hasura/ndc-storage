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

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.
> Note that the file content is encoded to base64 string so the response is 33% increased. If the maximum download size is 30 MB the actual allowed size is 20 MB only.

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

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.

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

#### Filter Arguments

You can use either `clientId`, `bucket`, `prefix`, or `where` boolean expression to filter object results. The `where` argument is mainly used for permissions. The filter expression is evaluated twice, before and after fetching the results. Cloud storage APIs usually support filtering by the name prefix only. Other operators (`_contains`, `_icontains`) are filtered from fetched results by pure logic.

```graphql
query ListObjects {
  storageObjects(prefix: "hello", where: { object: { _contains: "world" } }) {
    objects {
      name
      # ...
    }
  }
}
```

#### Pagination

Relay style suits object listing because most cloud storage services only support cursor-based pagination. The object name is used as the cursor ID.

```graphql
query ListObjects {
  storageObjects(after: "hello.txt", first: 3) {
    pageInfo {
      hasNextPage
    }
    edges {
      cursor
      node {
        clientId
        bucket
        name
        blobType
        serverEncrypted
        size
        storageClass
        tagCount
        tags
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
