package fs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hasura/ndc-sdk-go/v2/schema"
	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/sync/errgroup"
)

// ListObjects list objects in a bucket.
func (c *Client) ListObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	_, span := c.startOtelSpan(ctx, "ListObjects", bucketName)
	defer span.End()

	var count int

	result := &common.StorageObjectListResults{
		Objects: make([]common.StorageObject, 0),
	}

	root := filepath.Clean(opts.Prefix)
	prefixPath := filepath.Join(bucketName, root)

	span.SetAttributes(
		attribute.String("storage.object.prefix", prefixPath),
		attribute.Bool("storage.option.recursive", opts.Recursive),
	)

	prefixFile, err := c.lstatIfPossible(prefixPath)
	if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	if prefixFile == nil {
		baseDir := filepath.Dir(prefixPath)

		filterFn := func(name string) bool {
			if root != "" && !strings.HasPrefix(name, root) {
				return false
			}

			return predicate == nil || predicate(name)
		}

		result, err = NewObjectWalker(c.client, bucketName, opts, filterFn).WalkDir(baseDir)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, schema.UnprocessableContentError(err.Error(), nil)
		}

		return result, nil
	}

	if !prefixFile.IsDir() {
		if predicate == nil || predicate(root) {
			result.Objects = append(result.Objects, serializeStorageObject(root, prefixFile))
		}

		return result, nil
	}

	result, err = NewObjectWalker(c.client, bucketName, opts, predicate).WalkDirEntries(root)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	span.SetAttributes(attribute.Int("storage.object_count", count))

	return result, nil
}

// ListIncompleteUploads list partially uploaded objects in a bucket.
func (c *Client) ListIncompleteUploads(
	ctx context.Context,
	bucketName string,
	args common.ListIncompleteUploadsOptions,
) ([]common.StorageObjectMultipartInfo, error) {
	return []common.StorageObjectMultipartInfo{}, nil
}

// RemoveIncompleteUpload removes a partially uploaded object.
func (c *Client) RemoveIncompleteUpload(
	ctx context.Context,
	bucketName string,
	objectName string,
) error {
	return nil
}

// ListDeletedObjects list deleted objects in a bucket.
func (c *Client) ListDeletedObjects(
	ctx context.Context,
	bucketName string,
	opts *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) (*common.StorageObjectListResults, error) {
	_, span := c.startOtelSpan(ctx, "ListDeletedObjects", bucketName)
	defer span.End()

	results := &common.StorageObjectListResults{
		Objects: []common.StorageObject{},
	}

	return results, nil
}

// GetObject returns a stream of the object data. Most of the common errors occur when reading the stream.
func (c *Client) GetObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (io.ReadCloser, error) {
	_, span := c.startOtelSpan(ctx, "GetObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	filePath := filepath.Join(bucketName, objectName)

	object, err := c.client.Open(filePath)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	return object, nil
}

// PutObject uploads objects that are less than 128MiB in a single PUT operation. For objects that are greater than 128MiB in size,
// PutObject seamlessly uploads the object as parts of 128MiB or more depending on the actual file size. The max upload size for an object is 5TB.
func (c *Client) PutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts *common.PutStorageObjectOptions,
	reader io.Reader,
	objectSize int64,
) (*common.StorageUploadInfo, error) {
	_, span := c.startOtelSpan(ctx, "PutObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Int64("http.response.body.size", objectSize),
	)

	filePath := filepath.Join(bucketName, objectName)
	if strings.Contains(objectName, "/") || strings.Contains(objectName, "\\") {
		// ensure that the directory exists
		baseDir := filepath.Dir(filePath)

		err := c.client.MkdirAll(baseDir, os.FileMode(c.permissions.Directory))
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			return nil, schema.UnprocessableContentError(err.Error(), nil)
		}
	}

	file, err := c.client.OpenFile(
		filePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		os.FileMode(c.permissions.File),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	defer func() {
		_ = file.Close()
	}()

	if _, err := io.Copy(file, reader); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	return &common.StorageUploadInfo{
		Bucket: bucketName,
		Name:   objectName,
		Size:   &objectSize,
	}, nil
}

// CopyObject creates or replaces an object through server-side copying of an existing object.
// It supports conditional copying, copying a part of an object and server-side encryption of destination and decryption of source.
// To copy multiple source objects into a single destination object see the ComposeObject API.
func (c *Client) CopyObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	src common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	ctx, span := c.startOtelSpan(ctx, "CopyObject", dest.Bucket)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", dest.Name),
		attribute.String("storage.copy_source", src.Name),
	)

	srcPath := filepath.Join(src.Bucket, src.Name)

	srcFile, err := c.client.Open(srcPath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	defer func() {
		_ = srcFile.Close()
	}()

	result, err := c.PutObject(
		ctx,
		dest.Bucket,
		dest.Name,
		&common.PutStorageObjectOptions{},
		srcFile,
		-1,
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	return result, nil
}

// ComposeObject creates an object by concatenating a list of source objects using server-side copying.
func (c *Client) ComposeObject(
	ctx context.Context,
	dest common.StorageCopyDestOptions,
	sources []common.StorageCopySrcOptions,
) (*common.StorageUploadInfo, error) {
	return nil, errNotSupported
}

// StatObject fetches metadata of an object.
func (c *Client) StatObject(
	ctx context.Context,
	bucketName, objectName string,
	opts common.GetStorageObjectOptions,
) (*common.StorageObject, error) {
	_, span := c.startOtelSpan(ctx, "StatObject", bucketName)
	defer span.End()

	span.SetAttributes(attribute.String("storage.key", objectName))

	filePath := filepath.Join(bucketName, objectName)

	object, err := c.lstatIfPossible(filePath)
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return nil, nil
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, schema.UnprocessableContentError(err.Error(), nil)
	}

	objectSize := object.Size()
	result := &common.StorageObject{
		Bucket:       bucketName,
		Name:         objectName,
		Size:         &objectSize,
		IsDirectory:  object.IsDir(),
		LastModified: object.ModTime(),
	}

	return result, nil
}

// RemoveObject removes an object with some specified options.
func (c *Client) RemoveObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.RemoveStorageObjectOptions,
) error {
	_, span := c.startOtelSpan(ctx, "RemoveObject", bucketName)
	defer span.End()

	span.SetAttributes(
		attribute.String("storage.key", objectName),
		attribute.Bool("storage.options.force_delete", opts.ForceDelete),
		attribute.Bool("storage.options.governance_bypass", opts.GovernanceBypass),
	)

	if opts.VersionID != "" {
		span.SetAttributes(attribute.String("storage.options.version", opts.VersionID))
	}

	filePath := filepath.Join(bucketName, objectName)

	err := c.client.RemoveAll(filePath)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return schema.UnprocessableContentError(err.Error(), nil)
	}

	return nil
}

// RemoveObjects removes a list of objects obtained from an input channel. The call sends a delete request to the server up to 1000 objects at a time.
// The errors observed are sent over the error channel.
func (c *Client) RemoveObjects(
	ctx context.Context,
	bucketName string,
	opts *common.RemoveStorageObjectsOptions,
	predicate func(string) bool,
) []common.RemoveStorageObjectError {
	ctx, span := c.startOtelSpan(ctx, "RemoveObjects", bucketName)
	defer span.End()

	objects, err := c.ListObjects(ctx, bucketName, &opts.ListStorageObjectsOptions, predicate)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return []common.RemoveStorageObjectError{
			{
				Error: err.Error(),
			},
		}
	}

	span.SetAttributes(attribute.Int("storage.object_count", len(objects.Objects)))

	errs := make([]common.RemoveStorageObjectError, 0)

	if opts.NumThreads <= 1 {
		for _, object := range objects.Objects {
			filePath := filepath.Join(object.Bucket, object.Name)

			err := c.client.RemoveAll(filePath)
			if err != nil {
				errs = append(errs, common.RemoveStorageObjectError{
					ObjectName: object.Name,
					Error:      err.Error(),
				})
			}
		}
	} else {
		eg := errgroup.Group{}
		eg.SetLimit(opts.NumThreads)

		removeFunc := func(name string) {
			eg.Go(func() error {
				filePath := filepath.Join(bucketName, name)

				err := c.client.RemoveAll(filePath)
				if err != nil {
					errs = append(errs, common.RemoveStorageObjectError{
						ObjectName: name,
						Error:      err.Error(),
					})
				}

				return nil
			})
		}

		for _, item := range objects.Objects {
			removeFunc(item.Name)
		}

		err := eg.Wait()
		if err != nil {
			return []common.RemoveStorageObjectError{
				{
					Error: err.Error(),
				},
			}
		}
	}

	if len(errs) > 0 {
		bs, err := json.Marshal(errs)
		if err != nil {
			slog.Error(err.Error())
		}

		span.SetAttributes(attribute.String("errors", string(bs)))
		span.SetStatus(codes.Error, "failed to remove objects")
	}

	return errs
}

// UpdateObject updates object configurations.
func (c *Client) UpdateObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.UpdateStorageObjectOptions,
) error {
	_, span := c.startOtelSpan(ctx, "UpdateObject", bucketName)
	defer span.End()

	return nil
}

// RestoreObject restores a soft-deleted object.
func (c *Client) RestoreObject(ctx context.Context, bucketName string, objectName string) error {
	return errNotSupported
}

// PresignedGetObject generates a presigned URL for HTTP GET operations. Browsers/Mobile clients may point to this URL to directly download objects even if the bucket is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational.
// The maximum expiry is 604800 seconds (i.e. 7 days) and minimum is 1 second.
func (c *Client) PresignedGetObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts common.PresignedGetStorageObjectOptions,
) (string, error) {
	return "", errNotSupported
}

// PresignedPutObject generates a presigned URL for HTTP PUT operations. Browsers/Mobile clients may point to this URL to upload objects directly to a bucket even if it is private.
// This presigned URL can have an associated expiration time in seconds after which it is no longer operational. The default expiry is set to 7 days.
func (c *Client) PresignedPutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	expiry time.Duration,
) (string, error) {
	return "", errNotSupported
}
