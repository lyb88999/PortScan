package scanner

// ScanOptions 扫描参数
type ScanOptions struct {
	IP        string
	Port      int
	BandWidth string
}

// ScanResult 扫描结果
type ScanResult struct {
	Protocol string
	IP       string
	Port     int
}

type PortScanner interface {
	Scan(opts ScanOptions) ([]ScanResult, error)
	GetProgress(ip string, port int) (float64, error)
}
