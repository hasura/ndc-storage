# Upload / Download Objects

## Upload Objects

You can upload object files directly by encoding the file content to base64-encoded string or generate a presigned URL and let the client uploads the file to that URL. Presigned URLs are recommended especially with large files because GraphQL doesn't support file streaming. You have to encode the entire file to string that comsumes a lot of memory.

### Generate presigned URL (recommended)

Input the object path and the optional expiry of the presigned URL (if the `defaultPresignedExpiry` setting is configured).

```gql
query PresignedUploadUrl {
  storagePresignedUploadUrl(object: "hello.txt", expiry: "1h") {
    url
    expiredAt
  }
}
```

### Direct Upload

The object data must be encoded as base64 string.

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

Use the `uploadStorageObjectText` mutation if you are confident that object content is plain text. The request size is least than base64-encoded string 30%.

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

Similar to upload. You can download object files directly by encoding the file content to base64-encoded string or generate a presigned URL. Presigned URLs are also recommended to avoid memory leaks.

### Generate presigned URL (recommended)

Input the object path and the optional expiry of the presigned URL (if the `defaultPresignedExpiry` setting is configured).

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

### Upload Text Objects

Use the `uploadStorageObjectText` mutation if you are confident that object content is plain text.

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
