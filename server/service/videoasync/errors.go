package videoasync

import "errors"

var (
	ErrAsyncDisabled    = errors.New("未启用 video-async，请在 config.yaml 配置 video-async.enabled=true")
	ErrRedisRequired    = errors.New("异步视频生成需要 Redis：请设置 system.use-redis=true 并配置 redis.addr")
	ErrWorkerNotStarted = errors.New("video-async Worker 未启动")
	ErrChannelFull      = errors.New("视频任务通道已满，请稍后重试")
)
