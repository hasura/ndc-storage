package fs

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hasura/ndc-storage/connector/storage/common"
	"github.com/spf13/afero"
)

type objectWalker struct {
	client     afero.Fs
	bucketName string
	options    *common.ListStorageObjectsOptions
	startAfter string
	started    bool
	predicate  func(string) bool
	result     *common.StorageObjectListResults
}

// NewObjectWalker creates an objectWalker instance.
func NewObjectWalker(
	client afero.Fs,
	bucketName string,
	options *common.ListStorageObjectsOptions,
	predicate func(string) bool,
) *objectWalker {
	startAfter := strings.TrimRight(options.StartAfter, "/")

	return &objectWalker{
		client:     client,
		bucketName: bucketName,
		options:    options,
		predicate:  predicate,
		startAfter: startAfter,
		started:    startAfter == "",
		result:     &common.StorageObjectListResults{},
	}
}

// WalkDir walks and filters child objects in the directory.
func (ow *objectWalker) WalkDir(root string) (*common.StorageObjectListResults, error) {
	err := ow.walkDir(root)
	if err != nil {
		return nil, err
	}

	return ow.result, nil
}

// WalkDirEntries walks and filters child objects in the directory.
func (ow *objectWalker) WalkDirEntries(root string) (*common.StorageObjectListResults, error) {
	err := ow.walkDirEntries(root)
	if err != nil {
		return nil, err
	}

	return ow.result, nil
}

func (ow *objectWalker) walkDir(root string) error {
	rootStat, err := lstatIfPossible(ow.client, filepath.Join(ow.bucketName, root))
	if err != nil {
		if errors.Is(err, afero.ErrFileNotFound) {
			return nil
		}

		return err
	}

	if !rootStat.IsDir() {
		ow.addObject(serializeStorageObject(root, rootStat))

		return nil
	}

	return ow.walkDirEntries(root)
}

func (ow *objectWalker) walkDirEntries(root string) error {
	dir, err := ow.client.Open(filepath.Join(ow.bucketName, root))
	if err != nil {
		return err
	}

	names, err := dir.Readdirnames(-1)
	_ = dir.Close()

	if err != nil {
		return err
	}

	slices.Sort(names)

	for i, name := range names {
		var stopped bool

		relPath := filepath.Join(root, name)
		absPath := filepath.Join(ow.bucketName, relPath)
		stat, err := lstatIfPossible(ow.client, absPath)

		switch {
		case err != nil:
			if errors.Is(err, afero.ErrFileNotFound) {
				continue
			}

			stopped = ow.addObject(common.StorageObject{
				Bucket: ow.bucketName,
				Name:   relPath,
			})
		case !ow.options.Recursive || !stat.IsDir():
			stopped = ow.addObject(serializeStorageObject(relPath, stat))
		default:
			err = ow.walkDirEntries(relPath)
			if err != nil {
				return err
			}

			stopped = ow.hasMaxResults()
		}

		if stopped {
			ow.result.PageInfo.HasNextPage = ow.result.PageInfo.HasNextPage || i < len(names)-1

			return nil
		}
	}

	return nil
}

func (ow *objectWalker) addObject(object common.StorageObject) bool {
	if !ow.started {
		ow.started = ow.startAfter == object.Name

		return false
	}

	if ow.predicate != nil && !ow.predicate(object.Name) {
		return false
	}

	ow.result.Objects = append(ow.result.Objects, object)

	return ow.hasMaxResults()
}

func (ow *objectWalker) hasMaxResults() bool {
	return ow.options.MaxResults > 0 && len(ow.result.Objects) >= ow.options.MaxResults
}
