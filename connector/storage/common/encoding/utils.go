package encoding

import (
	"mime"
	"path/filepath"
)

// ContentTypeFromFilePath tries to guess the content type from the extension of file path.
func ContentTypeFromFilePath(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return ""
	}

	return mime.TypeByExtension(ext)
}
