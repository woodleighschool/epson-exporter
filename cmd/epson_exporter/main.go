package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/bootstrap"
	"github.com/prometheus/exporter-toolkit/web"

	"github.com/woodleighschool/epson-exporter/internal/config"
	"github.com/woodleighschool/epson-exporter/internal/exporter"
)

var (
	errConfigCheck = errors.New("configuration check complete")
	configFile     = kingpin.Flag("config.file", "Optional configuration file. If unset, the built-in default module is used.").String()
	configCheck    = kingpin.Flag("config.check", "Validate the config file and exit.").Default("false").Bool()
)

func main() {
	runner := bootstrap.New(bootstrap.Config{
		App:            kingpin.CommandLine,
		Name:           "epson_exporter",
		Description:    "Prometheus exporter for Epson printer status and usage pages",
		DefaultAddress: ":9788",
		LandingConfig: web.LandingConfig{
			Links: []web.LandingLinks{{Address: "/probe", Text: "Probe"}},
		},
		MetricsHandlerFactory: newMetricsHandler,
	})
	if err := runner.Run(); err != nil && !errors.Is(err, errConfigCheck) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newMetricsHandler(b *bootstrap.Bootstrap) (http.Handler, error) {
	cfg := config.Default()
	if *configFile != "" {
		var err error
		cfg, err = config.LoadFile(*configFile)
		if err != nil {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}
	if configJSON, err := json.Marshal(cfg); err == nil {
		b.Logger.Info("Loaded config", "file", *configFile, "config", string(configJSON))
	}
	if *configCheck {
		return nil, errConfigCheck
	}

	b.Handle("/probe", exporter.NewProbeHandler(cfg, b.Logger))

	registry := prometheus.NewRegistry()
	registry.MustRegister(versioncollector.NewCollector("epson_exporter"))
	if !b.DisableExporterMetrics {
		registry.MustRegister(
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
	}

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		MaxRequestsInFlight: b.MaxRequests,
	})
	if !b.DisableExporterMetrics {
		handler = promhttp.InstrumentMetricHandler(registry, handler)
	}
	return handler, nil
}
