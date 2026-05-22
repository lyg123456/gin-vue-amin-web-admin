package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type OfficeWebCrawlService struct{}

type crawlProduct struct {
	Name       string
	Price      string
	Image      string
	Seller     string
	SourceURL  string
}

// CrawlProductsExcel 抓取页面商品信息导出 Excel
func (s *OfficeWebCrawlService) CrawlProductsExcel(pageURL string) ([]byte, string, error) {
	rawHTML, finalURL, err := fetchOfficeWebPage(pageURL)
	if err != nil {
		return nil, "", err
	}
	base, _ := url.Parse(finalURL)
	products := extractProducts(rawHTML, base)
	if len(products) == 0 {
		return nil, "", fmt.Errorf("未识别到商品信息：页面可能为纯 JS 渲染，请换用含结构化数据或静态 HTML 的商品列表页")
	}
	if len(products) > 200 {
		products = products[:200]
	}
	return productsToExcel(products, finalURL)
}

func extractProducts(html string, base *url.URL) []crawlProduct {
	var list []crawlProduct
	seen := map[string]bool{}
	add := func(p crawlProduct) {
		key := p.Name + "|" + p.Price + "|" + p.Image
		if p.Name == "" || seen[key] {
			return
		}
		seen[key] = true
		if base != nil && p.Image != "" {
			p.Image = resolveOfficeURL(base, p.Image)
		}
		if p.SourceURL == "" && base != nil {
			p.SourceURL = base.String()
		}
		list = append(list, p)
	}

	// JSON-LD Product / ItemList
	for _, block := range extractJSONLDBlocks(html) {
		var generic map[string]interface{}
		if json.Unmarshal([]byte(block), &generic) != nil {
			continue
		}
		collectJSONLDProducts(generic, add)
	}

	// 常见电商块：data-price、class 含 product
	productBlockRe := regexp.MustCompile(`(?is)<(?:div|li|article)[^>]*(?:class|id)=["'][^"']*(?:product|item|goods)[^"']*["'][^>]*>(.*?)</(?:div|li|article)>`)
	priceRe := regexp.MustCompile(`(?i)(?:¥|￥|price|data-price)[^>]*>?\s*([0-9][0-9.,]*)`)
	imgRe := regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
	nameRe := regexp.MustCompile(`(?is)<h[1-4][^>]*>([^<]{2,120})</h[1-4]>`)
	sellerRe := regexp.MustCompile(`(?i)(?:店铺|商家|厂家|品牌|seller|brand)[：:\s]*([^<\n]{2,40})`)

	for _, block := range productBlockRe.FindAllStringSubmatch(html, 30) {
		chunk := block[1]
		names := nameRe.FindAllStringSubmatch(chunk, 1)
		if len(names) == 0 {
			continue
		}
		p := crawlProduct{Name: cleanText(names[0][1])}
		if pm := priceRe.FindStringSubmatch(chunk); len(pm) > 1 {
			p.Price = cleanText(pm[1])
		}
		if im := imgRe.FindStringSubmatch(chunk); len(im) > 1 {
			p.Image = cleanText(im[1])
		}
		if sm := sellerRe.FindStringSubmatch(chunk); len(sm) > 1 {
			p.Seller = cleanText(sm[1])
		}
		add(p)
	}

	// og:title + 页面级价格兜底
	if len(list) == 0 {
		ogTitle := extractMeta(html, `property=["']og:title["']\s+content=["']([^"']+)`)
		ogImage := extractMeta(html, `property=["']og:image["']\s+content=["']([^"']+)`)
		if ogTitle != "" {
			p := crawlProduct{Name: ogTitle, Image: ogImage}
			if pm := priceRe.FindStringSubmatch(html); len(pm) > 1 {
				p.Price = pm[1]
			}
			add(p)
		}
	}
	return list
}

func extractJSONLDBlocks(html string) []string {
	re := regexp.MustCompile(`(?is)<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := re.FindAllStringSubmatch(html, -1)
	var blocks []string
	for _, m := range matches {
		if len(m) > 1 {
			blocks = append(blocks, strings.TrimSpace(m[1]))
		}
	}
	return blocks
}

func collectJSONLDProducts(node map[string]interface{}, add func(crawlProduct)) {
	if node == nil {
		return
	}
	t, _ := node["@type"].(string)
	types := strings.ToLower(t)
	if strings.Contains(types, "product") {
		p := crawlProduct{}
		if v, ok := node["name"].(string); ok {
			p.Name = v
		}
		if offers, ok := node["offers"].(map[string]interface{}); ok {
			if pr, ok := offers["price"].(string); ok {
				p.Price = pr
			}
			if pr, ok := offers["price"].(float64); ok {
				p.Price = fmt.Sprintf("%.2f", pr)
			}
			if b, ok := offers["seller"].(map[string]interface{}); ok {
				if n, ok := b["name"].(string); ok {
					p.Seller = n
				}
			}
		}
		if img, ok := node["image"].(string); ok {
			p.Image = img
		}
		if imgs, ok := node["image"].([]interface{}); ok && len(imgs) > 0 {
			if s, ok := imgs[0].(string); ok {
				p.Image = s
			}
		}
		if br, ok := node["brand"].(map[string]interface{}); ok {
			if n, ok := br["name"].(string); ok {
				p.Seller = n
			}
		}
		add(p)
	}
	if strings.Contains(types, "itemlist") {
		if items, ok := node["itemListElement"].([]interface{}); ok {
			for _, it := range items {
				if m, ok := it.(map[string]interface{}); ok {
					if item, ok := m["item"].(map[string]interface{}); ok {
						collectJSONLDProducts(item, add)
					} else {
						collectJSONLDProducts(m, add)
					}
				}
			}
		}
	}
	if g, ok := node["@graph"].([]interface{}); ok {
		for _, it := range g {
			if m, ok := it.(map[string]interface{}); ok {
				collectJSONLDProducts(m, add)
			}
		}
	}
}

func resolveOfficeURL(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(u).String()
}

func cleanText(s string) string {
	s = html.UnescapeString(s)
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func productsToExcel(products []crawlProduct, source string) ([]byte, string, error) {
	f := excelize.NewFile()
	sheet := "商品列表"
	_ = f.SetSheetName("Sheet1", sheet)
	headers := []string{"序号", "商品名称", "价格", "图片链接", "厂家/商家", "来源网址"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for i, p := range products {
		row := i + 2
		vals := []interface{}{i + 1, p.Name, p.Price, p.Image, p.Seller, source}
		for c, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(c+1, row)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	_ = f.SetColWidth(sheet, "B", "B", 36)
	_ = f.SetColWidth(sheet, "D", "D", 48)
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "product-crawl-" + time.Now().Format("20060102-150405") + ".xlsx", nil
}
