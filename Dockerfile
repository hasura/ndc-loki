# build context at repo root: docker build -f Dockerfile .
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -v -o ndc-loki ./server

# stage 2: production image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/ndc-loki /ndc-loki

ENV HASURA_CONFIGURATION_DIRECTORY=/etc/connector

ENTRYPOINT ["/ndc-loki"]

# Run the web service on container startup.
CMD ["serve"]