package videoasync

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/redis/go-redis/v9"
)

func queueKey() string {
	k := global.GVA_CONFIG.VideoAsync.QueueKey
	if k == "" {
		return "gva:video:gen:queue"
	}
	return k
}

func redisPopTimeout() time.Duration {
	sec := global.GVA_CONFIG.VideoAsync.RedisPopTimeoutSec
	if sec <= 0 {
		sec = 5
	}
	return time.Duration(sec) * time.Second
}

// PushRedis 任务写入 Redis 列表（RPUSH，Worker 端 BLPOP）
func PushRedis(ctx context.Context, payload JobPayload) error {
	if global.GVA_REDIS == nil {
		return errors.New("Redis 未初始化")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return global.GVA_REDIS.RPush(ctx, queueKey(), raw).Err()
}

// PopRedisBlocking 阻塞弹出一条（带超时，便于优雅退出）
func PopRedisBlocking(ctx context.Context) (JobPayload, bool, error) {
	if global.GVA_REDIS == nil {
		return JobPayload{}, false, errors.New("Redis 未初始化")
	}
	res, err := global.GVA_REDIS.BLPop(ctx, redisPopTimeout(), queueKey()).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return JobPayload{}, false, nil
		}
		return JobPayload{}, false, err
	}
	if len(res) < 2 {
		return JobPayload{}, false, nil
	}
	var p JobPayload
	if err := json.Unmarshal([]byte(res[1]), &p); err != nil {
		return JobPayload{}, false, err
	}
	if p.ShortVideoID == 0 {
		return JobPayload{}, false, errors.New("无效任务: shortVideoId 为空")
	}
	return p, true, nil
}

// QueueLength Redis 队列长度（监控用）
func QueueLength(ctx context.Context) (int64, error) {
	if global.GVA_REDIS == nil {
		return 0, errors.New("Redis 未初始化")
	}
	return global.GVA_REDIS.LLen(ctx, queueKey()).Result()
}
