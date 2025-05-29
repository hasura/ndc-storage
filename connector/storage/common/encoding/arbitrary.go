package encoding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"slices"
	"strings"
)

// DecodeArbitraryData guesses and decodes the arbitrary data of a file from the content type.
func DecodeArbitraryData(
	ctx context.Context,
	name string,
	contentType string,
	reader io.Reader,
) (any, error) {
	result, decoded, err := decodeArbitraryDataFromContentType(ctx, reader, contentType)
	if err != nil {
		return nil, err
	}

	if decoded {
		return result, nil
	}

	fileContentType := ContentTypeFromFilePath(name)

	result, decoded, err = decodeArbitraryDataFromContentType(ctx, reader, fileContentType)
	if err != nil {
		return nil, err
	}

	if !decoded {
		ct := contentType
		if ct == "" {
			ct = fileContentType
		}

		return nil, fmt.Errorf("failed to decode file %s, unsupported content type %s", name, ct)
	}

	return result, nil
}

func decodeArbitraryDataFromContentType(
	ctx context.Context,
	reader io.Reader,
	contentType string,
) (any, bool, error) {
	if contentType == "" {
		return nil, false, nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, false, nil //nolint:nilerr
	}

	switch {
	case mediaType == ContentTypeApplicationJSON || strings.Contains(mediaType, "+json"):
		var result any
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
			return nil, false, err
		}

		return result, true, nil
	case slices.Contains(enums_contentTypeCSV, mediaType):
		r := createDefaultCsvReader(reader)
		r.Comma = evalCSVComma("", mediaType)

		result, err := decodeCSVMatrix(ctx, r)
		if err != nil {
			return nil, false, err
		}

		return result, true, nil
	case strings.HasPrefix(mediaType, "text/"):
		result, err := io.ReadAll(reader)
		if err != nil {
			return nil, false, err
		}

		return string(result), true, nil
	}

	return nil, false, nil
}
