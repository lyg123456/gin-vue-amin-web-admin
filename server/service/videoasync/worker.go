package videoasync

import (
	"context"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// Processor 由 initialize 注入，避免与 content 包循环依赖
type Processor func(shortVideoID uint) error

var (
	processor   Processor
	processorMu sync.RWMutex
)

func RegisterProcessor(fn Processor) {
	processorMu.Lock()
	defer processorMu.Unlock()
	processor = fn
}

func getProcessor() Processor {
	processorMu.RLock()
	defer processorMu.RUnlock()
	return processor
}

// Service 异步视频生成：Redis BLPOP → channel → 多 Worker 协程
type Service struct {
	jobCh    chan JobPayload
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	startOnce sync.Once
}

var defaultSvc = &Service{}

func Default() *Service {
	return defaultSvc
}

func Enabled() bool {
	return global.GVA_CONFIG.VideoAsync.Enabled
}

func (s *Service) Start() {
	if !Enabled() {
		global.GVA_LOG.Info("video-async 未启用，跳过 Worker 启动")
		return
	}
	s.startOnce.Do(func() {
		cfg := global.GVA_CONFIG.VideoAsync
		buf := cfg.ChannelBuffer
		if buf <= 0 {
			buf = 64
		}
		workers := cfg.WorkerCount
		if workers <= 0 {
			workers = 2
		}

		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		s.jobCh = make(chan JobPayload, buf)

		// Redis 分发协程：BLPOP → channel
		if global.GVA_CONFIG.System.UseRedis && global.GVA_REDIS != nil {
			s.wg.Add(1)
			go s.redisDispatcher(ctx)
			global.GVA_LOG.Info("video-async Redis 分发协程已启动", zap.String("queue", queueKey()))
		} else if global.GVA_CONFIG.VideoAsync.RequireRedis {
			global.GVA_LOG.Warn("video-async 要求 Redis 但未启用 system.use-redis，仅支持内存入队")
		} else {
			global.GVA_LOG.Info("video-async 使用内存 channel 队列（未接 Redis）")
		}

		for i := 0; i < workers; i++ {
			s.wg.Add(1)
			go s.workerLoop(ctx, i+1)
		}
		global.GVA_LOG.Info("video-async Worker 池已启动", zap.Int("workers", workers), zap.Int("channelBuffer", buf))
	})
}

func (s *Service) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Service) redisDispatcher(ctx context.Context) {
	defer s.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			payload, ok, err := PopRedisBlocking(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				global.GVA_LOG.Warn("video-async Redis 弹出失败", zap.Error(err))
				time.Sleep(time.Second)
				continue
			}
			if !ok {
				continue
			}
			s.dispatch(ctx, payload)
		}
	}
}

func (s *Service) dispatch(ctx context.Context, payload JobPayload) {
	select {
	case s.jobCh <- payload:
	case <-ctx.Done():
	case <-time.After(30 * time.Second):
		global.GVA_LOG.Error("video-async channel 已满，任务丢弃", zap.Uint("shortVideoId", payload.ShortVideoID))
	}
}

// Enqueue 提交任务：优先 Redis；无 Redis 时写入内存 channel
func (s *Service) Enqueue(ctx context.Context, payload JobPayload) error {
	if !Enabled() {
		return ErrAsyncDisabled
	}
	if global.GVA_CONFIG.VideoAsync.RequireRedis {
		if !global.GVA_CONFIG.System.UseRedis || global.GVA_REDIS == nil {
			return ErrRedisRequired
		}
		return PushRedis(ctx, payload)
	}
	if global.GVA_CONFIG.System.UseRedis && global.GVA_REDIS != nil {
		return PushRedis(ctx, payload)
	}
	if s.jobCh == nil {
		return ErrWorkerNotStarted
	}
	select {
	case s.jobCh <- payload:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return ErrChannelFull
	}
}

func (s *Service) workerLoop(ctx context.Context, workerID int) {
	defer s.wg.Done()
	fn := getProcessor()
	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-s.jobCh:
			if !ok {
				return
			}
			if fn == nil {
				global.GVA_LOG.Error("video-async 未注册 Processor", zap.Int("worker", workerID))
				continue
			}
			global.GVA_LOG.Info("video-async 开始处理",
				zap.Int("worker", workerID),
				zap.Uint("jobId", payload.JobID),
				zap.Uint("shortVideoId", payload.ShortVideoID),
			)
			if err := fn(payload.ShortVideoID); err != nil {
				global.GVA_LOG.Warn("video-async 处理失败",
					zap.Uint("shortVideoId", payload.ShortVideoID),
					zap.Error(err),
				)
			}
		}
	}
}
