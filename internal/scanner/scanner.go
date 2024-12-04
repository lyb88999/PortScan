package scanner

import "github.com/lyb88999/PortScan/internal/models"

type PortScanner interface {
	Scan(opts models.ScanOptions) (chan models.ScanResult, chan float64, error)
}
