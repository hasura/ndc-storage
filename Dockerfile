# build context at repo root: docker build -f Dockerfile .
FROM golang:1.24 AS builder

WORKDIR /app

ARG VERSION
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build \
  -ldflags "-X github.com/hasura/ndc-storage/configuration/version.BuildVersion=${VERSION}" \
  -v -o ndc-cli ./server

RUN mkdir /data && chmod 755 -R /data

# stage 2: production image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy the binary to the production image from the builder stage.
COPY --from=builder /etc/mime.types /etc/mime.types
COPY --from=builder /app/ndc-cli /ndc-cli
COPY --from=builder --chown=65532:65532 /data /home/nonroot/data
ENV HASURA_CONFIGURATION_DIRECTORY=/etc/connector

ENTRYPOINT ["/ndc-cli"]

# Run the web service on container startup.
CMD ["serve"]