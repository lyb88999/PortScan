package scanner

import "github.com/lyb88999/PortScan/internal/models"

type PortScanner interface {
	Scan(opts models.ScanOptions) ([]models.ScanResult, error)
	GetProgress(ip string, port int) (float64, error)
}
