package connector

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/schema"
	"gotest.tools/v3/assert"
)

func TestMaxUploadSizeValidation(t *testing.T) {
	setConnectorTestEnv(t)

	server, err := connector.NewServer(&Connector{}, &connector.ServerOptions{
		Configuration: "../tests/configuration",
	}, connector.WithoutRecovery())
	assert.NilError(t, err)

	httpServer := server.BuildTestServer()
	defer httpServer.Close()

	dataText := strings.Repeat("x", 10*1024*1024+1)
	dataTextBase64 := base64.StdEncoding.EncodeToString([]byte(dataText))

	getQueryBody := func(name string, content string) string {
		return fmt.Sprintf(`{
  "collection_relationships": {},
  "operations": [
    {
      "type": "procedure",
      "name": "%s",
      "arguments": {
        "bucket": "minio-bucket-test",
        "data": "%s",
        "name": "public/random-failed.txt"
      },
      "fields": {
        "fields": {
          "name": {
            "column": "name",
            "type": "column"
          },
          "size": {
            "column": "size",
            "type": "column"
          }
        },
        "type": "object"
      }
    }
  ]
}`, name, content)
	}

	testCases := []struct {
		Name    string
		Content string
	}{
		{
			Name:    "upload_storage_object_as_text",
			Content: dataText,
		},
		{
			Name:    "upload_storage_object_as_base64",
			Content: dataTextBase64,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := http.DefaultClient.Post(httpServer.URL+"/mutation", "application/json", strings.NewReader(getQueryBody(tc.Name, tc.Content)))
			assert.NilError(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			var respBody schema.ErrorResponse
			assert.NilError(t, json.NewDecoder(resp.Body).Decode(&respBody))
			assert.Equal(t, respBody.Message, fmt.Sprintf("file size > %d MB is not allowed to be upload directly. Please use presignedPutObject function for large files", 8))
			resp.Body.Close()
		})
	}
}

func TestMaxUploadSizeURLValidation(t *testing.T) {
	setConnectorTestEnv(t)

	server, err := connector.NewServer(&Connector{}, &connector.ServerOptions{
		Configuration: "../tests/configuration",
	}, connector.WithoutRecovery())
	assert.NilError(t, err)

	httpServer := server.BuildTestServer()
	defer httpServer.Close()

	queryBody := `{
		"operations": [
			{
			"type": "procedure",
			"name": "upload_storage_object_from_url",
			"arguments": {
				"bucket": "minio-bucket-test",
				"name": "movies-2000s.json",
				"url": "https://raw.githubusercontent.com/hasura/ndc-stripe/refs/heads/main/config/schema.json"
			},
			"fields": {
				"type": "object",
				"fields": {
				"name": {
					"type": "column",
					"column": "name",
					"fields": null
				},
				"size": {
					"type": "column",
					"column": "size",
					"fields": null
				}
				}
			}
			}
		],
		"collection_relationships": {}
	}`

	resp, err := http.DefaultClient.Post(httpServer.URL+"/mutation", "application/json", strings.NewReader(queryBody))
	assert.NilError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var respBody schema.ErrorResponse
	assert.NilError(t, json.NewDecoder(resp.Body).Decode(&respBody))
	assert.Equal(t, respBody.Message, fmt.Sprintf("file size > %d MB is not allowed to be upload directly. Please use presignedPutObject function for large files", 8))
	resp.Body.Close()
}
