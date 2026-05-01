package epson

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSnapshotFromAnonymizedWorkForceFixture(t *testing.T) {
	snapshot := parseFixtureSnapshot(t)

	if snapshot.Model != "AM-C6000 Series" {
		t.Fatalf("Model = %q", snapshot.Model)
	}
	if snapshot.Firmware != "07.84.GW19Q3" {
		t.Fatalf("Firmware = %q", snapshot.Firmware)
	}
	if snapshot.MacAddress != "00:00:5E:00:53:30" {
		t.Fatalf("MacAddress = %q", snapshot.MacAddress)
	}
	if snapshot.PrinterStatus != "Available." || snapshot.ScannerStatus != "Available." {
		t.Fatalf("statuses = printer %q scanner %q", snapshot.PrinterStatus, snapshot.ScannerStatus)
	}
	if snapshot.FirstPrintUnixTime == 0 {
		t.Fatal("FirstPrintUnixTime was not parsed")
	}

	assertConsumable(t, snapshot, "ink", "BK", "black", "T08E1", 1, true)
	assertConsumable(t, snapshot, "ink", "Y", "yellow", "T08E4", 17, false)
	assertConsumable(t, snapshot, "maintenance_box", "maintenance_box", "maintenance_box", "C9371", 49, false)
	assertPageTotal(t, snapshot, "total", 10444)
	assertPageTotal(t, snapshot, "black_and_white", 3825)
	assertPageTotal(t, snapshot, "color", 6619)
	assertPagesByFunction(t, snapshot, "print_from_computer_or_mobile_device", "color", 4571)
	assertPagesByLanguage(t, snapshot, "pcl", 1465)
	assertPagesByInterface(t, snapshot, "standard_network", "color", 4571)
}

func parseFixtureSnapshot(t *testing.T) Snapshot {
	t.Helper()
	snapshot, err := ParseSnapshot(
		readTestFile(t, "product_status.html"),
		readTestFile(t, "usage_status.html"),
		readTestFile(t, "network_status.html"),
		readTestFile(t, "hardware_status.html"),
	)
	if err != nil {
		t.Fatalf("ParseSnapshot returned error: %v", err)
	}
	return snapshot
}

func assertConsumable(t *testing.T, snapshot Snapshot, typ, slot, color, model string, level float64, warning bool) {
	t.Helper()
	for _, consumable := range snapshot.Consumables {
		if consumable.Type == typ && consumable.Slot == slot && consumable.Color == color {
			if consumable.Model != model || consumable.LevelPercent != level || consumable.Warning != warning {
				t.Fatalf("consumable %+v did not match model=%q level=%v warning=%v", consumable, model, level, warning)
			}
			return
		}
	}
	t.Fatalf("missing consumable type=%q slot=%q color=%q", typ, slot, color)
}

func assertPageTotal(t *testing.T, snapshot Snapshot, kind string, value float64) {
	t.Helper()
	for _, total := range snapshot.PageTotals {
		if total.Kind == kind {
			if total.Value != value {
				t.Fatalf("page total %q = %v, want %v", kind, total.Value, value)
			}
			return
		}
	}
	t.Fatalf("missing page total %q", kind)
}

func assertPagesByFunction(t *testing.T, snapshot Snapshot, function, colorMode string, value float64) {
	t.Helper()
	for _, pages := range snapshot.PagesByFunction {
		if pages.Function == function && pages.ColorMode == colorMode {
			if pages.Value != value {
				t.Fatalf("pages by function %+v want %v", pages, value)
			}
			return
		}
	}
	t.Fatalf("missing pages by function %q/%q", function, colorMode)
}

func assertPagesByLanguage(t *testing.T, snapshot Snapshot, language string, value float64) {
	t.Helper()
	for _, pages := range snapshot.PagesByLanguage {
		if pages.Language == language {
			if pages.Value != value {
				t.Fatalf("pages by language %+v want %v", pages, value)
			}
			return
		}
	}
	t.Fatalf("missing pages by language %q", language)
}

func assertPagesByInterface(t *testing.T, snapshot Snapshot, iface, colorMode string, value float64) {
	t.Helper()
	for _, pages := range snapshot.PagesByInterface {
		if pages.Interface == iface && pages.ColorMode == colorMode {
			if pages.Value != value {
				t.Fatalf("pages by interface %+v want %v", pages, value)
			}
			return
		}
	}
	t.Fatalf("missing pages by interface %q/%q", iface, colorMode)
}

func readTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
