package content

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/service/videoasync"
)

func enqueueVideoJob(jobID, shortVideoID uint) error {
	return videoasync.Default().Enqueue(context.Background(), videoasync.JobPayload{
		JobID:        jobID,
		ShortVideoID: shortVideoID,
	})
}
