package scanner

import (
	"github.com/lyb88999/PortScan/internal/models"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestMasscanScanner(t *testing.T) {
	// 加载配置文件
	// cfg, err := config.LoadConfig("../..")
	// require.NoError(t, err)
	// 创建redis client
	// var ctx = context.Background()
	var ms = NewMasscanScanner()
	scanOptions := models.ScanOptions{
		IP:        "47.93.190.244/24",
		Port:      8080,
		BandWidth: 10,
	}
	resultChan, progressChan, err := ms.Scan(scanOptions)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for result := range resultChan {
			t.Logf("result: %+v", result)
		}
	}()
	go func() {
		defer wg.Done()
		for result := range progressChan {
			t.Logf("progress: %+v", result)
		}
	}()
	wg.Wait()

}
