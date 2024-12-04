package models

// ScanOptions 扫描参数
type ScanOptions struct {
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	BandWidth int    `json:"bandwidth"`
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

type Channels struct {
	OutputChan   chan []byte
	ErrorChan    chan error
	ResultChan   chan ScanResult
	ProgressChan chan float64
}

func NewChannels(buffersize int) *Channels {
	return &Channels{
		ResultChan:   make(chan ScanResult, buffersize),
		ProgressChan: make(chan float64, buffersize),
		OutputChan:   make(chan []byte, buffersize),
		ErrorChan:    make(chan error, buffersize),
	}
}
