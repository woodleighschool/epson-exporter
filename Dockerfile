FROM golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o epson_exporter ./cmd/epson_exporter

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/epson_exporter /bin/epson_exporter
EXPOSE 9788
USER 65532:65532
ENTRYPOINT ["/bin/epson_exporter"]
