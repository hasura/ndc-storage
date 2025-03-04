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
        "object": "public/random-failed.txt"
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
			Name:    "uploadStorageObjectAsText",
			Content: dataText,
		},
		{
			Name:    "uploadStorageObjectAsBase64",
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
			assert.Equal(t, respBody.Message, fmt.Sprintf("file size > %d MB is not allowed to be upload directly. Please use presignedPutObject function for large files", 1))
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
			"name": "uploadStorageObjectFromUrl",
			"arguments": {
				"bucket": "minio-bucket-test",
				"object": "movies-2000s.json",
				"url": "https://drive.google.com/uc?id=1IXQDp8Um3d-o7ysZLxkDyuvFj9gtlxqz&export=download"
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
	assert.Equal(t, respBody.Message, fmt.Sprintf("file size > %d MB is not allowed to be upload directly. Please use presignedPutObject function for large files", 1))
	resp.Body.Close()
}
