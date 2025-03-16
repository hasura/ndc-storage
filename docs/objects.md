# Manage Objects

## Upload Objects

You can upload object files directly by encoding the file content to base64-encoded string or generate a pre-signed URL and let the client upload the file to that URL. Presigned URLs are recommended especially with large files because GraphQL doesn't support file streaming. You have to encode the entire file to a string that consumes a lot of memory.

### Generate pre-signed URL (recommended)

Input the object path and the optional expiry of the pre-signed URL (if the `defaultPresignedExpiry` setting is configured).

```gql
query PresignedUploadUrl {
  storagePresignedUploadUrl(name: "hello.txt", expiry: "1h") {
    url
    expiredAt
  }
}
```

> [!NOTE]
> If the endpoint host is a private DNS the presigned-URL cannot be accessed. In this case, you must configure the `publicHost` in the client settings.

### Direct Upload

> [!NOTE]
> The connector limits the maximum upload size via the `runtime.maxUploadSizeMBs` setting to avoid memory leaks.
> Note that the file content is encoded to base64 string so the request body is 33% increased. For example, if the file size is 20 MB the request size will be 30 MB.

The object data must be encoded as a base64 string.

```gql
mutation UploadObject {
  uploadStorageObjectAsBase64(name: "hello.txt", data: "SGVsbG8gd29ybGQK") {
    bucket
    name
    size
    etag
  }
}
```

### Upload Text Objects

> [!NOTE]
> The connector limits the maximum upload size via the `runtime.maxUploadSizeMBs` setting to avoid memory leaks.

Use the `uploadStorageObjectAsText` mutation if you are confident that the object content is plain text. The request size is less than the base64-encoded string by 30%.

```gql
mutation UploadObjectText {
  uploadStorageObjectAsText(name: "hello2.txt", data: "Hello World") {
    bucket
    name
    size
    etag
  }
}
```

### Upload From a URL

> [!NOTE]
> The connector limits the maximum upload size via the `runtime.maxUploadSizeMBs` setting to avoid memory leaks.

The connector will download the file from the `url` argument via HTTP protocol and upload it to the storage service.

```gql
mutation UploadObjectFromURL {
  uploadStorageObjectFromUrl(
    name: "hello2.txt"
    url: "https://example.local/hello.txt"
  ) {
    name
    size
  }
}
```

## Download Objects

Similar to upload. You can download object files directly by encoding the file content to base64-encoded string or generating a pre-signed URL. Presigned URLs are also recommended to avoid memory leaks.

### Generate pre-signed URL (recommended)

Input the object path and the optional expiry of the pre-signed URL (if the `defaultPresignedExpiry` setting is configured).

```gql
query GetSignedDownloadURL {
  storagePresignedDownloadUrl(name: "hello.txt", expiry: "1h") {
    url
    expiredAt
  }
}
```

> [!NOTE]
> If the endpoint host is a private DNS the presigned-URL cannot be accessed. In this case, you must configure the `publicHost` in the client settings.

### Direct Download

The response is a base64-encode string. The client must decode the string to get the raw content.

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.
> Note that the file content is encoded to base64 string so the response is 33% increased. If the maximum download size is 30 MB the actual allowed size is 20 MB only.

```gql
query DownloadObject {
  downloadStorageObjectAsBase64(name: "hello.txt") {
    data
  }
}

# {
#   "data": {
#     "downloadStorageObjectAsBase64": {
#       "data": "SGVsbG8gd29ybGQK"
#     }
#   }
# }
```

### Download Objects as Text

Use the `downloadStorageObjectAsText` query if you are confident that the object content is plain text.

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.

```gql
query DownloadObjectAsText {
  downloadStorageObjectAsText(name: "hello.txt") {
    data
  }
}

# {
#   "data": {
#     "downloadStorageObjectAsText": {
#       "data": "Hello world\n"
#     }
#   }
# }
```

### Download and decode JSON files

Use the `downloadStorageObjectAsJson` query to download and decode text files to arbitrary JSON.

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.

```gql
query DownloadObjectAsJson {
  downloadStorageObjectAsJson(name: "hello.json") {
    data
  }
}

# {
#   "data": {
#     "downloadStorageObjectAsJson": {
#       "data": {
#         "hello": "world"
#       }
#     }
#   }
# }
```

### Download and decode CSV files

Use the `downloadStorageObjectAsCsv` query to download and decode CSV files. The file content must be in text.

> [!NOTE]
> The connector limits the maximum download size via the `runtime.maxDownloadSizeMBs` setting to avoid memory leaks. The GraphQL engine on Hasura Cloud also limits the max response size from connectors. The acceptable file size should be 30 MB in maximum.

```gql
query DownloadObjectAsCsv {
  downloadStorageObjectAsCsv(name: "hello.csv", options: { parse_json: true }) {
    data
  }
}

# {
#   "data": {
#     "downloadStorageObjectAsCsv": {
#       "data": [
#         {
#           "id": 1,
#           "name": "Pike",
#           "active": true
#         },
#         {
#           "id": 2,
#           "name": "Thompson",
#           "active": true
#         }
#       ]
#     }
#   }
# }
```

#### Options

- `no_header`: the connect uses values of the first row as object properties. If this option is true the connector will assume that first row is not the header. The result will be a 2-dimension matrix.
- `transpose`: transpose the matrix before converting.
- `parse_json`: if this option is set the connector will try parsing cell values using JSON encoding.
- `delimiter`: specify the delimiter character to decode CSV. The default delimiter is comma `,`.
- `comment`: the comment character. If not empty lines beginning with the comment character without preceding whitespace are ignored.
- `lazy_quotes`: the default is true. A quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.
- `trim_leading_space`: the default is true. Leading white space in a field is ignored.

### List Objects

#### Filter Arguments

You can use either `clientId`, `bucket`, `prefix`, or `where` boolean expression to filter object results. The `where` argument is mainly used for permissions. The filter expression is evaluated twice, before and after fetching the results. Cloud storage APIs usually support filtering by the name prefix only. Other operators (`_contains`, `_icontains`) are filtered from fetched results by pure logic.

> [!INFO]
> If you want to filter objects recursively, set the argument `recursive: true`.

```graphql
query RelayListObjects {
  storageObjectConnections(
    prefix: "hello"
    recursive: true
    where: { name: { _contains: "world" } }
  ) {
    edges {
      cursor
      node {
        clientId
        bucket
        # ...
      }
    }
  }
}
```

In `storageObjects` query, the `prefix` argument doesn't exist. You should use the `_starts_with` operator in the `where` predicate instead.

```graphql
query ListObjects {
  storageObjects(
    args: { recursive: true }
    where: { name: { _starts_with: "hello", _contains: "world" } }
  ) {
    clientId
    bucket
    name
    # ...
  }
}
```

#### Pagination

Relay style suits object listing because most cloud storage services only support cursor-based pagination. The object name is used as the cursor ID.

```graphql
query RelayListObjects {
  storageObjectConnections(recursive: true, after: "hello.txt", first: 3) {
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

The `storageObjects` collection doesn't return pagination information. You need to get the `name` in the last object to paginate by the `after` cursor.

> [!NOTE]
>
> **Why do `storageObjects` and `storageObjectConnections` operations exist?**
> The `storageObjects` collection provides a simpler response structure that PromptQL can query easily. The `storageObjectConnections` function returns a better cursor-based pagination response on but the schema is complicated for PromptQL to understand.

> [!NOTE]
> Sorting isn't supported.

```graphql
query ListObjects {
  storageObjects(after: "hello.txt", limit: 3) {
    clientId
    bucket
    name
    # ...
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
    name: "hello.txt"
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
