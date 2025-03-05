package common

import (
	"bytes"
	"crypto/md5"
	"hash"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/hasura/ndc-sdk-go/utils"
	md5simd "github.com/minio/md5-simd"
)

var md5Pool = sync.Pool{New: func() any { return md5.New() }}

func newMd5Hasher() md5simd.Hasher {
	hash, _ := md5Pool.Get().(hash.Hash)

	return &hashWrapper{
		Hash: hash,
	}
}

// hashWrapper implements the md5simd.Hasher interface.
type hashWrapper struct {
	hash.Hash
}

// Close will put the hasher back into the pool.
func (m *hashWrapper) Close() {
	if m.Hash != nil {
		m.Reset()
		md5Pool.Put(m.Hash)
	}

	m.Hash = nil
}

// CalculateContentMd5 caculates file content using MD5 algorithm.
func CalculateContentMd5(reader io.Reader) (io.Reader, []byte, error) {
	readSeeker, ok := reader.(io.ReadSeeker)
	hash := newMd5Hasher()

	if ok {
		if _, err := io.Copy(hash, reader); err != nil {
			return nil, nil, err
		}
		// Seek back to beginning of io.NewSectionReader's offset.
		_, err := readSeeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, nil, err
		}

		reader = readSeeker
	} else {
		// Create a buffer.
		rawBytes, err := io.ReadAll(reader)
		if err != nil {
			return nil, nil, err
		}

		hash.Write(rawBytes)
		reader = bytes.NewReader(rawBytes)
	}

	result := hash.Sum(nil)
	hash.Close()

	return reader, result, nil
}

// KeyValuesToStringMap converts a storage key value slice to a string map.
func KeyValuesToStringMap(inputs []StorageKeyValue) map[string]string {
	result := make(map[string]string)

	for _, item := range inputs {
		result[item.Key] = item.Value
	}

	return result
}

// StringMapToKeyValues converts a string map to a key value slice.
func StringMapToKeyValues(inputs map[string]string) []StorageKeyValue {
	if len(inputs) == 0 {
		return []StorageKeyValue{}
	}

	keys := utils.GetSortedKeys(inputs)
	result := make([]StorageKeyValue, len(keys))

	for i, key := range keys {
		value := inputs[key]
		result[i] = StorageKeyValue{
			Key:   key,
			Value: value,
		}
	}

	return result
}

// KeyValuesToHeader converts a storage key value slice to a http.Header instance.
func KeyValuesToHeaders(inputs []StorageKeyValue) http.Header {
	result := http.Header{}

	for _, item := range inputs {
		result.Add(item.Key, item.Value)
	}

	return result
}

// ContentTypeFromFilePath tries to guess the content type from the extension of file path.
func ContentTypeFromFilePath(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return ""
	}

	return mime.TypeByExtension(ext)
}
