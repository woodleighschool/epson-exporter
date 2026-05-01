package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/woodleighschool/epson-exporter/internal/config"
	"github.com/woodleighschool/epson-exporter/internal/exporter"
)

var (
	configFile   = kingpin.Flag("config.file", "Optional Epson exporter configuration file. If unset, the built-in default module is used.").String()
	configCheck  = kingpin.Flag("config.check", "Validate the config file and exit.").Default("false").Bool()
	metricsPath  = kingpin.Flag("web.telemetry-path", "Path under which to expose exporter metrics.").Default("/metrics").String()
	toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9788")
)

func main() {
	os.Exit(run())
}

func run() int {
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("epson_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promslogConfig)
	logger.Info("Starting epson_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	cfg := config.Default()
	if *configFile != "" {
		var err error
		cfg, err = config.LoadFile(*configFile)
		if err != nil {
			logger.Error("error loading config", "err", err)
			return 1
		}
	}
	if configJSON, err := json.Marshal(cfg); err == nil {
		logger.Info("Loaded config", "file", *configFile, "config", string(configJSON))
	}
	if *configCheck {
		return 0
	}

	prometheus.MustRegister(versioncollector.NewCollector("epson_exporter"))

	mux := http.NewServeMux()
	server := &exporter.Server{
		Config:      cfg,
		Logger:      logger,
		MetricsPath: *metricsPath,
	}
	if err := server.Register(mux); err != nil {
		logger.Error("error creating HTTP handlers", "err", err)
		return 1
	}

	httpServer := &http.Server{
		Handler: logRequests(mux, logger),
	}
	if err := web.ListenAndServe(httpServer, toolkitFlags, logger); err != nil {
		logger.Error("error running HTTP server", "err", err)
		return 1
	}
	return 0
}

func logRequests(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
