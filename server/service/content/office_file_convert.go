package content

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/jung-kurt/gofpdf"
)

type OfficeFileConvertService struct{}

type ConvertCapabilities struct {
	ImageConvert      bool     `json:"imageConvert"`
	TextToPdf         bool     `json:"textToPdf"`
	LibreOffice       bool     `json:"libreOffice"`
	LibreOfficePath   string   `json:"libreOfficePath"`
	GoOfficeFallback  bool     `json:"goOfficeFallback"`
	GoFallbackExts      []string `json:"goFallbackExts"`
	OfficePdfAvailable bool    `json:"officePdfAvailable"`
	LibreOfficeHint   string   `json:"libreOfficeHint"`
	SupportedTargets  []string `json:"supportedTargets"`
	OfficeToPdfExts   []string `json:"officeToPdfExts"`
	MaxUploadMB       int      `json:"maxUploadMB"`
}

func (s *OfficeFileConvertService) Capabilities() ConvertCapabilities {
	lo := s.libreOfficeBin()
	exts := []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".rtf"}
	goExts := []string{".docx", ".xlsx", ".pptx"}
	targets := []string{"png", "jpeg", "pdf"}
	hint := ""
	if lo == "" {
		hint = "未检测到 LibreOffice。Linux: sudo apt install libreoffice 或 scripts/install-office-tools-linux.sh，配置 libreoffice-path: /usr/bin/soffice；" +
			"Windows: 安装后配置 C:\\Program Files\\LibreOffice\\program\\soffice.exe；" +
			"当前可用 Go 简易转换 .docx/.xlsx（版式简化）。"
	} else {
		hint = "已启用 LibreOffice 高保真转换（推荐）。"
	}
	return ConvertCapabilities{
		ImageConvert:       true,
		TextToPdf:          true,
		LibreOffice:        lo != "",
		LibreOfficePath:    lo,
		GoOfficeFallback:   true,
		GoFallbackExts:     goExts,
		OfficePdfAvailable: lo != "" || true,
		LibreOfficeHint:    hint,
		SupportedTargets:   targets,
		OfficeToPdfExts:    exts,
		MaxUploadMB:        s.maxUploadMB(),
	}
}

func (s *OfficeFileConvertService) maxUploadMB() int {
	mb := global.GVA_CONFIG.OfficeTools.MaxUploadMB
	if mb <= 0 {
		return 20
	}
	return mb
}

func (s *OfficeFileConvertService) maxBytes() int64 {
	return int64(s.maxUploadMB()) * 1024 * 1024
}

func (s *OfficeFileConvertService) tempBase() string {
	return OfficeToolsTempRoot()
}

// Convert 返回转换后文件字节、下载文件名、Content-Type
func (s *OfficeFileConvertService) Convert(file *multipart.FileHeader, target, textContent string) ([]byte, string, string, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		target = "pdf"
	}

	if file == nil {
		textContent = strings.TrimSpace(textContent)
		if textContent == "" {
			return nil, "", "", errors.New("请上传文件或填写文本内容")
		}
		if target != "pdf" {
			return nil, "", "", errors.New("纯文本仅支持导出为 PDF")
		}
		return s.textToPDF(textContent)
	}

	if file.Size > s.maxBytes() {
		return nil, "", "", fmt.Errorf("文件超过 %dMB 限制", s.maxUploadMB())
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch {
	case isImageExt(ext):
		if target == "pdf" {
			return s.imageToPDF(file)
		}
		return s.convertImage(file, target)
	case isOfficeExt(ext) && target == "pdf":
		return s.convertOfficeToPDF(file)
	case ext == ".pdf" && target == "pdf":
		return s.readThrough(file, ".pdf")
	default:
		return nil, "", "", fmt.Errorf("不支持从 %s 转为 %s", ext, target)
	}
}

func (s *OfficeFileConvertService) readThrough(file *multipart.FileHeader, ext string) ([]byte, string, string, error) {
	b, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	name := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)) + ext
	return b, name, mimeByExt(ext), nil
}

func readUploadFile(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, 50<<20))
}

func (s *OfficeFileConvertService) convertImage(file *multipart.FileHeader, target string) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	img, err := imaging.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", fmt.Errorf("无法解码图片: %w", err)
	}
	var buf bytes.Buffer
	var ext, mime string
	switch target {
	case "png":
		err = imaging.Encode(&buf, img, imaging.PNG)
		ext, mime = ".png", "image/png"
	case "jpeg", "jpg":
		err = imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(92))
		ext, mime = ".jpg", "image/jpeg"
	default:
		return nil, "", "", fmt.Errorf("图片目标格式仅支持 png、jpeg")
	}
	if err != nil {
		return nil, "", "", err
	}
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	return buf.Bytes(), base + ext, mime, nil
}

func (s *OfficeFileConvertService) imageToPDF(file *multipart.FileHeader) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	pdfW := float64(w) * 72 / 96
	pdfH := float64(h) * 72 / 96
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "pt",
		Size:    gofpdf.SizeType{Wd: pdfW, Ht: pdfH},
	})
	pdf.AddPage()
	var buf bytes.Buffer
	switch strings.ToLower(filepath.Ext(file.Filename)) {
	case ".png":
		_ = png.Encode(&buf, img)
		pdf.RegisterImageOptionsReader("img", gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(buf.Bytes()))
	default:
		buf.Reset()
		_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 92})
		pdf.RegisterImageOptionsReader("img", gofpdf.ImageOptions{ImageType: "JPG"}, bytes.NewReader(buf.Bytes()))
	}
	pdf.ImageOptions("img", 0, 0, pdfW, pdfH, false, gofpdf.ImageOptions{}, 0, "")
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, "", "", err
	}
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	return out.Bytes(), base + ".pdf", "application/pdf", nil
}

func (s *OfficeFileConvertService) convertOfficeToPDF(file *multipart.FileHeader) ([]byte, string, string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if lo := s.libreOfficeBin(); lo != "" {
		return s.convertOfficeWithLibreOffice(file, lo)
	}
	switch ext {
	case ".docx":
		return s.convertDocxToPDFGo(file)
	case ".xlsx":
		return s.convertXlsxToPDFGo(file)
	case ".pptx":
		return s.convertPptxToPDFGo(file)
	case ".xls":
		return nil, "", "", errors.New(".xls 旧格式需安装 LibreOffice；请另存为 .xlsx 或安装 LibreOffice 后重试")
	default:
		return nil, "", "", fmt.Errorf(
			"%s 需 LibreOffice 高保真转换：请在本机安装 LibreOffice 并在 config.yaml 设置 office-tools.libreoffice-path（Windows 示例: C:\\Program Files\\LibreOffice\\program\\soffice.exe）",
			ext,
		)
	}
}

func saveUpload(file *multipart.FileHeader, dest string) error {
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, f)
	return err
}

func (s *OfficeFileConvertService) textToPDF(text string) ([]byte, string, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 11)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		pdf.MultiCell(0, 7, tr(line), "", "", false)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), "export.pdf", "application/pdf", nil
}

func isImageExt(ext string) bool {
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}

func isOfficeExt(ext string) bool {
	switch ext {
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp", ".rtf":
		return true
	}
	return false
}

func mimeByExt(ext string) string {
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
