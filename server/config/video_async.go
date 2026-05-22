package config

// VideoAsync 短视频成片异步 Worker（goroutine + channel + Redis 队列）
type VideoAsync struct {
	Enabled            bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	RequireRedis       bool   `mapstructure:"require-redis" json:"require-redis" yaml:"require-redis"`
	WorkerCount        int    `mapstructure:"worker-count" json:"worker-count" yaml:"worker-count"`
	QueueKey           string `mapstructure:"queue-key" json:"queue-key" yaml:"queue-key"`
	ChannelBuffer      int    `mapstructure:"channel-buffer" json:"channel-buffer" yaml:"channel-buffer"`
	RedisPopTimeoutSec int    `mapstructure:"redis-pop-timeout-sec" json:"redis-pop-timeout-sec" yaml:"redis-pop-timeout-sec"`
	MaxAttempts        int    `mapstructure:"max-attempts" json:"max-attempts" yaml:"max-attempts"`
}
