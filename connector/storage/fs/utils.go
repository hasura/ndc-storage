package fs

import "github.com/hasura/ndc-sdk-go/schema"

var errNotSupported = schema.NotSupportedError("FileStore doesn't support this method", nil)
