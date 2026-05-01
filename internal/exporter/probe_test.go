package exporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	promconfig "github.com/prometheus/common/config"

	"github.com/woodleighschool/epson-exporter/internal/config"
)

func TestProbeRequiresTarget(t *testing.T) {
	server := testServer(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()
	server.Probe(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestProbeUnknownModule(t *testing.T) {
	server := testServer(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/probe?target=http://example.test&module=missing", nil)
	rec := httptest.NewRecorder()
	server.Probe(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestProbeHealthy(t *testing.T) {
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveProbeFixture(t, w, r.URL.Path)
	}))
	defer device.Close()

	server := testServer(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(device.URL), nil)
	rec := httptest.NewRecorder()
	server.Probe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		"epson_scrape_success 1",
		"epson_printer_info",
		`epson_pages_total{kind="total"} 10444`,
		`epson_consumable_level_percent{color="black",slot="BK",type="ink"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("probe body missing %q:\n%s", want, body)
		}
	}
}

func TestProbeUpstreamFailureIsMetricFailure(t *testing.T) {
	device := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "offline", http.StatusBadGateway)
	}))
	defer device.Close()

	server := testServer(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/probe?target="+url.QueryEscape(device.URL), nil)
	rec := httptest.NewRecorder()
	server.Probe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "epson_scrape_success 0") {
		t.Fatalf("probe body missing failure metric:\n%s", rec.Body.String())
	}
}

func testServer(t *testing.T, cfg *config.Config) *Server {
	t.Helper()
	if cfg == nil {
		cfg = &config.Config{
			Modules: map[string]config.Module{
				"default": {
					Timeout:            5 * time.Second,
					ProductStatusPath:  "/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP",
					UsageStatusPath:    "/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP",
					NetworkStatusPath:  "/PRESENTATION/ADVANCED/INFO_NWINFO/TOP",
					HardwareStatusPath: "/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP",
					HTTPClientConfig:   promconfig.DefaultHTTPClientConfig,
				},
			},
		}
	}
	return &Server{Config: *cfg, MetricsPath: "/metrics"}
}

func serveProbeFixture(t *testing.T, w http.ResponseWriter, reqPath string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/html")
	switch reqPath {
	case "/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP":
		writeProbeFixture(t, w, "../epson/testdata/product_status.html")
	case "/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP":
		writeProbeFixture(t, w, "../epson/testdata/usage_status.html")
	case "/PRESENTATION/ADVANCED/INFO_NWINFO/TOP":
		writeProbeFixture(t, w, "../epson/testdata/network_status.html")
	case "/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP":
		writeProbeFixture(t, w, "../epson/testdata/hardware_status.html")
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func writeProbeFixture(t *testing.T, w http.ResponseWriter, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(data)
}
