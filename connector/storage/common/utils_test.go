package common

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
			ContentType: ContentTypeTextCSV,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Path, func(t *testing.T) {
			contentType := ContentTypeFromFilePath(tc.Path)
			assert.Assert(t, contentType == tc.ContentType || strings.HasPrefix(contentType, tc.ContentType+";"))
		})
	}
}
