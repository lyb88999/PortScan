package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
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
	eg, ctx := errgroup.WithContext(w.ctx)

	// 消费者启动
	eg.Go(func() error {
		return w.cg.Consume(ctx)
	})

	// 处理消息启动
	eg.Go(func() error {
		return w.processMessages(ctx)
	})

	// 等待所有goroutine完成
	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("Worker stopped with err")
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

func (w *Worker) processMessages(ctx context.Context) error {
	for {
		select {
		case err := <-w.cg.ErrorChan:
			return fmt.Errorf("consumer error: %w", err)
		// 从Kafka中读取出任务 从cg.outputChan中读取任务 调用handleMessage处理
		case msgValue, ok := <-w.cg.OutputChan:
			if !ok {
				return nil
			}
			if err := w.handleMessage(ctx, msgValue); err != nil {
				log.Error().Err(err).Msg("failed to handle message")
				continue
			}
		case <-ctx.Done():
			return ctx.Err()
		}

	}
}

func (w *Worker) handleMessage(ctx context.Context, msgValue []byte) error {
	// 解析任务 调用portScanner的Scan方法 返回resultChan和progressChan
	var scanOptions models.ScanOptions
	if err := json.Unmarshal(msgValue, &scanOptions); err != nil {
		return fmt.Errorf("failde to unmarshal scan options: %w", err)
	}
	resultChan, progressChan, err := w.ps.Scan(scanOptions)
	if err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}
	eg, ctx := errgroup.WithContext(ctx)
	// 处理进度写入Redis
	redisKey := fmt.Sprintf("%s:%d:%d", scanOptions.IP, scanOptions.Port, time.Now().UnixNano())
	eg.Go(func() error {
		for progress := range progressChan {
			log.Info().Float64("progress", progress).Msg("port scan progress")
			if err := w.redisCli.Set(ctx, redisKey, progress, time.Hour).Err(); err != nil {
				return fmt.Errorf("failed to save progress to redis: %w", err)
			}
		}
		log.Info().Msg("save progress to redis done")
		return nil
	})

	// 处理解析结果发送到Kafka
	eg.Go(func() error {
		for result := range resultChan {
			if err := w.producer.Send(result); err != nil {
				return fmt.Errorf("failed to send result: %w", err)
			}
		}
		log.Info().Msg("send port scan result")
		return nil
	})
	return eg.Wait()
}
