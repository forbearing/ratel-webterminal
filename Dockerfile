FROM golang:1.18 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY ratel-webterminal.go ratel-webterminal.go
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ratel-webterminal ratel-webterminal.go


# Use distroless as minimal base image to package the ratel-webterminal binary
#FROM gcr.io/distroless/static:nonroot
#FROM alpine:3.15
FROM debian

WORKDIR /

COPY --from=builder /workspace/ratel-webterminal .
COPY frontend/ frontend/
#USER 65532:65532
USER 0:0

ENTRYPOINT ["/ratel-webterminal"]
