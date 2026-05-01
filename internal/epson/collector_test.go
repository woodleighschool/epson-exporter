package epson

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestEmitSnapshot(t *testing.T) {
	registry := prometheus.NewPedanticRegistry()
	collector := staticCollector{snapshot: parseFixtureSnapshot(t)}
	registry.MustRegister(collector)

	expected := `
# HELP epson_consumable_level_percent Consumable level reported by the printer.
# TYPE epson_consumable_level_percent gauge
epson_consumable_level_percent{color="black",slot="BK",type="ink"} 1
epson_consumable_level_percent{color="cyan",slot="C",type="ink"} 12
epson_consumable_level_percent{color="maintenance_box",slot="maintenance_box",type="maintenance_box"} 49
epson_consumable_level_percent{color="magenta",slot="M",type="ink"} 17
epson_consumable_level_percent{color="yellow",slot="Y",type="ink"} 17
# HELP epson_consumable_warning Whether the printer marks the consumable with a warning icon.
# TYPE epson_consumable_warning gauge
epson_consumable_warning{color="black",slot="BK",type="ink"} 1
epson_consumable_warning{color="cyan",slot="C",type="ink"} 0
epson_consumable_warning{color="maintenance_box",slot="maintenance_box",type="maintenance_box"} 0
epson_consumable_warning{color="magenta",slot="M",type="ink"} 0
epson_consumable_warning{color="yellow",slot="Y",type="ink"} 0
# HELP epson_pages_total Lifetime pages reported by the printer.
# TYPE epson_pages_total counter
epson_pages_total{kind="black_and_white"} 3825
epson_pages_total{kind="color"} 6619
epson_pages_total{kind="duplex"} 8771
epson_pages_total{kind="simplex"} 1673
epson_pages_total{kind="total"} 10444
# HELP epson_status_info Human-readable status reported by a printer component.
# TYPE epson_status_info gauge
epson_status_info{component="hdd",status="Working normally."} 1
epson_status_info{component="printer",status="Available."} 1
epson_status_info{component="scanner",status="Available."} 1
epson_status_info{component="scanner",status="Working normally."} 1
epson_status_info{component="tpm",status="Working normally."} 1
`
	if err := testutil.GatherAndCompare(registry, strings.NewReader(expected),
		"epson_consumable_level_percent",
		"epson_consumable_warning",
		"epson_pages_total",
		"epson_status_info",
	); err != nil {
		t.Fatal(err)
	}
}

type staticCollector struct {
	snapshot Snapshot
}

func (c staticCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range allDescs {
		ch <- desc
	}
}

func (c staticCollector) Collect(ch chan<- prometheus.Metric) {
	emitSnapshot(ch, c.snapshot)
}
