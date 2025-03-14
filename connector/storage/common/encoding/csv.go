package encoding

import (
	"encoding/csv"
	"errors"
	"io"
)

const (
	contentTypeTextCSV                   = "text/csv"
	contentTypeTextXCSV                  = "text/x-csv"
	contentTypeTextCommaSeparatedValues  = "text/comma-separated-values"
	contentTypeTextXCommaSeparatedValues = "text/x-comma-separated-values"
	contentTypeTextTabSeparatedValues    = "text/tab-separated-values"
	contentTypeApplicationCSV            = "application/csv"
	contentTypeApplicationXCSV           = "application/x-csv"
)

var enums_contentTypeCSV = []string{
	contentTypeTextCSV,
	contentTypeTextXCSV,
	contentTypeTextCommaSeparatedValues,
	contentTypeTextXCommaSeparatedValues,
	contentTypeTextTabSeparatedValues,
	contentTypeApplicationCSV,
	contentTypeApplicationXCSV,
}

func decodeArbitraryCSV(reader io.Reader, contentType string) ([][]string, error) {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	if contentType == contentTypeTextTabSeparatedValues {
		r.Comma = '\t'
	}

	rows := [][]string{}

	for {
		record, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		rows = append(rows, record)
	}

	return rows, nil
}
