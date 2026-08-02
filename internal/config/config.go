package config

import (
	"errors"
	"os"
	"strings"
	"time"

	promconfig "github.com/prometheus/common/config"
	"go.yaml.in/yaml/v3"
)

const (
	defaultProductStatusPath  = "/PRESENTATION/ADVANCED/INFO_PRTINFO/TOP"
	defaultUsageStatusPath    = "/PRESENTATION/ADVANCED/INFO_MENTINFO/TOP"
	defaultNetworkStatusPath  = "/PRESENTATION/ADVANCED/INFO_NWINFO/TOP"
	defaultHardwareStatusPath = "/PRESENTATION/ADVANCED/INFO_BEHAVIORINFO/TOP"
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Timeout            time.Duration               `yaml:"timeout"`
	ProductStatusPath  string                      `yaml:"product_status_path"`
	UsageStatusPath    string                      `yaml:"usage_status_path"`
	NetworkStatusPath  string                      `yaml:"network_status_path"`
	HardwareStatusPath string                      `yaml:"hardware_status_path"`
	HTTPClientConfig   promconfig.HTTPClientConfig `yaml:"http_client_config"`
}

func (m *Module) UnmarshalYAML(unmarshal func(any) error) error {
	type plain Module
	*m = Module{
		HTTPClientConfig: defaultHTTPClientConfig(),
	}
	return unmarshal((*plain)(m))
}

func LoadFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	return Load(data)
}

func Default() Config {
	cfg := Config{
		Modules: map[string]Module{
			"default": {
				Timeout:            5 * time.Second,
				ProductStatusPath:  defaultProductStatusPath,
				UsageStatusPath:    defaultUsageStatusPath,
				NetworkStatusPath:  defaultNetworkStatusPath,
				HardwareStatusPath: defaultHardwareStatusPath,
				HTTPClientConfig:   defaultHTTPClientConfig(),
			},
		},
	}
	return cfg
}

func Load(data []byte) (Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if len(cfg.Modules) == 0 {
		return Config{}, errors.New("config must define at least one module")
	}

	return normalize(cfg)
}

func normalize(cfg Config) (Config, error) {
	for name, module := range cfg.Modules {
		if module.Timeout == 0 {
			module.Timeout = 5 * time.Second
		}
		if module.ProductStatusPath == "" {
			module.ProductStatusPath = defaultProductStatusPath
		}
		if module.UsageStatusPath == "" {
			module.UsageStatusPath = defaultUsageStatusPath
		}
		if module.NetworkStatusPath == "" {
			module.NetworkStatusPath = defaultNetworkStatusPath
		}
		if module.HardwareStatusPath == "" {
			module.HardwareStatusPath = defaultHardwareStatusPath
		}
		module.ProductStatusPath = normalizePath(module.ProductStatusPath)
		module.UsageStatusPath = normalizePath(module.UsageStatusPath)
		module.NetworkStatusPath = normalizePath(module.NetworkStatusPath)
		module.HardwareStatusPath = normalizePath(module.HardwareStatusPath)
		if err := module.HTTPClientConfig.Validate(); err != nil {
			return Config{}, err
		}
		cfg.Modules[name] = module
	}

	return cfg, nil
}

func normalizePath(value string) string {
	return "/" + strings.Trim(strings.TrimSpace(value), "/")
}

func defaultHTTPClientConfig() promconfig.HTTPClientConfig {
	httpConfig := promconfig.DefaultHTTPClientConfig
	// Epson printer web UIs commonly force HTTPS with locally-issued certificates.
	// Keep this default practical for printer scraping; users can override it in a
	// module when they have a trusted CA or want stricter validation.
	httpConfig.TLSConfig.InsecureSkipVerify = true
	return httpConfig
}
