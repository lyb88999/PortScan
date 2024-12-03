package models

// Data 客户端输入参数
type Data struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Bandwidth int    `json:"bandwidth"`
}

// ScanOptions 扫描参数
type ScanOptions struct {
	IP        string
	Port      int
	BandWidth string
}

// ScanResult 返回结果
type ScanResult struct {
	Protocol string `json:"protocol"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}

// MasscanResult raw结果
type MasscanResult struct {
	IP        string `json:"ip"`
	TimeStamp string `json:"timestamp"`
	Ports     []struct {
		Port   int    `json:"port"`
		Proto  string `json:"proto"`
		Status string `json:"status"`
		Reason string `json:"reason"`
		TTL    int    `json:"ttl"`
	} `json:"ports"`
}
