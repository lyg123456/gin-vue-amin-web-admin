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

type OfficeMediaService struct{}

type MediaCapabilities struct {
	FFmpeg      bool   `json:"ffmpeg"`
	FFmpegPath  string `json:"ffmpegPath"`
	ImageMerge  bool   `json:"imageMerge"`
	MaxUploadMB int    `json:"maxUploadMB"`
	Hint        string `json:"hint"`
}

func (s *OfficeMediaService) maxUploadMB() int {
	mb := global.GVA_CONFIG.OfficeTools.MaxUploadMB
	if mb <= 0 {
		return 50
	}
	return mb
}

func (s *OfficeMediaService) Capabilities() MediaCapabilities {
	ff := ffmpegBin()
	hint := "未检测到 FFmpeg：Linux 执行 apt install ffmpeg 或 scripts/install-office-tools-linux.sh；Windows 运行 install-ffmpeg.ps1；或配置 office-tools.ffmpeg-path"
	if ff != "" {
		hint = "FFmpeg 已就绪：" + ff
	}
	return MediaCapabilities{
		FFmpeg:      ff != "",
		FFmpegPath:  ff,
		ImageMerge:  true,
		MaxUploadMB: s.maxUploadMB(),
		Hint:        hint,
	}
}

func (s *OfficeMediaService) tempBase() string {
	return OfficeToolsTempRoot()
}

func (s *OfficeMediaService) ProcessVideo(action string, file *multipart.FileHeader) ([]byte, string, string, error) {
	ff := ffmpegBin()
	if ff == "" {
		return nil, "", "", errors.New("未检测到 FFmpeg：请运行 scripts/install-ffmpeg.ps1 或在 config.yaml 设置 office-tools.ffmpeg-path")
	}
	if file == nil {
		return nil, "", "", errors.New("请上传视频文件")
	}
	workDir, err := os.MkdirTemp(s.tempBase(), "media-*")
	if err != nil {
		return nil, "", "", err
	}
	defer os.RemoveAll(workDir)

	inPath := filepath.Join(workDir, filepath.Base(file.Filename))
	if err := saveUpload(file, inPath); err != nil {
		return nil, "", "", err
	}
	base := strings.TrimSuffix(filepath.Base(file.Filename), filepath.Ext(file.Filename))

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	switch action {
	case "extract_audio":
		outPath := filepath.Join(workDir, base+".mp3")
		cmd := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-vn", "-acodec", "libmp3lame", "-q:a", "2", outPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, "", "", fmt.Errorf("提取音频失败: %v %s", err, string(out))
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			return nil, "", "", err
		}
		return data, base + ".mp3", "audio/mpeg", nil
	case "remove_audio":
		outPath := filepath.Join(workDir, base+"_silent.mp4")
		cmd := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-c:v", "copy", "-an", outPath)
		if _, err := cmd.CombinedOutput(); err != nil {
			cmd2 := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-c:v", "libx264", "-an", outPath)
			if out2, err2 := cmd2.CombinedOutput(); err2 != nil {
				return nil, "", "", fmt.Errorf("去除音频失败: %v %s", err2, string(out2))
			}
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			return nil, "", "", err
		}
		return data, base + "_silent.mp4", "video/mp4", nil
	case "extract_frame":
		outPath := filepath.Join(workDir, base+"_frame.jpg")
		cmd := exec.CommandContext(ctx, ff, "-y", "-i", inPath, "-vframes", "1", "-q:v", "2", outPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, "", "", fmt.Errorf("提取画面失败: %v %s", err, string(out))
		}
		data, err := os.ReadFile(outPath)
		if err != nil {
			return nil, "", "", err
		}
		return data, base + "_frame.jpg", "image/jpeg", nil
	default:
		return nil, "", "", fmt.Errorf("不支持的操作: %s", action)
	}
}

// CompositeImages 前景图叠加到背景图
func (s *OfficeMediaService) CompositeImages(bgFile, fgFile *multipart.FileHeader, posX, posY, scalePct int) ([]byte, string, string, error) {
	if bgFile == nil || fgFile == nil {
		return nil, "", "", errors.New("请上传背景图与前景图")
	}
	bgRaw, err := readUploadFile(bgFile)
	if err != nil {
		return nil, "", "", err
	}
	fgRaw, err := readUploadFile(fgFile)
	if err != nil {
		return nil, "", "", err
	}
	bgImg, err := imaging.Decode(bytes.NewReader(bgRaw))
	if err != nil {
		return nil, "", "", fmt.Errorf("背景图解码失败: %w", err)
	}
	fgImg, err := imaging.Decode(bytes.NewReader(fgRaw))
	if err != nil {
		return nil, "", "", fmt.Errorf("前景图解码失败: %w", err)
	}
	if scalePct <= 0 {
		scalePct = 100
	}
	if scalePct != 100 {
		w := fgImg.Bounds().Dx() * scalePct / 100
		h := fgImg.Bounds().Dy() * scalePct / 100
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		fgImg = imaging.Resize(fgImg, w, h, imaging.Lanczos)
	}
	result := imaging.Clone(bgImg)
	result = imaging.Overlay(result, fgImg, image.Pt(posX, posY), 1.0)
	var buf bytes.Buffer
	ext := strings.ToLower(filepath.Ext(bgFile.Filename))
	if ext == ".png" {
		_ = png.Encode(&buf, result)
		base := strings.TrimSuffix(bgFile.Filename, ext)
		return buf.Bytes(), base + "_merged.png", "image/png", nil
	}
	_ = jpeg.Encode(&buf, result, &jpeg.Options{Quality: 92})
	base := strings.TrimSuffix(bgFile.Filename, filepath.Ext(bgFile.Filename))
	return buf.Bytes(), base + "_merged.jpg", "image/jpeg", nil
}

// ExtractImageBackground 简易提取：取四角主色铺底作为背景层（演示用，复杂抠图请用专业工具）
func (s *OfficeMediaService) ExtractImageBackground(file *multipart.FileHeader) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	img, err := imaging.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", err
	}
	b := img.Bounds()
	// 采样四角平均色生成纯色背景图
	corners := []image.Point{{b.Min.X, b.Min.Y}, {b.Max.X - 1, b.Min.Y}, {b.Min.X, b.Max.Y - 1}, {b.Max.X - 1, b.Max.Y - 1}}
	var r, g, bl, n int
	for _, p := range corners {
		c := img.At(p.X, p.Y)
		cr, cg, cb, _ := c.RGBA()
		r += int(cr >> 8)
		g += int(cg >> 8)
		bl += int(cb >> 8)
		n++
	}
	bgColor := color.NRGBA{
		R: uint8(r / n),
		G: uint8(g / n),
		B: uint8(bl / n),
		A: 255,
	}
	bg := imaging.New(b.Dx(), b.Dy(), bgColor)
	var buf bytes.Buffer
	_ = png.Encode(&buf, bg)
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	return buf.Bytes(), base + "_background.png", "image/png", nil
}

func parseMediaInt(form string, def int) int {
	if form == "" {
		return def
	}
	v, err := strconv.Atoi(form)
	if err != nil {
		return def
	}
	return v
}
