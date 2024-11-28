package models

type Data struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Bandwidth int    `json:"bandwidth"`
}

type ScanResult struct {
	Protocol string `json:"protocol"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}

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
