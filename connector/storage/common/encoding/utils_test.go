package encoding

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestContentTypeFromExtension(t *testing.T) {
	testCases := []struct {
		Path        string
		ContentType string
	}{
		{
			Path:        "https://raw.githubusercontent.com/samayo/country-json/refs/heads/master/src/country-by-name.json",
			ContentType: ContentTypeApplicationJSON,
		},
		{
			Path:        "/path/to/file.txt",
			ContentType: ContentTypeTextPlain,
		},
		{
			Path:        "/path/to/file.csv",
			ContentType: "text/csv",
		},
		{
			Path:        "public/hello.csv",
			ContentType: "text/csv",
		},
		{
			Path:        "/path/to/file.tsv",
			ContentType: "text/tab-separated-values",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Path, func(t *testing.T) {
			contentType := ContentTypeFromFilePath(tc.Path)
			assert.Assert(t, contentType == tc.ContentType || strings.HasPrefix(contentType, tc.ContentType+";"))
		})
	}
}

func TestTransposeMatrixString(t *testing.T) {
	testCases := []struct {
		Input    [][]string
		Expected [][]string
	}{
		{
			Input:    [][]string{},
			Expected: [][]string{},
		},
		{
			Input: [][]string{
				{"id", "1", "2", "3", "4"},
				{"name", "Jack", "John", "Tom", "Jane"},
				{"active", "true", "false", "true", "false"},
			},
			Expected: [][]string{
				{"id", "name", "active"},
				{"1", "Jack", "true"},
				{"2", "John", "false"},
				{"3", "Tom", "true"},
				{"4", "Jane", "false"},
			},
		},
	}

	for _, tc := range testCases {
		assert.DeepEqual(t, tc.Expected, transposeMatrixString(tc.Input))
	}
}
