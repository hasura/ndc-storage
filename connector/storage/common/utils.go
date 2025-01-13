package common

import (
	"bytes"
	"crypto/md5"
	"hash"
	"io"
	"sync"

	md5simd "github.com/minio/md5-simd"
)

var md5Pool = sync.Pool{New: func() interface{} { return md5.New() }}

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
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			return nil, nil, err
		}

		hash.Write(rawBytes)
		reader = bytes.NewReader(rawBytes)
	}

	result := hash.Sum(nil)
	hash.Close()

	return reader, result, nil
}
