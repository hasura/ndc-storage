package encoding

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/hasura/ndc-storage/connector/storage/common"
)

// DecodeArbitraryData guesses and decodes the arbitrary data of a file from the content type.
func DecodeArbitraryData(name string, contentType string, reader io.Reader) (any, error) {
	result, decoded, err := decodeArbitraryDataFromContentType(contentType, reader)
	if err != nil {
		return nil, err
	}

	if decoded {
		return result, nil
	}

	fileContentType := common.ContentTypeFromFilePath(name)

	result, decoded, err = decodeArbitraryDataFromContentType(fileContentType, reader)
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

func decodeArbitraryDataFromContentType(contentType string, reader io.Reader) (any, bool, error) {
	if contentType == "" {
		return nil, false, nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, false, nil
	}

	switch {
	case mediaType == common.ContentTypeApplicationJSON || strings.Contains(mediaType, "+json"):
		var result any
		if err := json.NewDecoder(reader).Decode(&result); err != nil {
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
