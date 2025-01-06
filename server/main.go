package main

import (
	"github.com/hasura/ndc-sdk-go/connector"
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
	if err := connector.Start(
		&storage.Connector{},
		connector.WithMetricsPrefix("ndc_storage"),
		connector.WithDefaultServiceName("ndc-storage"),
		connector.WithVersion(version.BuildVersion),
	); err != nil {
		panic(err)
	}
}
