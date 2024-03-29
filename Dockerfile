# Build the manager binary
FROM golang:1.18 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN apt update && apt install ca-certificates libgnutls30 -y
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY commands/ commands/
COPY controllers/ controllers/
COPY errors/ errors/
COPY operations/ operations/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use debian-base instead since unable to overwrite the file on the host with distroless
FROM k8s.gcr.io/debian-base:v1.0.0
RUN apt update && apt install openssh-client -y
WORKDIR /
COPY --from=builder /workspace/manager .

ENTRYPOINT ["/manager"]
