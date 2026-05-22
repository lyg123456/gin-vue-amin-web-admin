package content

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"sort"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type OfficeWatermarkService struct{}

type WatermarkPreset struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Platform    string  `json:"platform"`
	Description string  `json:"description"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	W           float64 `json:"w"`
	H           float64 `json:"h"`
	Media       string  `json:"media"`
}

type WatermarkCapabilities struct {
	ImageSupport bool                `json:"imageSupport"`
	VideoSupport bool                `json:"videoSupport"`
	FFmpeg       bool                `json:"ffmpeg"`
	MaxUploadMB  int                 `json:"maxUploadMB"`
	Presets      []WatermarkPreset   `json:"presets"`
	Methods      []map[string]string `json:"methods"`
	Hint         string              `json:"hint"`
}

func watermarkPresets() []WatermarkPreset {
	return []WatermarkPreset{
		{ID: "douyin", Name: "抖音", Platform: "抖音", Description: "右下水印与账号条", X: 72, Y: 82, W: 26, H: 14, Media: "both"},
		{ID: "kuaishou", Name: "快手", Platform: "快手", Description: "右下平台标识", X: 70, Y: 80, W: 28, H: 16, Media: "both"},
		{ID: "xiaohongshu", Name: "小红书", Platform: "小红书", Description: "底部整条水印栏", X: 0, Y: 86, W: 100, H: 14, Media: "both"},
		{ID: "bilibili", Name: "B站", Platform: "哔哩哔哩", Description: "右上 UP 标识", X: 80, Y: 2, W: 18, H: 12, Media: "both"},
		{ID: "tiktok", Name: "TikTok", Platform: "TikTok", Description: "右下用户名", X: 74, Y: 84, W: 24, H: 12, Media: "both"},
		{ID: "weixin_channels", Name: "视频号", Platform: "微信视频号", Description: "右下视频号标", X: 68, Y: 84, W: 30, H: 14, Media: "both"},
		{ID: "instagram", Name: "Instagram", Platform: "Instagram", Description: "顶部用户条", X: 0, Y: 0, W: 100, H: 10, Media: "both"},
		{ID: "youtube_shorts", Name: "YouTube Shorts", Platform: "YouTube", Description: "右下频道元素", X: 76, Y: 88, W: 22, H: 10, Media: "both"},
		{ID: "generic_br", Name: "通用-右下", Platform: "通用", Description: "右下角角标", X: 76, Y: 84, W: 22, H: 12, Media: "both"},
		{ID: "generic_bl", Name: "通用-左下", Platform: "通用", Description: "左下角标", X: 2, Y: 84, W: 22, H: 12, Media: "both"},
		{ID: "jimeng", Name: "即梦AI", Platform: "即梦 / AI绘图", Description: "底部居中 AI 标识", X: 0, Y: 86, W: 100, H: 14, Media: "image"},
		{ID: "ai_bottom", Name: "AI底栏", Platform: "通用AI", Description: "底部整条半透明标", X: 0, Y: 84, W: 100, H: 16, Media: "image"},
		{ID: "custom", Name: "自定义", Platform: "自定义", Description: "请框住完整水印含文字", X: 0, Y: 85, W: 100, H: 15, Media: "both"},
	}
}

func presetByID(id string) WatermarkPreset {
	for _, p := range watermarkPresets() {
		if p.ID == id {
			return p
		}
	}
	return watermarkPresets()[len(watermarkPresets())-1]
}

func (s *OfficeWatermarkService) maxUploadMB() int {
	mb := global.GVA_CONFIG.OfficeTools.MaxUploadMB
	if mb <= 0 {
		return 50
	}
	return mb
}

func (s *OfficeWatermarkService) Capabilities() WatermarkCapabilities {
	ff := ffmpegBin() != ""
	hint := "图片：区域模糊或裁剪；视频：FFmpeg delogo。固定角标效果较好。"
	if !ff {
		hint = "未检测到 FFmpeg，当前仅支持图片去水印。"
	}
	return WatermarkCapabilities{
		ImageSupport: true,
		VideoSupport: ff,
		FFmpeg:       ff,
		MaxUploadMB:  s.maxUploadMB(),
		Presets:      watermarkPresets(),
		Methods: []map[string]string{
			{"id": "fill", "name": "智能填充", "desc": "推荐：采样周围背景色覆盖，适合即梦AI等半透明标"},
			{"id": "blur", "name": "强力模糊", "desc": "多遍高斯模糊"},
			{"id": "crop", "name": "裁剪去除", "desc": "裁掉含水印边缘"},
			{"id": "delogo", "name": "Delogo", "desc": "视频专用"},
		},
		Hint: hint,
	}
}

type watermarkRegion struct {
	x, y, w, h int
}

func resolveRegion(imgW, imgH int, presetID, method string, cx, cy, cw, ch float64) watermarkRegion {
	p := presetByID(presetID)
	xp, yp, wp, hp := p.X, p.Y, p.W, p.H
	if presetID == "custom" && cx >= 0 {
		xp, yp, wp, hp = cx, cy, cw, ch
	}
	x := int(float64(imgW) * xp / 100)
	y := int(float64(imgH) * yp / 100)
	w := int(float64(imgW) * wp / 100)
	h := int(float64(imgH) * hp / 100)
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	if x+w > imgW {
		w = imgW - x
	}
	if y+h > imgH {
		h = imgH - y
	}
	if method == "crop" && (presetID == "xiaohongshu" || hp >= 50) {
		return watermarkRegion{0, 0, imgW, y}
	}
	return expandRegion(watermarkRegion{x, y, w, h}, imgW, imgH, padPercent(method))
}

// padPercent 处理区域外扩，避免水印文字压在框边缘漏去
func padPercent(method string) float64 {
	switch method {
	case "crop":
		return 0
	case "delogo":
		return 12
	default:
		return 18
	}
}

func expandRegion(r watermarkRegion, imgW, imgH int, padPct float64) watermarkRegion {
	if padPct <= 0 {
		return r
	}
	padX := int(float64(r.w) * padPct / 100)
	padY := int(float64(r.h) * padPct / 100)
	if padX < 10 {
		padX = 10
	}
	if padY < 10 {
		padY = 10
	}
	x := r.x - padX
	y := r.y - padY
	w := r.w + 2*padX
	h := r.h + 2*padY
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > imgW {
		w = imgW - x
	}
	if y+h > imgH {
		h = imgH - y
	}
	if w < 8 {
		w = 8
	}
	if h < 8 {
		h = 8
	}
	return watermarkRegion{x, y, w, h}
}

func encodeImage(out image.Image, ext string) ([]byte, string, string, error) {
	buf := new(bytes.Buffer)
	baseExt := ".png"
	mime := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		baseExt = ".jpg"
		mime = "image/jpeg"
		if err := jpeg.Encode(buf, out, &jpeg.Options{Quality: 92}); err != nil {
			return nil, "", "", err
		}
	} else {
		if err := png.Encode(buf, out); err != nil {
			return nil, "", "", err
		}
	}
	return buf.Bytes(), "removed_wm" + baseExt, mime, nil
}

func (s *OfficeWatermarkService) RemoveWatermark(
	file *multipart.FileHeader,
	presetID, method string,
	customX, customY, customW, customH float64,
) ([]byte, string, string, error) {
	if file == nil {
		return nil, "", "", errors.New("请上传图片或视频")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	isVideo := ext == ".mp4" || ext == ".mov" || ext == ".webm" || ext == ".mkv" || ext == ".avi"
	if isVideo {
		if method == "crop" {
			method = "delogo"
		}
		return s.removeVideoWatermark(file, presetID, method, customX, customY, customW, customH)
	}
	if method == "delogo" {
		method = "fill"
	}
	if method == "" {
		method = "fill"
	}
	return s.removeImageWatermark(file, presetID, method, customX, customY, customW, customH)
}

func (s *OfficeWatermarkService) removeImageWatermark(
	file *multipart.FileHeader,
	presetID, method string,
	cx, cy, cw, ch float64,
) ([]byte, string, string, error) {
	f, err := file.Open()
	if err != nil {
		return nil, "", "", err
	}
	defer f.Close()
	img, err := imaging.Decode(f)
	if err != nil {
		return nil, "", "", fmt.Errorf("无法解码图片: %w", err)
	}
	bounds := img.Bounds()
	imgW, imgH := bounds.Dx(), bounds.Dy()
	region := resolveRegion(imgW, imgH, presetID, method, cx, cy, cw, ch)

	var out image.Image
	switch method {
	case "crop":
		cropH := region.h
		if presetID == "xiaohongshu" || region.y > 0 {
			cropH = region.y
		}
		if cropH <= 0 || cropH >= imgH {
			return nil, "", "", errors.New("裁剪高度无效，请调整区域或使用模糊")
		}
		out = imaging.Crop(img, image.Rect(0, 0, imgW, cropH))
	default:
		out = applyWatermarkRemoval(img, region, method)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	data, name, mime, err := encodeImage(out, ext)
	if err != nil {
		return nil, "", "", err
	}
	base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	return data, base + "_nowm" + filepath.Ext(name), mime, nil
}

func applyWatermarkRemoval(img image.Image, region watermarkRegion, method string) image.Image {
	r := image.Rect(region.x, region.y, region.x+region.w, region.y+region.h)
	bounds := img.Bounds()
	out := imaging.Clone(img)

	switch method {
	case "blur":
		patch := imaging.Crop(img, r)
		patch = imaging.Blur(patch, 18.0)
		patch = imaging.Blur(patch, 18.0)
		out = imaging.Paste(out, patch, image.Pt(region.x, region.y))
	default:
		// fill：采样水印区外一圈像素的中位色，再覆盖并轻微模糊接缝
		fill := sampleSurroundingFill(img, r, bounds)
		patch := imaging.Clone(fill)
		patch = imaging.Blur(patch, 2.5)
		out = imaging.Paste(out, patch, image.Pt(region.x, region.y))
		// 二次：对整块再采样一次（覆盖漏边）
		r2 := expandRegion(region, bounds.Dx(), bounds.Dy(), 6)
		rBig := image.Rect(r2.x, r2.y, r2.x+r2.w, r2.y+r2.h)
		fill2 := sampleSurroundingFill(out, rBig, bounds)
		out = imaging.Paste(out, fill2, image.Pt(r2.x, r2.y))
	}
	return out
}

// sampleSurroundingFill 取区域外围环带像素的中位色并生成填充块
func sampleSurroundingFill(img image.Image, r image.Rectangle, bounds image.Rectangle) image.Image {
	ring := 4
	samples := make([][3]int, 0, 256)
	collect := func(x0, y0, x1, y1 int) {
		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x < bounds.Min.X || y < bounds.Min.Y || x >= bounds.Max.X || y >= bounds.Max.Y {
					continue
				}
				if x >= r.Min.X && x < r.Max.X && y >= r.Min.Y && y < r.Max.Y {
					continue
				}
				c := img.At(x, y)
				rd, g, b, _ := c.RGBA()
				samples = append(samples, [3]int{int(rd >> 8), int(g >> 8), int(b >> 8)})
			}
		}
	}
	outer := image.Rect(
		clampMin(bounds.Min.X, r.Min.X-ring),
		clampMin(bounds.Min.Y, r.Min.Y-ring),
		clampMax(bounds.Max.X, r.Max.X+ring),
		clampMax(bounds.Max.Y, r.Max.Y+ring),
	)
	collect(outer.Min.X, outer.Min.Y, outer.Max.X, r.Min.Y)
	collect(outer.Min.X, r.Max.Y, outer.Max.X, outer.Max.Y)
	collect(outer.Min.X, r.Min.Y, r.Min.X, r.Max.Y)
	collect(r.Max.X, r.Min.Y, outer.Max.X, r.Max.Y)

	var fill color.RGBA
	if len(samples) == 0 {
		fill = color.RGBA{240, 220, 220, 255}
	} else {
		var rs, gs, bs []int
		for _, s := range samples {
			rs = append(rs, s[0])
			gs = append(gs, s[1])
			bs = append(bs, s[2])
		}
		fill = color.RGBA{
			R: uint8(medianInt(rs)),
			G: uint8(medianInt(gs)),
			B: uint8(medianInt(bs)),
			A: 255,
		}
	}
	w, h := r.Dx(), r.Dy()
	patch := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			patch.Set(x, y, fill)
		}
	}
	return patch
}

func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int(nil), vals...)
	sort.Ints(cp)
	return cp[len(cp)/2]
}

func clampMin(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func clampMax(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *OfficeWatermarkService) removeVideoWatermark(
	file *multipart.FileHeader,
	presetID, method string,
	cx, cy, cw, ch float64,
) ([]byte, string, string, error) {
	ff := ffmpegBin()
	if ff == "" {
		return nil, "", "", errors.New("视频去水印需要 FFmpeg")
	}
	workDir, err := os.MkdirTemp(OfficeToolsTempRoot(), "wm-*")
	if err != nil {
		return nil, "", "", err
	}
	defer os.RemoveAll(workDir)

	inPath := filepath.Join(workDir, filepath.Base(file.Filename))
	if err := saveUpload(file, inPath); err != nil {
		return nil, "", "", err
	}

	// 用 ffprobe 获取分辨率
	probe := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0:s=x", inPath)
	outProbe, err := probe.CombinedOutput()
	imgW, imgH := 1080, 1920
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(outProbe)), "x")
		if len(parts) == 2 {
			imgW, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			imgH, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		}
	}
	if imgW <= 0 {
		imgW = 1080
	}
	if imgH <= 0 {
		imgH = 1920
	}

	region := resolveRegion(imgW, imgH, presetID, "delogo", cx, cy, cw, ch)
	// show=0 减轻 delogo 残影；band 略加大
	filter := fmt.Sprintf("delogo=x=%d:y=%d:w=%d:h=%d:show=0", region.x, region.y, region.w, region.h)

	ext := filepath.Ext(file.Filename)
	base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
	outPath := filepath.Join(workDir, base+"_nowm.mp4")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-vf", filter, "-c:a", "copy", outPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		cmd2 := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-vf", filter, "-c:v", "libx264", "-preset", "fast", "-c:a", "aac", outPath)
		if out2, err2 := cmd2.CombinedOutput(); err2 != nil {
			return nil, "", "", fmt.Errorf("视频去水印失败: %v %s", err2, string(out2)+string(out))
		}
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, "", "", err
	}
	return data, base + "_nowm.mp4", "video/mp4", nil
}
