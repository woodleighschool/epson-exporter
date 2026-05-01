package exporter

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"

	"github.com/woodleighschool/epson-exporter/internal/config"
)

type Server struct {
	Config      config.Config
	Logger      *slog.Logger
	MetricsPath string
}

func (s *Server) Register(mux *http.ServeMux) error {
	metricsPath := s.MetricsPath
	if metricsPath == "" {
		metricsPath = "/metrics"
	}

	mux.Handle(metricsPath, promhttp.Handler())
	mux.HandleFunc("/probe", s.Probe)

	landingPage, err := web.NewLandingPage(web.LandingConfig{
		Name:        "epson_exporter",
		Description: "Prometheus exporter for Epson printer status and usage pages",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{Address: metricsPath, Text: "Metrics"},
			{Address: "/probe", Text: "Probe"},
		},
	})
	if err != nil {
		return err
	}
	mux.Handle("/", landingPage)
	return nil
}
