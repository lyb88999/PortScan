package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const scannerName = "masscan"

type masscanScanner struct {
	redisCli *redis.Client
}

func NewMasscanScanner(redisCli *redis.Client) PortScanner {
	return &masscanScanner{
		redisCli: redisCli,
	}
}

func (m *masscanScanner) getProgressKey(ip string, port int) string {
	return fmt.Sprintf("scan_progress:%s:%s:%d", scannerName, ip, port)
}

func (m *masscanScanner) Scan(opts ScanOptions) ([]ScanResult, error) {
	// 检查 masscan 是否已安装
	if _, err := exec.LookPath("masscan"); err != nil {
		return nil, fmt.Errorf("masscan not found in PATH: %v", err)
	}

	// 构建命令行参数
	args := []string{
		opts.IP,
		"-p", fmt.Sprintf("%d", opts.Port),
		"--rate", opts.BandWidth,
		"--wait", "0", // 扫描完成后立即退出
		"-oJ", "-", // 输出JSON格式到标准输出
	}

	// 创建命令
	cmd := exec.Command("masscan", args...)

	// 获取标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	// 获取标准错误管道
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	// 存储结果的切片
	var results []ScanResult
	var resultsMutex sync.Mutex

	// 创建wg 对应两个协程分别来处理stdout和stderr
	var wg sync.WaitGroup
	wg.Add(2) // 一个用于stdout，一个用于stderr

	// 处理标准输出
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// log.Println(line)
			// 跳过不是 JSON 开头的行（无关行）
			if !strings.HasPrefix(line, "{") {
				continue
			}
			// 跳过空行和中括号
			if line == "" || line == "[" || line == "]" {
				continue
			}
			// 如果行末尾有逗号，去除它
			if strings.HasSuffix(line, ",") {
				line = line[:len(line)-1]
			}

			// 解析 masscan 的 JSON 输出
			var masscanResult struct {
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

			if err := json.Unmarshal([]byte(line), &masscanResult); err != nil {
				fmt.Printf("Error parsing JSON: %v, line: %s\n", err, line)
				continue
			}

			// 转换为 ScanResult 格式
			for _, port := range masscanResult.Ports {
				result := ScanResult{
					IP:       masscanResult.IP,
					Port:     port.Port,
					Protocol: port.Proto,
				}

				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading stdout: %v\n", err)
		}
	}()

	// 处理标准错误（进度信息）
	go func() {
		defer wg.Done()
		progressRegex := regexp.MustCompile(`rate:\s+[\d.]+-kpps,\s+([\d.]+)% done`)
		buffer := make([]byte, 1024)

		for {
			n, err := stderr.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("Error reading stderr: %v", err)
				}
				return
			}

			line := string(buffer[:n])
			if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
				if progress, err := strconv.ParseFloat(matches[1], 64); err == nil {
					log.Println(progress)
					// 将进度保存到 Redis
					progressKey := m.getProgressKey(opts.IP, opts.Port)
					if err := m.redisCli.Set(context.Background(), progressKey, progress, 0).Err(); err != nil {
						fmt.Printf("Error saving progress to Redis: %v\n", err)
					}
				}
			}
		}
	}()
	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start masscan: %v", err)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("masscan command failed: %v", err)
	}

	return results, nil
}

//	// 处理标准错误（进度信息）
//	go func() {
//		defer wg.Done()
//		scanner := bufio.NewScanner(stderr)
//		progressRegex := regexp.MustCompile(`\s*(\d+\.\d+)%\s*done`)
//
//		for scanner.Scan() {
//			line := scanner.Text()
//			if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
//				if progress, err := strconv.ParseFloat(matches[1], 64); err == nil {
//					// 将进度保存到 Redis
//					progressKey := m.getProgressKey(opts.IP, opts.Port)
//					if err := m.redisCli.Set(context.Background(), progressKey, progress, 0).Err(); err != nil {
//						fmt.Printf("Error saving progress to Redis: %v\n", err)
//					}
//				}
//			}
//		}
//
//		if err := scanner.Err(); err != nil {
//			fmt.Printf("Error reading stderr: %v\n", err)
//		}
//	}()
//
//	// 启动命令
//	if err := cmd.Start(); err != nil {
//		return nil, fmt.Errorf("failed to start masscan: %v", err)
//	}
//
//	// 等待所有goroutine完成
//	wg.Wait()
//
//	// 等待命令完成
//	if err := cmd.Wait(); err != nil {
//		return nil, fmt.Errorf("masscan command failed: %v", err)
//	}
//
//	return results, nil
//}

func (m *masscanScanner) GetProgress(ip string, port int) (float64, error) {
	progressKey := m.getProgressKey(ip, port)
	progress, err := m.redisCli.Get(context.Background(), progressKey).Float64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get progress from Redis: %v", err)
	}
	return progress, nil
}
