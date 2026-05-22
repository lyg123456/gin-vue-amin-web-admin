package content

import (
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

func videoGenProvider() string {
	if dashScopeVideoService.Enabled() {
		return "dashscope"
	}
	if volcArkVideoGenerateService.Enabled() {
		return "volc"
	}
	return ""
}

// ExecuteVideoGeneration 同步执行成片（供异步 Worker 或关闭 async 时直接调用）
func ExecuteVideoGeneration(shortVideoID uint) error {
	v, err := shortVideoService.Find(shortVideoID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(v.Script) == "" {
		return errEmptyScript
	}

	job, jobErr := videoGenJobService.FindLatestByShortVideo(shortVideoID)
	if jobErr == nil && (job.Status == "queued" || job.Status == "processing") {
		_ = videoGenJobService.MarkProcessing(job.ID)
	}
	_ = shortVideoService.SetGenerating(shortVideoID)

	var taskID, videoURL string
	SyncShortVideoFrames(&v)
	if dashScopeVideoService.Enabled() {
		taskID, videoURL, err = dashScopeVideoService.SubmitGeneration(v.Script, v.FirstFrameURL, v.LastFrameURL, v.DurationSec)
	} else {
		taskID, videoURL, err = volcArkVideoGenerateService.SubmitGeneration(v.Script, v.SourceImages, v.DurationSec)
	}

	if jobErr == nil {
		if err != nil {
			_ = videoGenJobService.MarkFinished(job.ID, "failed", taskID, err.Error())
		} else {
			_ = videoGenJobService.MarkFinished(job.ID, "succeeded", taskID, "")
		}
	}

	if err != nil {
		_ = shortVideoService.SetGenerationResult(shortVideoID, "failed", "", taskID, err.Error())
		return err
	}
	st := "ready"
	if videoURL == "" && taskID != "" {
		st = "generating"
	}
	_ = shortVideoService.SetGenerationResult(shortVideoID, st, videoURL, taskID, "")
	return nil
}

var errEmptyScript = &genError{msg: "请先生成或填写脚本"}

type genError struct{ msg string }

func (e *genError) Error() string { return e.msg }

// EnqueueVideoGeneration 异步入队（Redis + channel Worker）
func EnqueueVideoGeneration(shortVideoID uint) (jobID uint, err error) {
	if !global.GVA_CONFIG.VideoAsync.Enabled {
		return 0, errAsyncUseSync
	}
	v, err := shortVideoService.Find(shortVideoID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(v.Script) == "" {
		return 0, errEmptyScript
	}
	pending, err := videoGenJobService.HasPending(shortVideoID)
	if err != nil {
		return 0, err
	}
	if pending {
		return 0, errJobAlreadyPending
	}
	provider := videoGenProvider()
	if provider == "" {
		return 0, errNoVideoProvider
	}

	job, err := videoGenJobService.CreateQueued(shortVideoID, provider)
	if err != nil {
		return 0, err
	}
	_ = shortVideoService.SetGenerationResult(shortVideoID, "queued", "", "", "")

	if err := enqueueVideoJob(job.ID, shortVideoID); err != nil {
		_ = videoGenJobService.MarkFinished(job.ID, "failed", "", err.Error())
		_ = shortVideoService.SetGenerationResult(shortVideoID, "failed", "", "", err.Error())
		return 0, err
	}
	return job.ID, nil
}

var (
	errAsyncUseSync      = &genError{msg: "异步未启用"}
	errJobAlreadyPending = &genError{msg: "该短视频已有排队或进行中的生成任务"}
	errNoVideoProvider   = &genError{msg: "未配置 dashscope-video 或 volc-ark-video，无法生成成片"}
)
