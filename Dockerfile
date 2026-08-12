# syntax=docker/dockerfile:1

# Defaults keep local builds self-contained. Renovate updates this alongside
# the matching Mise and module pins.
ARG GO_VERSION=1.26.5

# ---- Go build -------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BRANCH=unknown
ARG BUILD_DATE=unknown

RUN apk add --no-cache upx
WORKDIR /workspace

# Cache module downloads before copying source.
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags "-s -w -X github.com/prometheus/common/version.Version=${VERSION} -X github.com/prometheus/common/version.Revision=${COMMIT} -X github.com/prometheus/common/version.Branch=${BRANCH} -X github.com/prometheus/common/version.BuildUser=docker -X github.com/prometheus/common/version.BuildDate=${BUILD_DATE}" \
    -o epson_exporter ./cmd/epson_exporter
RUN upx --best --lzma epson_exporter

# ---- Runtime --------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/epson_exporter /epson_exporter
EXPOSE 9788
USER 65532:65532
ENTRYPOINT ["/epson_exporter"]
