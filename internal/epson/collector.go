package epson

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	Client  *Client
	Timeout time.Duration
	Logger  *slog.Logger
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range allDescs {
		ch <- desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	snapshot, err := c.Client.Scrape(ctx)
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
	if err != nil {
		if c.Logger != nil {
			c.Logger.Warn("epson scrape failed", "err", err)
		}
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, 1)
	emitSnapshot(ch, snapshot)
}

func emitSnapshot(ch chan<- prometheus.Metric, snapshot Snapshot) {
	ch <- prometheus.MustNewConstMetric(
		printerInfoDesc,
		prometheus.GaugeValue,
		1,
		snapshot.Model,
		snapshot.Firmware,
		snapshot.MacAddress,
		snapshot.DeviceName,
		snapshot.IPAddress,
		snapshot.DNSHostName,
		snapshot.ConnectionStatus,
	)
	emitStatus(ch, "printer", snapshot.PrinterStatus)
	emitStatus(ch, "scanner", snapshot.ScannerStatus)
	for _, status := range snapshot.HardwareStatuses {
		emitStatus(ch, status.Component, status.Status)
	}

	for _, consumable := range snapshot.Consumables {
		ch <- prometheus.MustNewConstMetric(
			consumableLevelDesc,
			prometheus.GaugeValue,
			consumable.LevelPercent,
			consumable.Type,
			consumable.Slot,
			consumable.Color,
		)
		ch <- prometheus.MustNewConstMetric(
			consumableInfoDesc,
			prometheus.GaugeValue,
			1,
			consumable.Type,
			consumable.Slot,
			consumable.Color,
			consumable.Model,
		)
		ch <- prometheus.MustNewConstMetric(
			consumableWarningDesc,
			prometheus.GaugeValue,
			boolFloat(consumable.Warning),
			consumable.Type,
			consumable.Slot,
			consumable.Color,
		)
	}

	for _, source := range snapshot.PaperSources {
		ch <- prometheus.MustNewConstMetric(
			paperSourceInfoDesc,
			prometheus.GaugeValue,
			1,
			source.Name,
			source.Size,
			source.Type,
			source.Level,
		)
		if source.HasLevel {
			ch <- prometheus.MustNewConstMetric(
				paperSourceRemainingDesc,
				prometheus.GaugeValue,
				source.LevelFraction,
				source.Name,
			)
		}
	}

	if snapshot.FirstPrintUnixTime > 0 {
		ch <- prometheus.MustNewConstMetric(firstPrintTimestampDesc, prometheus.GaugeValue, snapshot.FirstPrintUnixTime)
	}
	for _, total := range snapshot.PageTotals {
		ch <- prometheus.MustNewConstMetric(pageTotalDesc, prometheus.CounterValue, total.Value, total.Kind)
	}
	for _, pages := range snapshot.PagesBySize {
		ch <- prometheus.MustNewConstMetric(pagesBySizeDesc, prometheus.CounterValue, pages.Value, pages.Size, pages.Sides, pages.ColorMode)
	}
	for _, pages := range snapshot.PagesByFunction {
		ch <- prometheus.MustNewConstMetric(pagesByFunctionDesc, prometheus.CounterValue, pages.Value, pages.Function, pages.ColorMode)
	}
	for _, pages := range snapshot.PagesByLanguage {
		ch <- prometheus.MustNewConstMetric(pagesByLanguageDesc, prometheus.CounterValue, pages.Value, pages.Language)
	}
	for _, pages := range snapshot.PagesByInterface {
		ch <- prometheus.MustNewConstMetric(pagesByInterfaceDesc, prometheus.CounterValue, pages.Value, pages.Interface, pages.ColorMode)
	}
}

func emitStatus(ch chan<- prometheus.Metric, component, status string) {
	if component == "" || status == "" {
		return
	}
	ch <- prometheus.MustNewConstMetric(statusInfoDesc, prometheus.GaugeValue, 1, component, status)
}

func boolFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
