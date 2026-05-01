package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load([]byte(`
modules:
  default: {}
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	module := cfg.Modules["default"]
	if module.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s, want 5s", module.Timeout)
	}
	if module.ProductStatusPath != defaultProductStatusPath {
		t.Fatalf("ProductStatusPath = %q, want default", module.ProductStatusPath)
	}
	if !module.HTTPClientConfig.TLSConfig.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify = false, want practical printer default")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	module, ok := cfg.Modules["default"]
	if !ok {
		t.Fatal("default module missing")
	}
	if module.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %s, want 5s", module.Timeout)
	}
	if module.ProductStatusPath != defaultProductStatusPath {
		t.Fatalf("ProductStatusPath = %q, want default", module.ProductStatusPath)
	}
	if !module.HTTPClientConfig.TLSConfig.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify = false, want practical printer default")
	}
}

func TestLoadTrimsPaths(t *testing.T) {
	cfg, err := Load([]byte(`
modules:
  default:
    timeout: 7s
    product_status_path: PRESENTATION/ADVANCED/INFO_PRTINFO/TOP/
    usage_status_path: /PRESENTATION/ADVANCED/INFO_MENTINFO/TOP
`))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	module := cfg.Modules["default"]
	if module.Timeout != 7*time.Second {
		t.Fatalf("Timeout = %s, want 7s", module.Timeout)
	}
	if module.ProductStatusPath != defaultProductStatusPath {
		t.Fatalf("ProductStatusPath = %q, want normalized default path", module.ProductStatusPath)
	}
	if module.UsageStatusPath != defaultUsageStatusPath {
		t.Fatalf("UsageStatusPath = %q, want normalized default path", module.UsageStatusPath)
	}
}

func TestLoadRequiresModule(t *testing.T) {
	if _, err := Load([]byte(`modules: {}`)); err == nil {
		t.Fatal("Load succeeded with no modules")
	}
}

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(`
modules:
  default:
    timeout: 3s
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}
	if cfg.Modules["default"].Timeout != 3*time.Second {
		t.Fatalf("Timeout = %s, want 3s", cfg.Modules["default"].Timeout)
	}
}
