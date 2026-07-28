ARG GO_VERSION=1.26.5
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags "-s -w" -o epson_exporter ./cmd/epson_exporter

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/epson_exporter /epson_exporter
EXPOSE 9788
USER 65532:65532
ENTRYPOINT ["/epson_exporter"]
