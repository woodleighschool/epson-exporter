package epson

type Paths struct {
	ProductStatus  string
	UsageStatus    string
	NetworkStatus  string
	HardwareStatus string
}

type Snapshot struct {
	Model              string
	Firmware           string
	MacAddress         string
	DeviceName         string
	ConnectionStatus   string
	IPAddress          string
	DNSHostName        string
	PrinterStatus      string
	ScannerStatus      string
	FirstPrintDate     string
	FirstPrintUnixTime float64
	Consumables        []Consumable
	PaperSources       []PaperSource
	PageTotals         []PageTotal
	PagesBySize        []PagesBySize
	PagesByFunction    []PagesByFunction
	PagesByLanguage    []PagesByLanguage
	PagesByInterface   []PagesByInterface
	HardwareStatuses   []ComponentStatus
}

type Consumable struct {
	Type         string
	Slot         string
	Color        string
	Model        string
	LevelPercent float64
	Warning      bool
}

type PaperSource struct {
	Name          string
	Size          string
	Type          string
	Level         string
	LevelFraction float64
	HasLevel      bool
}

type PageTotal struct {
	Kind  string
	Value float64
}

type PagesBySize struct {
	Size      string
	Sides     string
	ColorMode string
	Value     float64
}

type PagesByFunction struct {
	Function  string
	ColorMode string
	Value     float64
}

type PagesByLanguage struct {
	Language string
	Value    float64
}

type PagesByInterface struct {
	Interface string
	ColorMode string
	Value     float64
}

type ComponentStatus struct {
	Component string
	Status    string
}
