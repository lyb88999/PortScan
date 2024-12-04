package scanner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/lyb88999/PortScan/internal/models"
	"golang.org/x/sync/errgroup"
)

type masscanScanner struct {
}

func NewMasscanScanner() PortScanner {
	return &masscanScanner{}
}

func (m *masscanScanner) Scan(opts models.ScanOptions) (chan models.ScanResult, chan float64, error) {
	// 检查 masscan 是否已安装
	if _, err := exec.LookPath("masscan"); err != nil {
		return nil, nil, fmt.Errorf("masscan not found in PATH: %v", err)
	}

	// 构建命令行参数
	args := []string{
		opts.IP,
		"-p", fmt.Sprintf("%d", opts.Port),
		"--rate", strconv.Itoa(opts.BandWidth),
		"--wait", "0", // 扫描完成后立即退出
		// "-oJ", "-", // 输出JSON格式到标准输出
	}

	// 创建命令
	cmd := exec.Command("masscan", args...)

	// 获取标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// 获取标准错误管道
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	resultChan := make(chan models.ScanResult, 1000)
	progressChan := make(chan float64, 1000)

	eg := &errgroup.Group{}

	// 处理标准输出
	eg.Go(func() error {
		defer close(resultChan)
		return m.processDefaultOutput(stdout, resultChan)
	})

	// 处理标准错误（进度信息）
	eg.Go(func() error {
		defer close(progressChan)
		return m.processProgressOutput(stderr, progressChan)
	})

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start masscan: %v", err)
	}

	// 在后台等待命令完成并处理可能的错误
	go func() {
		if err := eg.Wait(); err != nil {
			log.Printf("Error in goroutines: %v", err)
		}
		if err := cmd.Wait(); err != nil {
			log.Printf("Masscan command failed: %v", err)
		}
	}()

	return resultChan, progressChan, nil
}

// 处理 JSON 输出格式
func (m *masscanScanner) processJSONOutput(stdout io.Reader, resultChan chan<- models.ScanResult) error {
	scanner := bufio.NewScanner(stdout)
	outputRegex := regexp.MustCompile(`\{.*?\} ] }`)

	for scanner.Scan() {
		// 解析 masscan 的 JSON 输出
		var masscanResult models.MasscanResult
		if matches := outputRegex.FindStringSubmatch(scanner.Text()); len(matches) >= 1 {
			if err := json.Unmarshal([]byte(matches[0]), &masscanResult); err != nil {
				log.Printf("Error parsing JSON: %v, line: %s\n", err, scanner.Text())
				continue
			}

			// 转换并发送结果
			for _, port := range masscanResult.Ports {
				result := models.ScanResult{
					IP:       masscanResult.IP,
					Port:     port.Port,
					Protocol: port.Proto,
				}
				select {
				case resultChan <- result:
				default:
					log.Printf("Warning: result channel is full, dropping result for IP: %s, Port: %d",
						result.IP, result.Port)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read stdout: %s", err)
	}
	return nil
}

// 处理默认输出格式
func (m *masscanScanner) processDefaultOutput(stdout io.Reader, resultChan chan<- models.ScanResult) error {
	scanner := bufio.NewScanner(stdout)
	outputRegex := regexp.MustCompile(`Discovered open port (\d+)/(\w+) on ([0-9.]+)`)

	for scanner.Scan() {
		if matches := outputRegex.FindStringSubmatch(scanner.Text()); len(matches) >= 4 {
			port, err := strconv.Atoi(matches[1])
			if err != nil {
				log.Printf("Error parsing port: %v\n", err)
				continue
			}

			result := models.ScanResult{
				IP:       matches[3],
				Port:     port,
				Protocol: matches[2],
			}

			select {
			case resultChan <- result:
			default:
				log.Printf("Warning: result channel is full, dropping result for IP: %s, Port: %d",
					result.IP, result.Port)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read stdout: %s", err)
	}
	return nil
}

// 处理进度输出
func (m *masscanScanner) processProgressOutput(stderr io.Reader, progressChan chan<- float64) error {
	progressRegex := regexp.MustCompile(`rate:\s+[\d.]+-kpps,\s+([\d.]+)% done`)
	buffer := make([]byte, 1024)
	for {
		n, err := stderr.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("error reading stderr: %v", err)
			}
			return nil
		}

		line := string(buffer[:n])
		if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
			if progress, err := strconv.ParseFloat(matches[1], 64); err == nil {
				select {
				case progressChan <- progress:
				default:
					// 如果channel已满，记录警告但继续处理
					log.Printf("Warning: progress channel is full, dropping progress: %f ", progress)
				}
			}
		}
	}
}
