package main

import (
	"github.com/hasura/ndc-sdk-go/v2/connector"
	"github.com/hasura/ndc-storage/configuration/version"
	storage "github.com/hasura/ndc-storage/connector"
)

// Start the connector server at http://localhost:8080
//
//	go run . serve
//
// See [NDC Go SDK] for more information.
//
// [NDC Go SDK]: https://github.com/hasura/ndc-sdk-go
func main() {
	err := connector.Start(
		&storage.Connector{},
		connector.WithMetricsPrefix("ndc_storage"),
		connector.WithDefaultServiceName("ndc-storage"),
		connector.WithVersion(version.BuildVersion),
	)
	if err != nil {
		panic(err)
	}
}
