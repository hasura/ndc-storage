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

func transposeMatrixString(matrix [][]string) [][]string {
	totalRows := len(matrix)
	if totalRows == 0 {
		return matrix
	}

	totalCols := len(matrix[0])
	results := make([][]string, totalCols)

	for y := range totalCols {
		results[y] = make([]string, totalRows)
	}

	for y, row := range matrix {
		for x, col := range row {
			results[x][y] = col
		}
	}

	return results
}
