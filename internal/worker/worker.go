package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lyb88999/PortScan/internal/models"

	"github.com/lyb88999/PortScan/internal/kafka"
	"github.com/lyb88999/PortScan/internal/scanner"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Worker struct {
	ctx      context.Context
	redisCli *redis.Client
	producer *kafka.Producer
	cg       *kafka.ConsumerGroup
	ps       scanner.PortScanner
}

func NewWorker(ctx context.Context, redisCli *redis.Client, producer *kafka.Producer, cg *kafka.ConsumerGroup, ps scanner.PortScanner) *Worker {
	return &Worker{
		ctx:      ctx,
		redisCli: redisCli,
		producer: producer,
		cg:       cg,
		ps:       ps,
	}
}

func (w *Worker) Run() {
	// 1. 从Kafka中读取出任务: 从cg.outputChan中读取任务 调用PortScanner的Scan方法
	go func() {
		err := w.cg.Consume(w.ctx)
		if err != nil {
			fmt.Println("failed to consume: ", err)
		}
	}()

	for {
		select {
		case err := <-w.cg.ErrorChan:
			fmt.Printf("Consumer error: %v\n", err)
			return
		case msgValue, ok := <-w.cg.OutputChan:
			if !ok {
				// channel 已关闭，退出循环
				return
			}
			var scanOptions models.ScanOptions
			err := json.Unmarshal(msgValue, &scanOptions)
			if err != nil {
				fmt.Println("failed to unmarshal scan options: ", err)
				continue
			}
			resultChan, progressChan, err := w.ps.Scan(scanOptions)
			if err != nil {
				fmt.Println("failed to scan: ", err)
				continue
			}

			// 2. 扫描然后将进度写入Redis: 读取Scan方法返回的progressChan，将里面的数据写入Redis
			eg, ctx := errgroup.WithContext(w.ctx)
			redisKey := fmt.Sprintf("%s:%d:%d", scanOptions.IP, scanOptions.Port, time.Now().UnixNano())
			eg.Go(func() error {
				for progress := range progressChan {
					fmt.Println(progress)
					err = w.redisCli.Set(ctx, redisKey, progress, time.Hour).Err()
					if err != nil {
						return fmt.Errorf("failed to save progress to redis: %w", err)
					}
				}
				fmt.Println("port scan done")
				return nil
			})

			// 3. 将扫描结果发到Kafka: 读取Scan方法返回的resultChan，将里面的数据写入Kafka
			eg.Go(func() error {
				for result := range resultChan {
					if err = w.producer.Send(result); err != nil {
						return fmt.Errorf("failed to send task: %w", err)
					}
				}
				fmt.Println("send port scan result done")
				return nil
			})
			if err = eg.Wait(); err != nil {
				fmt.Println("error in worker: ", err)
			}
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Worker) Stop() error {
	var errors []error

	if err := w.cg.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close consumer group: %w", err))
	}

	if err := w.producer.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close producer: %w", err))
	}

	if err := w.redisCli.Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close redis client: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while stopping worker: %v", errors)
	}

	return nil
}
