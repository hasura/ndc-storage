package fs

import (
	"os"

	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
)

var errNotSupported = schema.NotSupportedError("FileStore doesn't support this method", nil)

func serializeStorageObject(filePath string, info os.FileInfo) common.StorageObject {
	result := common.StorageObject{
		Name:         filePath,
		IsDirectory:  info.IsDir(),
		LastModified: info.ModTime(),
	}

	if !result.IsDirectory {
		size := info.Size()
		result.Size = &size
	}

	return result
}
