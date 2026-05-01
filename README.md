# epson-exporter

Prometheus exporter for Epson printer embedded web pages, aimed at WorkForce / WF-style devices. `/metrics` exposes exporter self-metrics, and `/probe` scrapes one printer.

It reads the status and usage HTML pages Epson exposes without login, then emits printer status, consumable, paper, and usage metrics.

## Run locally

```sh
go run ./cmd/epson_exporter --web.listen-address=:9788
```

Example probe:

```sh
curl 'http://127.0.0.1:9788/probe?target=https://192.0.2.10&module=default'
```

The built-in `default` module uses these Epson paths and a 5 second timeout:

- `/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP`
- `/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP`
- `/PRESENTATION/ADVANCED/INFO_NWINFO/TOP`
- `/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP`

Epson printers commonly redirect HTTP to HTTPS and serve a locally-issued certificate. The default module skips target certificate verification for that reason. Use a config file only if a printer has different paths, timeout needs, or stricter TLS requirements.

## Prometheus

```yaml
scrape_configs:
  - job_name: epson
    metrics_path: /probe
    params:
      module: [default]
    static_configs:
      - targets: ["https://192.0.2.10"]
        labels:
          site: example
          printer: library
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [printer]
        target_label: instance
      - target_label: __address__
        replacement: epson-exporter:9788
```

## Metrics

| Metric                                | Type    | Labels                                                                                                | Description                                                                    |
| ------------------------------------- | ------- | ----------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `epson_printer_info`                  | gauge   | `model`, `firmware`, `mac_address`, `device_name`, `ip_address`, `dns_host_name`, `connection_status` | Printer metadata.                                                              |
| `epson_status_info`                   | gauge   | `component`, `status`                                                                                 | Human-readable printer, scanner, and hardware status.                          |
| `epson_scrape_success`                | gauge   |                                                                                                       | `1` when the printer scrape succeeds, otherwise `0`.                           |
| `epson_scrape_duration_seconds`       | gauge   |                                                                                                       | Printer scrape duration.                                                       |
| `epson_consumable_level_percent`      | gauge   | `type`, `slot`, `color`                                                                               | Ink or maintenance box level.                                                  |
| `epson_consumable_info`               | gauge   | `type`, `slot`, `color`, `model`                                                                      | Consumable metadata.                                                           |
| `epson_consumable_warning`            | gauge   | `type`, `slot`, `color`                                                                               | `1` when the printer marks the consumable with a warning icon.                 |
| `epson_paper_source_info`             | gauge   | `source`, `paper_size`, `paper_type`, `level`                                                         | Paper source metadata.                                                         |
| `epson_paper_source_remaining_ratio`  | gauge   | `source`                                                                                              | Approximate paper remaining ratio mapped from the printer's coarse level text. |
| `epson_first_print_timestamp_seconds` | gauge   |                                                                                                       | Unix timestamp for the first printing date reported by the printer.            |
| `epson_pages_total`                   | counter | `kind`                                                                                                | Lifetime pages by total kind.                                                  |
| `epson_pages_by_size_total`           | counter | `size`, `sides`, `color_mode`                                                                         | Lifetime pages by size, side mode, and color mode.                             |
| `epson_pages_by_function_total`       | counter | `function`, `color_mode`                                                                              | Lifetime pages by function and color mode.                                     |
| `epson_pages_by_language_total`       | counter | `language`                                                                                            | Lifetime pages by printer language.                                            |
| `epson_pages_by_interface_total`      | counter | `interface`, `color_mode`                                                                             | Lifetime pages by interface and color mode.                                    |
