package exporter

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/woodleighschool/epson-exporter/internal/epson"
)

func (s *Server) Probe(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		moduleName = "default"
	}

	module, ok := s.Config.Modules[moduleName]
	if !ok {
		http.Error(w, fmt.Sprintf("unknown module %q", moduleName), http.StatusBadRequest)
		return
	}
	if target == "" {
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}

	client, err := epson.NewClient(target, epson.Paths{
		ProductStatus:  module.ProductStatusPath,
		UsageStatus:    module.UsageStatusPath,
		NetworkStatus:  module.NetworkStatusPath,
		HardwareStatus: module.HardwareStatusPath,
	}, module.HTTPClientConfig, s.Logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(&epson.Collector{
		Client:  client,
		Timeout: module.Timeout,
		Logger:  s.Logger,
	})

	promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
