package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"time"

	"github.com/hasura/ndc-sdk-go/scalar"
	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/minio/minio-go/v7/pkg/s3utils"
)

// ListObjects lists objects in a bucket.
func (m *Manager) ListObjects(ctx context.Context, bucketInfo common.StorageBucketArguments, opts *common.ListStorageObjectsOptions, predicate func(string) bool) (*common.StorageObjectListResults, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return &common.StorageObjectListResults{ //nolint:nilerr
			Objects: []common.StorageObject{},
		}, nil
	}

	results, err := client.ListObjects(ctx, bucketName, opts, predicate)
	if err != nil {
		return nil, err
	}

	for i := range results.Objects {
		results.Objects[i].ClientID = string(client.id)
		results.Objects[i].Bucket = bucketName
	}

	return results, nil
}

// ListObjects lists deleted objects in a bucket.
func (m *Manager) ListDeletedObjects(ctx context.Context, bucketInfo common.StorageBucketArguments, opts *common.ListStorageObjectsOptions, predicate func(string) bool) (*common.StorageObjectListResults, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	results, err := client.ListDeletedObjects(ctx, bucketName, opts, predicate)
	if err != nil {
		return nil, err
	}

	for i := range results.Objects {
		results.Objects[i].ClientID = string(client.id)
		results.Objects[i].Bucket = bucketName
	}

	return results, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (m *Manager) ListIncompleteUploads(ctx context.Context, bucketInfo common.StorageBucketArguments, opts common.ListIncompleteUploadsOptions) ([]common.StorageObjectMultipartInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	return client.ListIncompleteUploads(ctx, bucketName, opts)
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (m *Manager) GetObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, io.ReadCloser, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, nil, err
	}

	objectStat, err := m.statObject(ctx, client, bucketName, objectName, opts)
	if err != nil || objectStat == nil {
		return nil, nil, err
	}

	if objectStat.IsDirectory {
		return nil, nil, schema.UnprocessableContentError("cannot download directory: "+objectName, nil)
	}

	if objectStat.Size == nil || *objectStat.Size > m.runtime.MaxDownloadSizeMBs*1024*1024 {
		return nil, nil, schema.UnprocessableContentError(fmt.Sprintf("file size > %d MB is not allowed to be downloaded directly. Please use presignedGetObject function for large files", m.runtime.MaxDownloadSizeMBs), nil)
	}

	content, err := client.GetObject(ctx, bucketName, objectName, opts)

	return objectStat, content, err
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (m *Manager) PutObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts *common.PutStorageObjectOptions, data []byte) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	contentLength := int64(len(data))
	if contentLength > m.runtime.MaxUploadSizeMBs*1024*1024 {
		return nil, maxUploadSizeLimitError(m.runtime.MaxUploadSizeMBs)
	}

	result, err := client.PutObject(ctx, bucketName, objectName, opts, bytes.NewReader(data), contentLength)
	if err != nil {
		return nil, err
	}

	result.Bucket = bucketName
	result.ClientID = string(client.id)

	return result, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (m *Manager) CopyObject(ctx context.Context, args *common.CopyStorageObjectArguments) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, common.StorageBucketArguments{
		StorageClientCredentialArguments: common.StorageClientCredentialArguments{
			ClientID: args.ClientID,
		},
		Bucket: args.Dest.Bucket,
	})
	if err != nil {
		return nil, err
	}

	args.Dest.Bucket = bucketName

	if args.Source.Bucket == "" {
		args.Source.Bucket = client.defaultBucket
	}

	result, err := client.CopyObject(ctx, args.Dest, args.Source)
	if err != nil {
		return nil, err
	}

	result.ClientID = string(client.id)

	return result, nil
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (m *Manager) ComposeObject(ctx context.Context, args *common.ComposeStorageObjectArguments) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, common.StorageBucketArguments{
		StorageClientCredentialArguments: common.StorageClientCredentialArguments{
			ClientID: args.ClientID,
		},
		Bucket: args.Dest.Bucket,
	})
	if err != nil {
		return nil, err
	}

	args.Dest.Bucket = bucketName
	srcs := make([]common.StorageCopySrcOptions, len(args.Sources))

	for i, src := range args.Sources {
		if src.Bucket == "" {
			src.Bucket = client.defaultBucket
		}

		srcs[i] = src
	}

	result, err := client.ComposeObject(ctx, args.Dest, srcs)
	if err != nil {
		return nil, err
	}

	result.ClientID = string(client.id)

	return result, nil
}

// StatObject fetches metadata of an object.
func (m *Manager) StatObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	return m.statObject(ctx, client, bucketName, objectName, opts)
}

func (m *Manager) statObject(ctx context.Context, client *Client, bucketName, objectName string, opts common.GetStorageObjectOptions) (*common.StorageObject, error) {
	result, err := client.StatObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	result.ClientID = string(client.id)
	result.Bucket = bucketName

	return result, nil
}

// RemoveObject removes an object with some specified options.
func (m *Manager) RemoveObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.RemoveStorageObjectOptions) error {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return err
	}

	return client.RemoveObject(ctx, bucketName, objectName, opts)
}

// UpdateObject updates object configuration.
func (m *Manager) UpdateObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.UpdateStorageObjectOptions) error {
	if opts.IsEmpty() {
		return nil
	}

	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return err
	}

	return client.UpdateObject(ctx, bucketName, objectName, opts)
}

// RemoveObjects remove a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (m *Manager) RemoveObjects(ctx context.Context, bucketInfo common.StorageBucketArguments, opts *common.RemoveStorageObjectsOptions, predicate func(string) bool) ([]common.RemoveStorageObjectError, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	return client.RemoveObjects(ctx, bucketName, opts, predicate), nil
}

// RestoreObject restores a soft-deleted object.
func (m *Manager) RestoreObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string) error {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return err
	}

	return client.RestoreObject(ctx, bucketName, objectName)
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (m *Manager) RemoveIncompleteUpload(ctx context.Context, args *common.RemoveIncompleteUploadArguments) error {
	client, bucketName, err := m.GetClientAndBucket(ctx, common.StorageBucketArguments{
		StorageClientCredentialArguments: common.StorageClientCredentialArguments{
			ClientID: args.ClientID,
		},
		Bucket: args.Bucket,
	})
	if err != nil {
		return err
	}

	return client.RemoveIncompleteUpload(ctx, bucketName, args.Name)
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (m *Manager) PresignedGetObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, opts common.PresignedGetStorageObjectOptions) (*common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	var exp time.Duration

	if opts.Expiry != nil {
		exp = opts.Expiry.Duration
	} else if client.defaultPresignedExpiry != nil {
		exp = *client.defaultPresignedExpiry
		opts.Expiry = &scalar.DurationString{Duration: exp}
	}

	if exp == 0 {
		return nil, schema.UnprocessableContentError("expiry is required and must be larger than 0", nil)
	}

	rawURL, err := client.PresignedGetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return nil, err
	}

	return &common.PresignedURLResponse{
		URL:       rawURL,
		ExpiredAt: time.Now().Add(exp),
	}, nil
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations.
// Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The default expiry is set to 7 days.
func (m *Manager) PresignedPutObject(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, expiry *scalar.DurationString) (*common.PresignedURLResponse, error) {
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	var exp time.Duration

	if expiry != nil {
		exp = expiry.Duration
	} else if client.defaultPresignedExpiry != nil {
		exp = *client.defaultPresignedExpiry
	}

	if exp == 0 {
		return nil, schema.UnprocessableContentError("expiry is required and must be larger than 0", nil)
	}

	rawURL, err := client.PresignedPutObject(ctx, bucketName, objectName, exp)
	if err != nil {
		return nil, err
	}

	return &common.PresignedURLResponse{
		URL:       rawURL,
		ExpiredAt: time.Now().Add(exp),
	}, nil
}

// UploadObjectFromURL uploads an object from an HTTP URL. The HTTP clients download the file and upload it to the storage bucket.
func (m *Manager) UploadObjectFromURL(ctx context.Context, bucketInfo common.StorageBucketArguments, objectName string, httpRequest *common.HTTPRequestOptions, opts *common.PutStorageObjectOptions) (*common.StorageUploadInfo, error) {
	client, bucketName, err := m.GetClientAndBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}

	var contentLength int64 = -1
	maxUploadSizeBytes := m.runtime.MaxUploadSizeMBs * 1024 * 1024

	if httpRequest.Method == nil || (*httpRequest.Method == "" || *httpRequest.Method == http.MethodGet) {
		contentLength = m.contentLengthFromHEAD(ctx, httpRequest)

		if contentLength > maxUploadSizeBytes {
			return nil, maxUploadSizeLimitError(m.runtime.MaxUploadSizeMBs)
		}
	}

	resp, err := m.httpClient.Request(ctx, httpRequest)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.ContentLength > 0 {
		contentLength = resp.ContentLength
	}

	respBody := resp.Body

	if contentLength < 0 {
		var buf bytes.Buffer

		contentLength, err = io.Copy(&buf, resp.Body)
		if err != nil {
			return nil, schema.UnprocessableContentError(err.Error(), nil)
		}

		resp.Body.Close()

		respBody = io.NopCloser(&buf)
	}

	if contentLength > maxUploadSizeBytes {
		return nil, maxUploadSizeLimitError(m.runtime.MaxUploadSizeMBs)
	}

	if contentType := resp.Header.Get(common.HeaderContentType); contentType != "" && opts.ContentType == "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return nil, schema.UnprocessableContentError("invalid response content type: "+err.Error(), nil)
		}

		opts.ContentType = contentType

		if mediaType == common.ContentTypeTextPlain {
			newContentType := common.ContentTypeFromFilePath(httpRequest.URL)
			if newContentType != "" {
				opts.ContentType = newContentType
			}
		}
	}

	if contentLanguage := resp.Header.Get(common.HeaderContentLanguage); contentLanguage != "" && opts.ContentLanguage == "" {
		opts.ContentLanguage = contentLanguage
	}

	result, err := client.PutObject(ctx, bucketName, objectName, opts, respBody, contentLength)
	if err != nil {
		return nil, err
	}

	result.Bucket = bucketName
	result.ClientID = string(client.id)

	return result, nil
}

// Try sending an HEAD request to get the content length
// and validate before requesting the GET request.
// It avoids reading the entire response body to estimate the file size.
func (m *Manager) contentLengthFromHEAD(ctx context.Context, httpRequest *common.HTTPRequestOptions) int64 {
	httpRequest.Method = utils.ToPtr(common.DownloadHTTPMethod(http.MethodHead))

	resp, err := m.httpClient.Request(ctx, httpRequest)
	if err != nil {
		slog.Debug(fmt.Sprintf("failed to send HEAD request: %s", err), slog.String("url", httpRequest.URL))

		return -1
	}

	if resp.StatusCode >= 300 {
		slog.Debug("failed to send HEAD request: "+resp.Status, slog.String("url", httpRequest.URL))

		return -1
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	httpRequest.Method = utils.ToPtr(common.DownloadHTTPMethod(http.MethodGet))

	return resp.ContentLength
}

func maxUploadSizeLimitError(mbs int64) error {
	return schema.UnprocessableContentError(fmt.Sprintf("file size > %d MB is not allowed to be upload directly. Please use presignedPutObject function for large files", mbs), nil)
}
