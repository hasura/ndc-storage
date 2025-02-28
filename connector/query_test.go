package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hasura/ndc-sdk-go/connector"
	"github.com/hasura/ndc-sdk-go/ndctest"
	"github.com/hasura/ndc-sdk-go/schema"
	"gotest.tools/v3/assert"
)

func TestConnectorQueries(t *testing.T) {
	connectorHost := "http://localhost:8080"
	clientIDs := []string{"minio", "azblob", "gcs"}

	for _, cid := range clientIDs {
		t.Run("create_bucket_"+cid, func(t *testing.T) {
			procedureRequest := schema.MutationRequest{
				CollectionRelationships: schema.MutationRequestCollectionRelationships{},
				Operations:              []schema.MutationOperation{},
			}

			for i := range 10 {
				procedureRequest.Operations = append(procedureRequest.Operations, schema.MutationOperation{
					Type: schema.MutationOperationProcedure,
					Name: "createStorageBucket",
					Arguments: []byte(fmt.Sprintf(`{
					"clientId": "%s",
					"name": "dummy-bucket-%d"
				}`, cid, i)),
					Fields: schema.NewNestedObject(map[string]schema.FieldEncoder{
						"success": schema.NewColumnField("success", nil),
					}).Encode(),
				})
			}

			rawBody, err := json.Marshal(procedureRequest)
			assert.NilError(t, err)

			resp, err := http.DefaultClient.Post(connectorHost+"/mutation", "application/json", bytes.NewReader(rawBody))
			assert.NilError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}

	objectFixtures := map[string]string{
		"movies/1900s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1900s.json",
		"movies/1910s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1910s.json",
		"movies/1920s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1920s.json",
		"movies/1930s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1930s.json",
		"movies/1940s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1940s.json",
		"movies/1950s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1950s.json",
		"movies/1960s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1960s.json",
		"movies/1970s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1970s.json",
		"movies/1980s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1980s.json",
		"movies/1990s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-1990s.json",
		"movies/2000s/movies.json": "https://raw.githubusercontent.com/prust/wikipedia-movie-data/refs/heads/master/movies-2000s.json",
	}

	for key, value := range objectFixtures {
		resp, err := http.DefaultClient.Get(value)
		assert.NilError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		rawBody, err := io.ReadAll(resp.Body)
		assert.NilError(t, err)
		resp.Body.Close()

		for _, cid := range clientIDs {
			t.Run(fmt.Sprintf("upload_object_%s/%s", cid, key), func(t *testing.T) {
				arguments := map[string]any{
					"clientId": cid,
					"bucket":   "dummy-bucket-0",
					"data":     string(rawBody),
					"object":   key,
					"options": map[string]any{
						"cacheControl":       "max-age=100",
						"contentDisposition": "attachment",
						"contentLanguage":    "en-US",
						"contentType":        "application/json",
						"expires":            "2099-01-01T00:00:00Z",
						"sendContentMd5":     true,
						"metadata": []map[string]any{
							{
								"key":   "Foo",
								"value": "Bar",
							},
						},
						"tags": []map[string]any{
							{
								"key":   "category",
								"value": "movie",
							},
						},
					},
				}

				rawArguments, err := json.Marshal(arguments)
				assert.NilError(t, err)

				procedureRequest := schema.MutationRequest{
					CollectionRelationships: schema.MutationRequestCollectionRelationships{},
					Operations: []schema.MutationOperation{
						{
							Type:      schema.MutationOperationProcedure,
							Name:      "uploadStorageObjectAsText",
							Arguments: rawArguments,
							Fields: schema.NewNestedObject(map[string]schema.FieldEncoder{
								"name": schema.NewColumnField("name", nil),
								"size": schema.NewColumnField("size", nil),
							}).Encode(),
						},
					},
				}

				uploadBytes, err := json.Marshal(procedureRequest)
				assert.NilError(t, err)

				resp, err = http.DefaultClient.Post(connectorHost+"/mutation", "application/json", bytes.NewReader(uploadBytes))
				assert.NilError(t, err)
				resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	}

	setConnectorTestEnv(t)

	for _, dir := range []string{"bucket", "object"} {
		ndctest.TestConnector(t, &Connector{}, ndctest.TestConnectorOptions{
			Configuration: "../tests/configuration",
			TestDataDir:   filepath.Join("testdata", dir),
		})
	}
}

func TestMaxDownloadSizeValidation(t *testing.T) {
	setConnectorTestEnv(t)

	server, err := connector.NewServer(&Connector{}, &connector.ServerOptions{
		Configuration: "../tests/configuration",
	}, connector.WithoutRecovery())
	assert.NilError(t, err)

	httpServer := server.BuildTestServer()
	defer httpServer.Close()

	getQueryBody := func(name string) string {
		return fmt.Sprintf(`{
		"arguments": {
			"clientId": {
				"type": "literal",
				"value": "minio"
			},
			"bucket": {
				"type": "literal",
				"value": "dummy-bucket-0"
			},
			"object": {
				"type": "literal",
				"value": "movies/2000s/movies.json"
			}
		},
		"collection": "%s",
		"collection_relationships": {},	
		"query": {
			"fields": {
				"__value": {
					"column": "__value",
					"fields": {
						"fields": {
							"data": {
								"column": "data",
								"type": "column"
							}
						},
						"type": "object"
					},
					"type": "column"
				}
			}
		}
	}`, name)
	}

	testCases := []struct {
		Name               string
		MaxDownloadSizeMBs float64
	}{
		{
			Name:               "downloadStorageObjectAsBase64",
			MaxDownloadSizeMBs: 1.33,
		},
		{
			Name:               "downloadStorageObjectAsText",
			MaxDownloadSizeMBs: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			resp, err := http.DefaultClient.Post(httpServer.URL+"/query", "application/json", strings.NewReader(getQueryBody(tc.Name)))
			assert.NilError(t, err)
			assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
			var respBody schema.ErrorResponse
			assert.NilError(t, json.NewDecoder(resp.Body).Decode(&respBody))
			assert.Equal(t, respBody.Message, fmt.Sprintf("file size >= %.2f MB is not allowed to be downloaded directly. Please use presignedGetObject function for large files", tc.MaxDownloadSizeMBs))
			resp.Body.Close()
		})
	}
}
