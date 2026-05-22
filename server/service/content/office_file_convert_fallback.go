package content

import (
	"archive/zip"
	"bytes"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

var docxTextRe = regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
var pptxTextRe = regexp.MustCompile(`<a:t[^>]*>([^<]*)</a:t>`)

func goFallbackOfficeExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".docx", ".xlsx", ".pptx":
		return true
	default:
		return false
	}
}

func extractDocxPlainText(raw []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return "", fmt.Errorf("无法读取 docx: %w", err)
	}
	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rc)
		_ = rc.Close()
		xml := buf.String()
		var b strings.Builder
		for _, m := range docxTextRe.FindAllStringSubmatch(xml, -1) {
			if len(m) > 1 && m[1] != "" {
				b.WriteString(m[1])
				b.WriteByte('\n')
			}
		}
		text := strings.TrimSpace(b.String())
		if text == "" {
			return "", fmt.Errorf("docx 中未提取到文本（可能主要为图片/复杂排版）")
		}
		return text, nil
	}
	return "", fmt.Errorf("无效的 docx 文件")
}

func extractPptxPlainText(raw []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, f := range zr.File {
		if !strings.HasPrefix(f.Name, "ppt/slides/slide") || !strings.HasSuffix(f.Name, ".xml") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rc)
		_ = rc.Close()
		for _, m := range pptxTextRe.FindAllStringSubmatch(buf.String(), -1) {
			if len(m) > 1 {
				b.WriteString(m[1])
				b.WriteByte('\n')
			}
		}
	}
	text := strings.TrimSpace(b.String())
	if text == "" {
		return "", fmt.Errorf("pptx 中未提取到文本")
	}
	return text, nil
}

func (s *OfficeFileConvertService) convertPptxToPDFGo(file *multipart.FileHeader) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	text, err := extractPptxPlainText(raw)
	if err != nil {
		return nil, "", "", err
	}
	header := "【Go 简易转换】仅提取幻灯片文字，版式/图片/动画不会保留。\n\n"
	return s.textToPDF(header + text)
}

func (s *OfficeFileConvertService) convertDocxToPDFGo(file *multipart.FileHeader) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	text, err := extractDocxPlainText(raw)
	if err != nil {
		return nil, "", "", err
	}
	header := "【Go 简易转换】仅提取文字，版式/图片/表格可能与原 Word 不一致。\n\n"
	return s.textToPDF(header + text)
}

func (s *OfficeFileConvertService) convertXlsxToPDFGo(file *multipart.FileHeader) ([]byte, string, string, error) {
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, "", "", err
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", fmt.Errorf("无法读取 Excel: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, "", "", fmt.Errorf("Excel 无工作表")
	}
	var b strings.Builder
	b.WriteString("【Go 简易转换】将表格导出为文本 PDF，复杂格式请安装 LibreOffice 获得高保真 PDF。\n\n")
	for _, sheet := range sheets {
		b.WriteString("=== ")
		b.WriteString(sheet)
		b.WriteString(" ===\n")
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		for _, row := range rows {
			b.WriteString(strings.Join(row, "\t"))
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	data, _, mime, err := s.textToPDF(b.String())
	if err != nil {
		return nil, "", "", err
	}
	return data, base + ".pdf", mime, nil
}
