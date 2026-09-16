#!/bin/bash
set -euo pipefail

export GO_VERSION="$(awk '$1 == "go" { print $2 }' go.mod)"
compose=(docker compose -p ndc-storage-test -f compose.yaml -f compose.test.yaml)
trap '"${compose[@]}" down -v' EXIT

mkdir -p ./tmp
NDC_SPEC_BASE_URL=https://github.com/hasura/ndc-spec/releases/download/v0.2.10

if [ ! -f ./tmp/ndc-test ]; then
  if [ "$(uname -s)" == "Darwin" ] && [ "$(uname -m)" == "arm64" ]; then
    curl -fL "$NDC_SPEC_BASE_URL/ndc-test-aarch64-apple-darwin" -o ./tmp/ndc-test
  elif [ "$(uname -s)" == "Darwin" ]; then
    curl -fL "$NDC_SPEC_BASE_URL/ndc-test-x86_64-apple-darwin" -o ./tmp/ndc-test
  else
    curl -fL "$NDC_SPEC_BASE_URL/ndc-test-x86_64-unknown-linux-gnu" -o ./tmp/ndc-test
  fi

  chmod +x ./tmp/ndc-test
fi

http_wait() {
  printf '%s:\t ' "$1"
  for i in {1..120};
  do
    local code
    code="$("${compose[@]}" exec -T test-runner curl -s -o /dev/null -m 2 -w '%{http_code}' "$1")" || true
    if [ "$code" != "200" ]; then
      printf "."
      sleep 1
    else
      printf '\r\033[K%s:\t OK\n' "$1"
      return 0
    fi
  done
  printf '\nERROR: cannot connect to %s.\n' "$1"
  exit 1
}

wait_services() {
  http_wait http://ndc-storage:8080/health
  http_wait http://local.hasura.dev:9000/minio/health/live
}

run_test() {
  "${compose[@]}" up -d minio s3mock azurite gcp-storage-emulator ndc-storage test-runner
  wait_services

  # go tests
  "${compose[@]}" exec -T -e CONFIG_DIR="$1" test-runner \
    go test -v -coverpkg=./... -race -timeout 5m -coverprofile=coverage.out.tmp ./...
  "${compose[@]}" down -v
}

"${compose[@]}" up -d --build minio s3mock azurite gcp-storage-emulator ndc-storage test-runner
wait_services
./tmp/ndc-test test --endpoint "http://$("${compose[@]}" port ndc-storage 8080)"

run_test ../tests/configuration-static
run_test ../tests/configuration

grep -Ev 'main.go|version.go|jsonschema' coverage.out.tmp > coverage.out
rm coverage.out.tmp
