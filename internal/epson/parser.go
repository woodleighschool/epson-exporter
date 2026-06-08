package epson

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var gradientPercentRE = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)%`)

func ParseSnapshot(productStatus, usageStatus, networkStatus, hardwareStatus []byte) (Snapshot, error) {
	productDoc, err := parseHTML(productStatus)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse product status: %w", err)
	}
	usageDoc, err := parseHTML(usageStatus)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse usage status: %w", err)
	}
	networkDoc, err := parseHTML(networkStatus)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse network status: %w", err)
	}
	hardwareDoc, err := parseHTML(hardwareStatus)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse hardware status: %w", err)
	}

	productValues := definitionValues(productDoc)
	usageValues := definitionValues(usageDoc)
	networkValues := definitionValues(networkDoc)

	snapshot := Snapshot{
		Model:            firstNonEmpty(pageTitle(productDoc), pageTitle(usageDoc), pageTitle(networkDoc)),
		Firmware:         productValues["Firmware"],
		MacAddress:       firstNonEmpty(productValues["Network MAC Address"], networkValues["MAC Address"]),
		DeviceName:       networkValues["Device Name"],
		ConnectionStatus: networkValues["Connection Status"],
		IPAddress:        networkValues["IP Address"],
		DNSHostName:      networkValues["DNS Host Name"],
		FirstPrintDate:   usageValues["First Printing Date"],
		Consumables:      parseConsumables(productDoc, productValues),
		PaperSources:     parsePaperSources(productDoc),
		PageTotals:       parsePageTotals(usageValues),
		PagesBySize:      parsePagesBySize(usageDoc),
		PagesByFunction:  parsePagesByFunction(usageDoc),
		PagesByLanguage:  parsePagesByLanguage(usageDoc),
		PagesByInterface: parsePagesByInterface(usageDoc),
		HardwareStatuses: parseHardwareStatuses(hardwareDoc),
	}
	if ts, ok := parsePrinterDate(snapshot.FirstPrintDate); ok {
		snapshot.FirstPrintUnixTime = float64(ts)
	}
	statuses := parseStatusFieldsets(productDoc)
	snapshot.PrinterStatus = statuses["Printer Status"]
	snapshot.ScannerStatus = statuses["Scanner Status"]

	return snapshot, nil
}

func parseHTML(data []byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(data))
}

func pageTitle(root *html.Node) string {
	for node := range nodes(root) {
		if node.Type == html.ElementNode && node.Data == "title" {
			return cleanText(textContent(node))
		}
	}
	return ""
}

func definitionValues(root *html.Node) map[string]string {
	values := map[string]string{}
	for _, kv := range definitionPairs(root) {
		values[kv.key] = kv.value
	}
	return values
}

type definitionPair struct {
	key   string
	value string
}

func definitionPairs(root *html.Node) []definitionPair {
	var pairs []definitionPair
	for node := range nodes(root) {
		if node.Type != html.ElementNode || node.Data != "dt" {
			continue
		}
		valueNode := nextElementSibling(node)
		if valueNode == nil || valueNode.Data != "dd" {
			continue
		}
		key := cleanKey(textContent(node))
		value := cleanText(textContent(valueNode))
		if key != "" {
			pairs = append(pairs, definitionPair{key: key, value: value})
		}
	}
	return pairs
}

func parseStatusFieldsets(root *html.Node) map[string]string {
	statuses := map[string]string{}
	for _, fieldset := range elements(root, "fieldset") {
		legend := cleanText(textContent(firstChildElement(fieldset, "legend")))
		if legend == "" {
			continue
		}
		// Epson's status sections do not use dt/dd pairs; the first preserve-white-space
		// div below each legend is the actual human status text.
		for node := range nodes(fieldset) {
			if node.Type == html.ElementNode && node.Data == "div" && hasClass(node, "preserve-white-space") {
				statuses[legend] = cleanText(textContent(node))
				break
			}
		}
	}
	return statuses
}

func parseConsumables(root *html.Node, values map[string]string) []Consumable {
	var consumables []Consumable
	for _, item := range elements(root, "li") {
		if !hasClass(item, "tank") {
			continue
		}
		slot := cleanText(textContent(firstDescendantWithClass(item, "clrname")))
		consumableType := "ink"
		color := colorForSlot(slot)
		if slot == "" && descendantImageContains(item, "Icn_Mb") {
			slot = "maintenance_box"
			consumableType = "maintenance_box"
			color = "maintenance_box"
		}
		if slot == "" {
			continue
		}

		warning := firstDescendantWithClass(item, "inkst") != nil
		emptyInk := descendantImageContains(item, "Icn_No")
		level, ok := tankLevel(item, emptyInk)
		if !ok {
			continue
		}
		consumables = append(consumables, Consumable{
			Type:         consumableType,
			Slot:         slot,
			Color:        color,
			Model:        consumableModel(slot, values),
			LevelPercent: level,
			Warning:      warning,
		})
	}
	return consumables
}

func tankLevel(item *html.Node, empty bool) (float64, bool) {
	for node := range nodes(item) {
		if node.Type != html.ElementNode || node.Data != "div" || !hasClass(node, "tank") {
			continue
		}
		matches := gradientPercentRE.FindAllStringSubmatch(attr(node, "style"), -1)
		if len(matches) < 2 {
			return 0, false
		}
		level, err := strconv.ParseFloat(matches[1][1], 64)
		if empty && err == nil && level > 0 {
			return 0, true
		}
		return level, err == nil
	}
	return 0, false
}

func consumableModel(slot string, values map[string]string) string {
	switch slot {
	case "BK":
		return values["Black (BK)"]
	case "Y":
		return values["Yellow (Y)"]
	case "M":
		return values["Magenta (M)"]
	case "C":
		return values["Cyan (C)"]
	case "maintenance_box":
		return values["Maintenance Box"]
	default:
		return ""
	}
}

func colorForSlot(slot string) string {
	switch slot {
	case "BK":
		return "black"
	case "Y":
		return "yellow"
	case "M":
		return "magenta"
	case "C":
		return "cyan"
	default:
		return strings.ToLower(slot)
	}
}

func parsePaperSources(root *html.Node) []PaperSource {
	var sources []PaperSource
	for _, fieldset := range elements(root, "fieldset") {
		name := cleanText(textContent(firstChildElement(fieldset, "legend")))
		if !strings.HasPrefix(name, "Cassette ") && name != "Paper Tray" {
			continue
		}
		values := definitionValues(fieldset)
		source := PaperSource{
			Name:  labelValue(name),
			Size:  values["Paper Size"],
			Type:  values["Paper Type"],
			Level: values["Paper Remaining Level"],
		}
		if fraction, ok := paperLevelFraction(source.Level); ok {
			source.LevelFraction = fraction
			source.HasLevel = true
		}
		sources = append(sources, source)
	}
	return sources
}

func paperLevelFraction(level string) (float64, bool) {
	switch labelValue(level) {
	case "full":
		return 1, true
	case "high":
		return 0.75, true
	case "middle", "medium", "half":
		return 0.5, true
	case "low":
		return 0.25, true
	case "empty", "none":
		return 0, true
	default:
		return 0, false
	}
}

func parsePageTotals(values map[string]string) []PageTotal {
	keyKinds := map[string]string{
		"Total Number of Pages":                  "total",
		"Total Number of B&W Pages":              "black_and_white",
		"Total Number of Color Pages":            "color",
		"Total Number of 2-Sided Printing Pages": "duplex",
		"Total Number of 1-Sided Printing Pages": "simplex",
	}

	var totals []PageTotal
	for key, kind := range keyKinds {
		if value, ok := parseCounter(values[key]); ok {
			totals = append(totals, PageTotal{Kind: kind, Value: value})
		}
	}
	return totals
}

func parsePagesBySize(root *html.Node) []PagesBySize {
	table := tableInFieldset(root, "Number of Pages Sorted by Size")
	var pages []PagesBySize
	for _, row := range tableRows(table) {
		if len(row) != 5 || row[0] == "" || row[0] == "1-sided" {
			continue
		}
		columns := []struct {
			sides     string
			colorMode string
			value     string
		}{
			{sides: "simplex", colorMode: "black_and_white", value: row[1]},
			{sides: "simplex", colorMode: "color", value: row[2]},
			{sides: "duplex", colorMode: "black_and_white", value: row[3]},
			{sides: "duplex", colorMode: "color", value: row[4]},
		}
		for _, column := range columns {
			if value, ok := parseCounter(column.value); ok {
				pages = append(pages, PagesBySize{
					Size:      labelValue(row[0]),
					Sides:     column.sides,
					ColorMode: column.colorMode,
					Value:     value,
				})
			}
		}
	}
	return pages
}

func parsePagesByFunction(root *html.Node) []PagesByFunction {
	fieldset := fieldsetByLegend(root, "Total Number of Pages Sorted by Function")
	values := definitionValues(fieldset)
	var pages []PagesByFunction
	for key, raw := range values {
		value, ok := parseCounter(raw)
		if !ok {
			continue
		}
		colorMode, function := splitColorPrefix(key)
		pages = append(pages, PagesByFunction{
			Function:  labelValue(function),
			ColorMode: colorMode,
			Value:     value,
		})
	}
	return pages
}

func parsePagesByLanguage(root *html.Node) []PagesByLanguage {
	fieldset := fieldsetByLegend(root, "Total Number of Pages Sorted by Print Language")
	values := definitionValues(fieldset)
	var pages []PagesByLanguage
	for language, raw := range values {
		if value, ok := parseCounter(raw); ok {
			pages = append(pages, PagesByLanguage{
				Language: labelValue(language),
				Value:    value,
			})
		}
	}
	return pages
}

func parsePagesByInterface(root *html.Node) []PagesByInterface {
	table := tableInFieldset(root, "Total Number of Pages Sorted by Interface")
	var pages []PagesByInterface
	for _, row := range tableRows(table) {
		if len(row) != 4 || row[0] == "" || row[0] == "Standard Network" {
			continue
		}
		colorMode := colorModeLabel(row[0])
		for idx, iface := range []string{"standard_network", "additional_network", "other"} {
			if value, ok := parseCounter(row[idx+1]); ok {
				pages = append(pages, PagesByInterface{
					Interface: iface,
					ColorMode: colorMode,
					Value:     value,
				})
			}
		}
	}
	return pages
}

func parseHardwareStatuses(root *html.Node) []ComponentStatus {
	values := definitionValues(root)
	var statuses []ComponentStatus
	for component, status := range values {
		statuses = append(statuses, ComponentStatus{
			Component: labelValue(component),
			Status:    status,
		})
	}
	return statuses
}

func splitColorPrefix(value string) (string, string) {
	switch {
	case strings.HasPrefix(value, "B&W "):
		return "black_and_white", strings.TrimPrefix(value, "B&W ")
	case strings.HasPrefix(value, "Color "):
		return "color", strings.TrimPrefix(value, "Color ")
	default:
		return "unknown", value
	}
}

func colorModeLabel(value string) string {
	switch value {
	case "Black and White", "B&W":
		return "black_and_white"
	case "Color":
		return "color"
	default:
		return labelValue(value)
	}
}

func parsePrinterDate(value string) (int64, bool) {
	if value == "" {
		return 0, false
	}
	// The page exposes only a date. UTC midnight keeps the metric stable and avoids
	// pretending the printer told us a timezone for this counter.
	parsed, err := time.ParseInLocation("02-01-2006", value, time.UTC)
	if err != nil {
		return 0, false
	}
	return parsed.Unix(), true
}

func parseCounter(value string) (float64, bool) {
	value = strings.ReplaceAll(value, ",", "")
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

func tableInFieldset(root *html.Node, legend string) *html.Node {
	fieldset := fieldsetByLegend(root, legend)
	if fieldset == nil {
		return nil
	}
	for node := range nodes(fieldset) {
		if node.Type == html.ElementNode && node.Data == "table" {
			return node
		}
	}
	return nil
}

func fieldsetByLegend(root *html.Node, legend string) *html.Node {
	for _, fieldset := range elements(root, "fieldset") {
		if strings.EqualFold(cleanText(textContent(firstChildElement(fieldset, "legend"))), legend) {
			return fieldset
		}
	}
	return nil
}

func tableRows(table *html.Node) [][]string {
	if table == nil {
		return nil
	}
	var rows [][]string
	for _, tr := range elements(table, "tr") {
		var row []string
		for child := tr.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && (child.Data == "th" || child.Data == "td") {
				row = append(row, cleanText(textContent(child)))
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}
	return rows
}

func cleanKey(value string) string {
	return strings.TrimSpace(strings.TrimSuffix(cleanText(value), ":"))
}

func cleanText(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(value, "\u00a0", " ")), " ")
}

func labelValue(value string) string {
	value = strings.ToLower(cleanText(value))
	replacer := strings.NewReplacer(
		"&", " and ",
		"/", "_",
		"-", "_",
		"(", "",
		")", "",
		",", "",
		".", "",
	)
	value = replacer.Replace(value)
	value = strings.Join(strings.Fields(value), "_")
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteString(" ")
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return builder.String()
}

func nodes(root *html.Node) func(func(*html.Node) bool) {
	return func(yield func(*html.Node) bool) {
		var walk func(*html.Node) bool
		walk = func(node *html.Node) bool {
			if node == nil {
				return true
			}
			if !yield(node) {
				return false
			}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if !walk(child) {
					return false
				}
			}
			return true
		}
		walk(root)
	}
}

func elements(root *html.Node, name string) []*html.Node {
	var result []*html.Node
	for node := range nodes(root) {
		if node.Type == html.ElementNode && node.Data == name {
			result = append(result, node)
		}
	}
	return result
}

func firstChildElement(root *html.Node, name string) *html.Node {
	if root == nil {
		return nil
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == name {
			return child
		}
	}
	return nil
}

func nextElementSibling(node *html.Node) *html.Node {
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}

func firstDescendantWithClass(root *html.Node, className string) *html.Node {
	for node := range nodes(root) {
		if node.Type == html.ElementNode && hasClass(node, className) {
			return node
		}
	}
	return nil
}

func descendantImageContains(root *html.Node, value string) bool {
	for node := range nodes(root) {
		if node.Type == html.ElementNode && node.Data == "img" && strings.Contains(attr(node, "src"), value) {
			return true
		}
	}
	return false
}

func hasClass(node *html.Node, className string) bool {
	for _, class := range strings.Fields(attr(node, "class")) {
		if class == className {
			return true
		}
	}
	return false
}

func attr(node *html.Node, name string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}
