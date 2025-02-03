module github.com/hasura/ndc-storage

go 1.23

require (
	cloud.google.com/go/storage v1.50.0
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.17.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.8.1
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.0
	github.com/alecthomas/kong v1.7.0
	github.com/go-viper/mapstructure/v2 v2.2.1
	github.com/hasura/ndc-sdk-go v1.7.1-0.20250203083219-765c8932cef4
	github.com/invopop/jsonschema v0.13.0
	github.com/lmittmann/tint v1.0.7
	github.com/minio/md5-simd v1.1.2
	github.com/minio/minio-go/v7 v7.0.84
	go.opentelemetry.io/otel v1.34.0
	go.opentelemetry.io/otel/trace v1.34.0
	golang.org/x/sync v0.10.0
	google.golang.org/api v0.219.0
	gopkg.in/yaml.v3 v3.0.1
	gotest.tools/v3 v3.5.1
)

require (
	cel.dev/expr v0.19.2 // indirect
	cloud.google.com/go v0.118.1 // indirect
	cloud.google.com/go/auth v0.14.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.7 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	cloud.google.com/go/iam v1.3.1 // indirect
	cloud.google.com/go/monitoring v1.23.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.10.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.3.3 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.26.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.50.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.50.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelslog v0.9.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.34.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.10.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.34.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.56.0 // indirect
	go.opentelemetry.io/otel/log v0.10.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk v1.34.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.10.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.34.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/time v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20250127172529-29210b9bc287 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250127172529-29210b9bc287 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250127172529-29210b9bc287 // indirect
	google.golang.org/grpc v1.70.0 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)
