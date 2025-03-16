package encoding

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// CSVDecodeOptions hold decode options for CSV.
type CSVDecodeOptions struct {
	// If the first row is not the header the result will be a 2-dimension matrix.
	NoHeader bool `json:"no_header,omitempty"`

	// The matrix is transposed.
	Transpose bool `json:"transpose,omitempty"`

	// Try to parse column values to JSON types.
	ParseJSON bool `json:"parse_json,omitempty"`

	// The field delimiter.
	Delimiter string `json:"delimiter,omitempty"`

	// Comment, if not 0, is the comment character. Lines beginning with the
	// Comment character without preceding whitespace are ignored.
	Comment string `json:"comment,omitempty"`

	// If LazyQuotes is true, a quote may appear in an unquoted field and a
	// non-doubled quote may appear in a quoted field.
	LazyQuotes *bool `json:"lazy_quotes"`

	// If TrimLeadingSpace is true, leading white space in a field is ignored.
	// This is done even if the field delimiter, Comma, is white space.
	TrimLeadingSpace *bool `json:"trim_leading_space"`
}

// NewReader creates a new CSV Reader instance from options.
func (cdo CSVDecodeOptions) NewReader(reader io.Reader) *csv.Reader {
	r := createDefaultCsvReader(reader)
	r.Comma = evalCSVComma(cdo.Delimiter, "")

	if cdo.LazyQuotes != nil {
		r.LazyQuotes = *cdo.LazyQuotes
	}

	if cdo.TrimLeadingSpace != nil {
		r.TrimLeadingSpace = *cdo.TrimLeadingSpace
	}

	if cdo.Comment != "" {
		r.Comment = rune(cdo.Comment[0])
	}

	return r
}

// DecodeCSV decodes the CSV content to a matrix or list of objects.
func DecodeCSV(ctx context.Context, reader io.Reader, options CSVDecodeOptions) (any, error) {
	matrix, err := decodeCSVMatrix(ctx, options.NewReader(reader))
	matrixLen := len(matrix)

	if err != nil || matrixLen == 0 {
		return matrix, err
	}

	if options.Transpose {
		matrix = transposeMatrixString(matrix)
		matrixLen = len(matrix)
	}

	if options.NoHeader {
		if !options.ParseJSON {
			return matrix, nil
		}

		results := make([][]any, matrixLen)

		for i, row := range matrix {
			result := make([]any, len(row))

			for j, cell := range row {
				value, _ := decodeCSVCellValue(cell)
				result[j] = value
			}

			results[i] = result
		}

		return results, nil
	}

	headerRow := matrix[0]
	results := make([]map[string]any, matrixLen-1)

	for i := 1; i < matrixLen; i++ {
		row := matrix[i]
		result := make(map[string]any)

		for j, key := range headerRow {
			if !options.ParseJSON {
				result[key] = row[j]

				continue
			}

			cell, _ := decodeCSVCellValue(row[j])
			result[key] = cell
		}

		results[i-1] = result
	}

	return results, nil
}

func decodeCSVMatrix(ctx context.Context, r *csv.Reader) ([][]string, error) {
	rows := [][]string{}

	for {
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		default:
			record, err := r.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return rows, nil
				}

				return nil, err
			}

			rows = append(rows, record)
		}
	}
}

func decodeCSVCellValue(cellValue string) (any, error) {
	switch strings.ToLower(cellValue) {
	case "":
		return "", nil
	case "null":
		return nil, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	if cellValue[0] == '[' || cellValue[0] == '{' || cellValue[0] == '"' {
		var result any
		if err := json.Unmarshal([]byte(cellValue), &result); err != nil {
			return cellValue, err
		}

		return result, nil
	}

	// try to decode number
	if cellValue[0] >= '0' || cellValue[0] <= '9' || cellValue[0] == '-' {
		numberResult, err := strconv.ParseFloat(cellValue, 64)
		if err == nil {
			return numberResult, nil
		}
	}

	return cellValue, nil
}

// IsValidCSVObject checks if the object is a valid CSV file.
func IsValidCSVObject(name string, contentType string) bool {
	return isValidCSVContentType(contentType) ||
		isValidCSVContentType(ContentTypeFromFilePath(name))
}

func isValidCSVContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)

	return err == nil && (mediaType == ContentTypeTextPlain || slices.Contains(enums_contentTypeCSV, contentType))
}

// CSVCommaFromContentType parses the csv comma from object name or content type.
func CSVCommaFromContentType(name string, contentType string) string {
	var mimeType string

	if contentType != "" {
		mimeType, _, _ = mime.ParseMediaType(contentType)
	}

	if mimeType == contentTypeTextCommaSeparatedValues || filepath.Ext(name) == ".tsv" {
		return "tab"
	}

	return ","
}

func evalCSVComma(comma string, contentType string) rune {
	switch comma {
	case "":
		if contentType == contentTypeTextTabSeparatedValues {
			return '\t'
		}

		return ','
	case "tab":
		return '\t'
	default:
		return rune(comma[0])
	}
}

func createDefaultCsvReader(reader io.Reader) *csv.Reader {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	return r
}
