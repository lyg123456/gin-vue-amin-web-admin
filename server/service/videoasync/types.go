package videoasync

// JobPayload Redis 队列与内存 channel 统一消息体
type JobPayload struct {
	JobID        uint `json:"jobId"`
	ShortVideoID uint `json:"shortVideoId"`
}
