# epson-exporter

[![Release](https://img.shields.io/github/v/release/woodleighschool/epson-exporter?display_name=tag&sort=semver)](https://github.com/woodleighschool/epson-exporter/releases/latest)
[![CI](https://github.com/woodleighschool/epson-exporter/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/woodleighschool/epson-exporter/actions/workflows/ci.yaml)
[![Go](https://img.shields.io/github/go-mod/go-version/woodleighschool/epson-exporter?logo=go)](https://github.com/woodleighschool/epson-exporter/blob/main/go.mod)
[![Container](https://img.shields.io/badge/container-ghcr.io-2496ED?logo=github&logoColor=white)](https://github.com/orgs/woodleighschool/packages/container/package/epson-exporter)
[![License](https://img.shields.io/github/license/woodleighschool/epson-exporter)](https://github.com/woodleighschool/epson-exporter/blob/main/LICENSE)

Prometheus multi-target exporter for Epson printer web interfaces. It reads status and usage pages exposed by WorkForce-style devices.

## 🚀 Usage

A container is published with each [release](https://github.com/woodleighschool/epson-exporter/releases/latest):

```bash
docker run --rm \
  --publish 9788:9788 \
  ghcr.io/woodleighschool/epson-exporter:rolling
```

Probe a printer:

```bash
curl 'http://127.0.0.1:9788/probe?target=https://192.0.2.10&module=default'
```

`/metrics` exposes exporter self-metrics. `/probe` performs the printer scrape.

A Prometheus scrape job looks like this:

```yaml
scrape_configs:
  - job_name: epson
    metrics_path: /probe
    params:
      module: [default]
    static_configs:
      - targets: ["https://192.0.2.10"]
        labels:
          printer: library
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [printer]
        target_label: instance
      - target_label: __address__
        replacement: epson-exporter:9788
```

## ⚙️ Configuration

The built-in `default` module uses a five-second timeout and Epson's standard status paths. It accepts locally issued printer certificates because these devices commonly redirect HTTP to HTTPS.

Use `--config.file=config.yml` only when a printer needs different paths, timeouts, or TLS settings. Start from [`config.example.yml`](config.example.yml), and validate it with the published container:

```bash
docker run --rm \
  --volume "$PWD/config.yml:/config.yml:ro" \
  ghcr.io/woodleighschool/epson-exporter:rolling \
  --config.file=/config.yml \
  --config.check
```

| Flag                   | Default         | Purpose                         |
| ---------------------- | --------------- | ------------------------------- |
| `--web.listen-address` | `:9788`         | HTTP listen address             |
| `--web.telemetry-path` | `/metrics`      | Self-metrics path               |
| `--config.file`        | Built-in module | YAML module configuration       |
| `--config.check`       | `false`         | Validate configuration and exit |

## 📈 Metrics

| Metric                                | Type    | Purpose                                          |
| ------------------------------------- | ------- | ------------------------------------------------ |
| `epson_printer_info`                  | gauge   | Printer identity, firmware, and network metadata |
| `epson_status_info`                   | gauge   | Printer, scanner, and hardware status            |
| `epson_scrape_success`                | gauge   | Whether the target scrape succeeded              |
| `epson_scrape_duration_seconds`       | gauge   | Target scrape duration                           |
| `epson_consumable_level_percent`      | gauge   | Ink or maintenance-box level                     |
| `epson_consumable_info`               | gauge   | Consumable identity                              |
| `epson_consumable_warning`            | gauge   | Printer-reported consumable warning              |
| `epson_paper_source_info`             | gauge   | Paper source metadata                            |
| `epson_paper_source_remaining_ratio`  | gauge   | Approximate remaining paper                      |
| `epson_first_print_timestamp_seconds` | gauge   | First printing date                              |
| `epson_pages_total`                   | counter | Lifetime pages by kind                           |
| `epson_pages_by_size_total`           | counter | Lifetime pages by size, sides, and colour mode   |
| `epson_pages_by_function_total`       | counter | Lifetime pages by function                       |
| `epson_pages_by_language_total`       | counter | Lifetime pages by printer language               |
| `epson_pages_by_interface_total`      | counter | Lifetime pages by interface                      |

## 🧑‍💻 Development

Run the current checkout directly:

```bash
go run ./cmd/epson_exporter --web.listen-address=:9788
```

Repository checks:

```bash
mise run deps
mise run build
mise run test
mise run lint
mise run fmt-check
```

Tests use captured printer pages and local servers; no printer is required.

## 📄 License

Licensed under the [Apache License 2.0](LICENSE).
