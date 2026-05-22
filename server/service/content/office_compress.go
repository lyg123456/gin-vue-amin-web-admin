package content

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"image/jpeg"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
)

type OfficeCompressService struct{}

type CompressResult struct {
	Data         []byte `json:"-"`
	Filename     string `json:"filename"`
	Mime         string `json:"mime"`
	OriginalSize int64  `json:"originalSize"`
	NewSize      int64  `json:"newSize"`
	Ratio        string `json:"ratio"`
}

func (s *OfficeCompressService) CompressImage(file *multipart.FileHeader, quality, maxWidth int) (*CompressResult, error) {
	if file == nil {
		return nil, fmt.Errorf("请上传图片")
	}
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	if maxWidth <= 0 {
		maxWidth = 1920
	}
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, err
	}
	img, err := imaging.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("图片解码失败: %w", err)
	}
	if img.Bounds().Dx() > maxWidth {
		img = imaging.Resize(img, maxWidth, 0, imaging.Lanczos)
	}
	var buf bytes.Buffer
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mime := "image/jpeg"
	outExt := ".jpg"
	if ext == ".png" && quality >= 90 {
		// 高画质保留 png
		if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
			return nil, err
		}
		mime = "image/png"
		outExt = ".png"
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	}
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	ratio := fmt.Sprintf("%.1f%%", float64(buf.Len())/float64(len(raw))*100)
	if len(raw) == 0 {
		ratio = "-"
	}
	return &CompressResult{
		Data:         buf.Bytes(),
		Filename:     base + "_compressed" + outExt,
		Mime:         mime,
		OriginalSize: int64(len(raw)),
		NewSize:      int64(buf.Len()),
		Ratio:        ratio,
	}, nil
}

// CompressExcel 通过仅保留单元格数据重建 xlsx，并 ZIP 最高压缩
func (s *OfficeCompressService) CompressExcel(file *multipart.FileHeader) (*CompressResult, error) {
	if file == nil {
		return nil, fmt.Errorf("请上传 Excel 文件")
	}
	raw, err := readUploadFile(file)
	if err != nil {
		return nil, err
	}
	f, err := excelize.OpenReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("无法读取 Excel: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel 无工作表")
	}
	out := excelize.NewFile()
	_ = out.SetSheetName(out.GetSheetName(0), sheets[0])
	for i, sheet := range sheets {
		target := sheet
		if i > 0 {
			_, _ = out.NewSheet(target)
		}
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		for r, row := range rows {
			for c, cell := range row {
				cellName, _ := excelize.CoordinatesToCellName(c+1, r+1)
				_ = out.SetCellValue(target, cellName, cell)
			}
		}
	}
	var buf bytes.Buffer
	if err := out.Write(&buf); err != nil {
		return nil, err
	}
	// 二次 zip 压缩
	compressed, err := rezipXLSX(buf.Bytes())
	if err != nil {
		compressed = buf.Bytes()
	}
	base := strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))
	ratio := fmt.Sprintf("%.1f%%", float64(len(compressed))/float64(len(raw))*100)
	return &CompressResult{
		Data:         compressed,
		Filename:     base + "_compressed.xlsx",
		Mime:         "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		OriginalSize: int64(len(raw)),
		NewSize:      int64(len(compressed)),
		Ratio:        ratio,
	}, nil
}

func rezipXLSX(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(rc)
		_ = rc.Close()
		hdr := &zip.FileHeader{Name: f.Name, Method: zip.Deflate}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			continue
		}
		_, _ = w.Write(body)
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
