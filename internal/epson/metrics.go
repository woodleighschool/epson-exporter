package epson

import "github.com/prometheus/client_golang/prometheus"

var (
	printerInfoDesc = prometheus.NewDesc(
		"epson_printer_info",
		"Epson printer information.",
		[]string{"model", "firmware", "mac_address", "device_name", "ip_address", "dns_host_name", "connection_status"},
		nil,
	)
	statusInfoDesc = prometheus.NewDesc(
		"epson_status_info",
		"Human-readable status reported by a printer component.",
		[]string{"component", "status"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		"epson_scrape_success",
		"Whether the Epson printer scrape completed successfully.",
		nil,
		nil,
	)
	scrapeDurationDesc = prometheus.NewDesc(
		"epson_scrape_duration_seconds",
		"Duration of the Epson printer scrape.",
		nil,
		nil,
	)
	consumableLevelDesc = prometheus.NewDesc(
		"epson_consumable_level_percent",
		"Consumable level reported by the printer.",
		[]string{"type", "slot", "color"},
		nil,
	)
	consumableInfoDesc = prometheus.NewDesc(
		"epson_consumable_info",
		"Consumable metadata reported by the printer.",
		[]string{"type", "slot", "color", "model"},
		nil,
	)
	consumableWarningDesc = prometheus.NewDesc(
		"epson_consumable_warning",
		"Whether the printer marks the consumable with a warning icon.",
		[]string{"type", "slot", "color"},
		nil,
	)
	paperSourceInfoDesc = prometheus.NewDesc(
		"epson_paper_source_info",
		"Paper source metadata reported by the printer.",
		[]string{"source", "paper_size", "paper_type", "level"},
		nil,
	)
	paperSourceRemainingDesc = prometheus.NewDesc(
		"epson_paper_source_remaining_ratio",
		"Approximate paper remaining ratio mapped from the printer's coarse level text.",
		[]string{"source"},
		nil,
	)
	firstPrintTimestampDesc = prometheus.NewDesc(
		"epson_first_print_timestamp_seconds",
		"Unix timestamp for the first printing date reported by the printer.",
		nil,
		nil,
	)
	pageTotalDesc = prometheus.NewDesc(
		"epson_pages_total",
		"Lifetime pages reported by the printer.",
		[]string{"kind"},
		nil,
	)
	pagesBySizeDesc = prometheus.NewDesc(
		"epson_pages_by_size_total",
		"Lifetime pages reported by size, side mode, and color mode.",
		[]string{"size", "sides", "color_mode"},
		nil,
	)
	pagesByFunctionDesc = prometheus.NewDesc(
		"epson_pages_by_function_total",
		"Lifetime pages reported by function and color mode.",
		[]string{"function", "color_mode"},
		nil,
	)
	pagesByLanguageDesc = prometheus.NewDesc(
		"epson_pages_by_language_total",
		"Lifetime pages reported by printer language.",
		[]string{"language"},
		nil,
	)
	pagesByInterfaceDesc = prometheus.NewDesc(
		"epson_pages_by_interface_total",
		"Lifetime pages reported by interface and color mode.",
		[]string{"interface", "color_mode"},
		nil,
	)
)

var allDescs = []*prometheus.Desc{
	printerInfoDesc,
	statusInfoDesc,
	scrapeSuccessDesc,
	scrapeDurationDesc,
	consumableLevelDesc,
	consumableInfoDesc,
	consumableWarningDesc,
	paperSourceInfoDesc,
	paperSourceRemainingDesc,
	firstPrintTimestampDesc,
	pageTotalDesc,
	pagesBySizeDesc,
	pagesByFunctionDesc,
	pagesByLanguageDesc,
	pagesByInterfaceDesc,
}
