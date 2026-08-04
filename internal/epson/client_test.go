package epson

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	promconfig "github.com/prometheus/common/config"
)

func TestClientScrape(t *testing.T) {
	seen := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.Path] = true
		if got := r.Header.Get("User-Agent"); got != userAgent {
			t.Fatalf("User-Agent = %q, want %q", got, userAgent)
		}
		serveFixture(t, w, r.URL.Path)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	snapshot, err := client.Scrape(context.Background())
	if err != nil {
		t.Fatalf("Scrape returned error: %v", err)
	}

	for _, path := range []string{
		"/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP",
		"/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP",
		"/PRESENTATION/ADVANCED/INFO_NWINFO/TOP",
		"/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP",
	} {
		if !seen[path] {
			t.Fatalf("expected request for %s", path)
		}
	}
	if snapshot.Model != "AM-C6000 Series" {
		t.Fatalf("Model = %q, want AM-C6000 Series", snapshot.Model)
	}
	if snapshot.DeviceName != "ANON-PRINTER" {
		t.Fatalf("DeviceName = %q, want anonymized fixture name", snapshot.DeviceName)
	}
	if len(snapshot.PageTotals) == 0 {
		t.Fatal("expected page totals")
	}
}

func TestClientNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	if _, err := client.Scrape(context.Background()); err == nil {
		t.Fatal("Scrape succeeded after non-2xx response")
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := client.Scrape(ctx); err == nil {
		t.Fatal("Scrape succeeded after context timeout")
	}
}

func TestNewClientRequiresHTTPURL(t *testing.T) {
	if _, err := NewClient("printer.local", defaultPaths(), promconfig.DefaultHTTPClientConfig, nil); err == nil {
		t.Fatal("NewClient accepted target without scheme")
	}
}

func newTestClient(t *testing.T, target string) *Client {
	t.Helper()
	client, err := NewClient(target, defaultPaths(), promconfig.DefaultHTTPClientConfig, nil)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	return client
}

func defaultPaths() Paths {
	return Paths{
		ProductStatus:  "/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP",
		UsageStatus:    "/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP",
		NetworkStatus:  "/PRESENTATION/ADVANCED/INFO_NWINFO/TOP",
		HardwareStatus: "/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP",
	}
}

func serveFixture(t *testing.T, w http.ResponseWriter, reqPath string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/html")
	switch reqPath {
	case "/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP":
		writeFixture(t, w, "product_status.html")
	case "/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP":
		writeFixture(t, w, "usage_status.html")
	case "/PRESENTATION/ADVANCED/INFO_NWINFO/TOP":
		writeFixture(t, w, "network_status.html")
	case "/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP":
		writeFixture(t, w, "hardware_status.html")
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func writeFixture(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(data)
}
