package content

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	xhsAPIHost     = "https://edith.xiaohongshu.com"
	xhsDefaultUA   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	xhsSearchNotes = "/api/sns/web/v1/search/notes"
	xhsFeedNote    = "/api/sns/web/v1/feed"
	xhsVideoCDN    = "https://sns-video-qc.xhscdn.com"
)

// XhsOfficialCategory 小红书内容垂类（搜索关键词）
type XhsOfficialCategory struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Keyword  string   `json:"keyword"`
	Matchers []string `json:"-"`
}

// OfficialXhsCategories 垂类列表
var OfficialXhsCategories = []XhsOfficialCategory{
	{ID: 1, Name: "美食", Keyword: "美食", Matchers: []string{"美食", "吃货", "探店", "菜谱"}},
	{ID: 2, Name: "旅行", Keyword: "旅行", Matchers: []string{"旅行", "旅游", "自驾", "户外"}},
	{ID: 3, Name: "美妆", Keyword: "美妆", Matchers: []string{"美妆", "护肤", "化妆", "彩妆"}},
	{ID: 4, Name: "穿搭", Keyword: "穿搭", Matchers: []string{"穿搭", "时尚", "ootd"}},
	{ID: 5, Name: "家居", Keyword: "家居", Matchers: []string{"家居", "装修", "收纳", "家装"}},
	{ID: 6, Name: "母婴", Keyword: "母婴", Matchers: []string{"母婴", "育儿", "亲子", "宝宝"}},
	{ID: 7, Name: "健身", Keyword: "健身", Matchers: []string{"健身", "运动", "减脂", "瑜伽"}},
	{ID: 8, Name: "教育", Keyword: "教育", Matchers: []string{"教育", "学习", "考研", "干货"}},
	{ID: 9, Name: "科技", Keyword: "科技", Matchers: []string{"科技", "数码", "AI", "手机"}},
	{ID: 10, Name: "职场", Keyword: "职场", Matchers: []string{"职场", "办公", "打工人"}},
	{ID: 11, Name: "情感", Keyword: "情感", Matchers: []string{"情感", "恋爱", "心理"}},
	{ID: 12, Name: "游戏", Keyword: "游戏", Matchers: []string{"游戏", "电竞", "王者"}},
	{ID: 13, Name: "汽车", Keyword: "汽车", Matchers: []string{"汽车", "车友", "新能源"}},
	{ID: 14, Name: "宠物", Keyword: "宠物", Matchers: []string{"宠物", "猫", "狗"}},
	{ID: 15, Name: "影视", Keyword: "影视", Matchers: []string{"影视", "电影", "综艺", "追剧"}},
	{ID: 16, Name: "音乐", Keyword: "音乐", Matchers: []string{"音乐", "唱歌", "翻唱"}},
	{ID: 17, Name: "摄影", Keyword: "摄影", Matchers: []string{"摄影", "拍照", "vlog"}},
	{ID: 18, Name: "搞笑", Keyword: "搞笑", Matchers: []string{"搞笑", "段子", "幽默"}},
}

// XhsCrawlVideo 单条笔记（视频类）
type XhsCrawlVideo struct {
	NoteID        string `json:"noteId"`
	Title         string `json:"title"`
	AuthorName    string `json:"authorName"`
	CoverURL      string `json:"coverUrl"`
	VideoURL      string `json:"videoUrl"`      // 笔记页（explore）
	PageURL       string `json:"pageUrl"`       // 同 VideoURL，便于导出
	DownloadURL   string `json:"downloadUrl"`   // MP4 直链（便于下载）
	PlayCount     int64  `json:"playCount"`
	DiggCount     int64  `json:"diggCount"`
	EffectiveStat int64  `json:"effectiveStat"`
	StatSource    string `json:"statSource"`
	Source        string `json:"source"`
	XsecToken     string `json:"-"`
}

// XhsCategoryCrawlResult 单垂类结果
type XhsCategoryCrawlResult struct {
	CategoryID   int             `json:"categoryId"`
	CategoryName string          `json:"categoryName"`
	Videos       []XhsCrawlVideo `json:"videos"`
	Source       string          `json:"source,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// XhsIndustryCrawlReq 抓取请求
type XhsIndustryCrawlReq struct {
	Cookie       string `json:"cookie"`
	CategoryIDs  []int  `json:"categoryIds"`
	MinPlayCount int64  `json:"minPlayCount"`
	LimitPerCat  int    `json:"limitPerCat"`
	Metric       string `json:"metric"`
}

// XhsCookieVerifyResult Cookie 检测
type XhsCookieVerifyResult struct {
	OK       bool   `json:"ok"`
	HasA1    bool   `json:"hasA1"`
	Message  string `json:"message"`
	SampleOK bool   `json:"sampleSearchOk"`
}

type xhsSession struct {
	cookie string
	client *http.Client
}

type OfficeXhsCrawlService struct{}

func (s *OfficeXhsCrawlService) ListCategories() []XhsOfficialCategory {
	out := make([]XhsOfficialCategory, len(OfficialXhsCategories))
	copy(out, OfficialXhsCategories)
	return out
}

func (s *OfficeXhsCrawlService) VerifyCookie(cookieRaw string) XhsCookieVerifyResult {
	sess, err := newXhsSession(cookieRaw)
	if err != nil {
		return XhsCookieVerifyResult{Message: err.Error()}
	}
	res := XhsCookieVerifyResult{HasA1: sess.hasA1()}
	if !res.HasA1 {
		res.Message = "Cookie 无效：请登录 www.xiaohongshu.com 后复制完整 Cookie（需含 a1）"
		return res
	}
	_, err = sess.searchNotes("美食", 1, 5, true)
	if err != nil {
		res.Message = "Cookie 含 a1，但接口不可用: " + err.Error()
		return res
	}
	res.OK = true
	res.SampleOK = true
	res.Message = "Cookie 有效，可按垂类搜索爆款视频笔记"
	return res
}

func (s *OfficeXhsCrawlService) CrawlByCookie(req XhsIndustryCrawlReq) ([]XhsCategoryCrawlResult, string, error) {
	sess, err := newXhsSession(req.Cookie)
	if err != nil {
		return nil, "", err
	}
	minStat := req.MinPlayCount
	if minStat <= 0 {
		minStat = 20000
	}
	limit := req.LimitPerCat
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	metric := strings.ToLower(strings.TrimSpace(req.Metric))
	if metric == "" {
		metric = "auto"
	}
	cats := s.resolveCategories(req.CategoryIDs)
	if len(cats) == 0 {
		return nil, "", fmt.Errorf("请至少选择一个垂类")
	}
	if len(cats) > 12 {
		return nil, "", fmt.Errorf("单次最多 12 个垂类")
	}

	note := "数据来源：小红书 Web 搜索；含笔记页与 MP4 直链（直链有时效）。若出现 HTTP 461，请在浏览器完成人机验证后重试。需 Python3 + pip install xhshow"
	results := make([]XhsCategoryCrawlResult, 0, len(cats))
	for i, cat := range cats {
		if i > 0 {
			time.Sleep(400 * time.Millisecond)
		}
		videos, source, err := s.crawlCategory(sess, cat, minStat, limit, metric)
		row := XhsCategoryCrawlResult{
			CategoryID:   cat.ID,
			CategoryName: cat.Name,
			Videos:       videos,
			Source:       source,
		}
		if err != nil {
			row.Error = err.Error()
		}
		results = append(results, row)
	}
	return results, note, nil
}

func (s *OfficeXhsCrawlService) resolveCategories(ids []int) []XhsOfficialCategory {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	var out []XhsOfficialCategory
	for _, c := range OfficialXhsCategories {
		if _, ok := set[c.ID]; ok {
			out = append(out, c)
		}
	}
	return out
}

func (s *OfficeXhsCrawlService) crawlCategory(
	sess *xhsSession,
	cat XhsOfficialCategory,
	minStat int64,
	limit int,
	metric string,
) ([]XhsCrawlVideo, string, error) {
	var pool []XhsCrawlVideo
	seen := make(map[string]struct{})
	keywords := xhsCategoryKeywords(cat)

	for _, kw := range keywords {
		for page := 1; page <= 4; page++ {
			items, err := sess.searchNotes(kw, page, 20, true)
			if err != nil {
				return nil, "", err
			}
			for _, raw := range items {
				v := mapXhsSearchItem(raw)
				if v == nil {
					continue
				}
				if !matchXhsCategory(cat, v.Title) {
					continue
				}
				v.EffectiveStat, v.StatSource = pickXhsMetric(v.PlayCount, v.DiggCount, metric)
				if v.EffectiveStat < minStat {
					continue
				}
				if _, ok := seen[v.NoteID]; ok {
					continue
				}
				seen[v.NoteID] = struct{}{}
				v.Source = "search"
				pool = append(pool, *v)
			}
			if len(pool) >= limit*4 {
				break
			}
			if len(items) < 10 {
				break
			}
			time.Sleep(300 * time.Millisecond)
		}
		if len(pool) >= limit*3 {
			break
		}
	}

	out := trimXhsTop(pool, limit)
	if len(out) == 0 {
		return nil, "", fmt.Errorf("未找到点赞/播放≥%d 的视频笔记，可降低阈值或换垂类", minStat)
	}
	enrichXhsDownloadURLs(sess, out)
	return out, "search", nil
}

func xhsCategoryKeywords(cat XhsOfficialCategory) []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(cat.Keyword)
	for i, m := range cat.Matchers {
		if i >= 3 {
			break
		}
		add(m)
	}
	return out
}

func newXhsSession(cookieRaw string) (*xhsSession, error) {
	cookie := normalizeXhsCookie(cookieRaw)
	if cookie == "" {
		return nil, fmt.Errorf("请粘贴 Cookie")
	}
	return &xhsSession{
		cookie: cookie,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func normalizeXhsCookie(raw string) string {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "cookie") {
		if i := strings.Index(raw, ":"); i >= 0 {
			raw = strings.TrimSpace(raw[i+1:])
		}
	}
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	raw = strings.ToValidUTF8(raw, "")
	return strings.TrimSpace(raw)
}

func (s *xhsSession) hasA1() bool {
	return extractCookieValue(s.cookie, "a1") != ""
}

func (s *xhsSession) searchNotes(keyword string, page, pageSize int, videoOnly bool) ([]map[string]interface{}, error) {
	noteType := 0
	if videoOnly {
		noteType = 1
	}
	payload := map[string]interface{}{
		"keyword":    keyword,
		"page":       page,
		"page_size":  pageSize,
		"search_id":  xhsNewSearchID(),
		"sort":       "popularity_descending",
		"note_type":  noteType,
	}
	body, err := s.postSigned(xhsSearchNotes, payload)
	if err != nil {
		return nil, err
	}
	return parseXhsSearchItems(body), nil
}

func (s *xhsSession) postSigned(uri string, payload map[string]interface{}) ([]byte, error) {
	headers, err := xhsSignRequest(xhsSignBridgeReq{
		Method:  "POST",
		URI:     uri,
		Cookie:  s.cookie,
		Payload: payload,
	})
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, xhsAPIHost+uri, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", xhsDefaultUA)
	req.Header.Set("Origin", "https://www.xiaohongshu.com")
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")
	req.Header.Set("Cookie", s.cookie)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if isXhsRiskHTTPStatus(resp.StatusCode) {
		return nil, fmt.Errorf("%s", xhsRiskHint(resp.StatusCode, resp))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateStr(string(body), 200))
	}
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return body, nil
	}
	if code := int64FromAny(root["code"]); code != 0 {
		msg, _ := root["msg"].(string)
		if msg == "" {
			msg, _ = root["message"].(string)
		}
		if strings.Contains(strings.ToLower(msg), "sign") {
			return nil, fmt.Errorf("签名失效，请确认已 pip install xhshow 且 Cookie 未过期: %s", msg)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	if success, ok := root["success"].(bool); ok && !success {
		return nil, fmt.Errorf("请求被拒绝: %s", truncateStr(string(body), 120))
	}
	if data, ok := root["data"].(map[string]interface{}); ok {
		if xhsDataLooksEmpty(data) && uri == xhsFeedNote {
			return nil, fmt.Errorf("%s", xhsRiskHint(461, resp))
		}
		b, _ := json.Marshal(data)
		return b, nil
	}
	return body, nil
}

func isXhsRiskHTTPStatus(code int) bool {
	return code == 461 || code == 471
}

func xhsRiskHint(code int, resp *http.Response) string {
	msg := fmt.Sprintf(
		"小红书触发人机验证（HTTP %d）：请用同一账号在浏览器打开 www.xiaohongshu.com，随便搜索一次或完成滑块验证，再重新复制 Cookie 并重试",
		code,
	)
	if resp == nil {
		return msg
	}
	for _, key := range []string{"Verify", "verify", "Xy-Verify-Id", "xy-verify-id"} {
		if v := strings.TrimSpace(resp.Header.Get(key)); v != "" {
			return msg + "（需验证）"
		}
	}
	return msg
}

func xhsDataLooksEmpty(data map[string]interface{}) bool {
	if len(data) == 0 {
		return true
	}
	items, _ := data["items"].([]interface{})
	return len(items) == 0
}

func parseXhsSearchItems(body []byte) []map[string]interface{} {
	var data map[string]interface{}
	if json.Unmarshal(body, &data) != nil {
		return nil
	}
	var out []map[string]interface{}
	for _, key := range []string{"items", "notes", "note_list"} {
		arr, ok := data[key].([]interface{})
		if !ok {
			continue
		}
		for _, it := range arr {
			if m, ok := it.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func mapXhsSearchItem(raw map[string]interface{}) *XhsCrawlVideo {
	noteID := stringifyID(raw["id"])
	if noteID == "" {
		noteID = stringifyID(raw["note_id"])
	}
	xsec, _ := raw["xsec_token"].(string)
	card, _ := raw["note_card"].(map[string]interface{})
	if xsec == "" && card != nil {
		xsec, _ = card["xsec_token"].(string)
	}
	if card == nil {
		card = raw
	}
	if noteID == "" {
		noteID = stringifyID(card["note_id"])
	}
	if noteID == "" {
		return nil
	}
	noteType, _ := card["type"].(string)
	if noteType != "" && noteType != "video" {
		return nil
	}
	title, _ := card["display_title"].(string)
	if title == "" {
		title, _ = card["title"].(string)
	}
	title = trimXhsTitle(title)
	author := ""
	if user, ok := card["user"].(map[string]interface{}); ok {
		author, _ = user["nickname"].(string)
		if author == "" {
			author, _ = user["nick_name"].(string)
		}
	}
	cover := ""
	if coverMap, ok := card["cover"].(map[string]interface{}); ok {
		cover, _ = coverMap["url_default"].(string)
		if cover == "" {
			cover, _ = coverMap["url"].(string)
		}
	}
	play, digg := int64(0), int64(0)
	if info, ok := card["interact_info"].(map[string]interface{}); ok {
		digg = parseXhsCountString(info["liked_count"])
		play = parseXhsCountString(info["view_count"])
		if play == 0 {
			play = parseXhsCountString(info["play_count"])
		}
	}
	pageURL := "https://www.xiaohongshu.com/explore/" + noteID
	if xsec != "" {
		pageURL += "?xsec_token=" + xsec + "&xsec_source=pc_search"
	}
	downloadURL := extractXhsVideoURLFromCard(card)
	return &XhsCrawlVideo{
		NoteID:      noteID,
		Title:       title,
		AuthorName:  author,
		CoverURL:    cover,
		VideoURL:    pageURL,
		PageURL:     pageURL,
		DownloadURL: downloadURL,
		PlayCount:   play,
		DiggCount:   digg,
		XsecToken:   xsec,
	}
}

func (s *xhsSession) fetchNoteCard(noteID, xsecToken string) (map[string]interface{}, error) {
	if strings.TrimSpace(xsecToken) == "" {
		return nil, fmt.Errorf("缺少 xsec_token")
	}
	payload := map[string]interface{}{
		"source_note_id": noteID,
		"image_formats":  []string{"jpg", "webp", "avif"},
		"extra":          map[string]interface{}{"need_body_topic": "1"},
		"xsec_source":    "pc_search",
		"xsec_token":     xsecToken,
	}
	body, err := s.postSigned(xhsFeedNote, payload)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	if json.Unmarshal(body, &data) != nil {
		return nil, fmt.Errorf("解析笔记详情失败")
	}
	items, _ := data["items"].([]interface{})
	if len(items) == 0 {
		return nil, fmt.Errorf("笔记详情为空")
	}
	first, ok := items[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("笔记详情格式异常")
	}
	if card, ok := first["note_card"].(map[string]interface{}); ok {
		return card, nil
	}
	return first, nil
}

func enrichXhsDownloadURLs(sess *xhsSession, videos []XhsCrawlVideo) {
	for i := range videos {
		if videos[i].DownloadURL != "" {
			continue
		}
		// 优先 HTML 解析，避免 /feed 接口频繁触发 461 风控
		if u := sess.fetchDownloadURLFromHTML(videos[i].NoteID, videos[i].XsecToken); u != "" {
			videos[i].DownloadURL = u
			sleepBetweenXhsNoteFetch(i, len(videos))
			continue
		}
		if videos[i].XsecToken == "" {
			continue
		}
		card, err := sess.fetchNoteCard(videos[i].NoteID, videos[i].XsecToken)
		if err != nil {
			continue
		}
		videos[i].DownloadURL = extractXhsVideoURLFromCard(card)
		sleepBetweenXhsNoteFetch(i, len(videos))
	}
}

func sleepBetweenXhsNoteFetch(i, total int) {
	if i < total-1 {
		time.Sleep(450 * time.Millisecond)
	}
}

var xhsHTMLVideoKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`"originVideoKey"\s*:\s*"([^"]+)"`),
	regexp.MustCompile(`"origin_video_key"\s*:\s*"([^"]+)"`),
	regexp.MustCompile(`"masterUrl"\s*:\s*"([^"]+)"`),
	regexp.MustCompile(`"master_url"\s*:\s*"([^"]+)"`),
}

func (s *xhsSession) fetchDownloadURLFromHTML(noteID, xsecToken string) string {
	if strings.TrimSpace(noteID) == "" {
		return ""
	}
	page := "https://www.xiaohongshu.com/explore/" + noteID
	if strings.TrimSpace(xsecToken) != "" {
		page += "?xsec_token=" + url.QueryEscape(xsecToken) + "&xsec_source=pc_search"
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, page, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", xhsDefaultUA)
	req.Header.Set("Cookie", s.cookie)
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")
	resp, err := s.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}
	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return ""
	}
	html := string(htmlBytes)
	for _, re := range xhsHTMLVideoKeyPatterns {
		m := re.FindStringSubmatch(html)
		if len(m) < 2 {
			continue
		}
		key := strings.TrimSpace(m[1])
		if key == "" {
			continue
		}
		if u := normalizeMediaURL(key); u != "" {
			return u
		}
		return buildXhsCDNURL(key)
	}
	if u := extractXhsVideoURLFromInitialState(html); u != "" {
		return u
	}
	return ""
}

func extractXhsVideoURLFromInitialState(html string) string {
	re := regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*(\{.*?\})\s*</script>`)
	m := re.FindStringSubmatch(html)
	if len(m) < 2 {
		return ""
	}
	raw := strings.ReplaceAll(m[1], "undefined", `""`)
	var state map[string]interface{}
	if json.Unmarshal([]byte(raw), &state) != nil {
		return ""
	}
	noteRoot, _ := state["note"].(map[string]interface{})
	if noteRoot == nil {
		return ""
	}
	detailMap, _ := noteRoot["noteDetailMap"].(map[string]interface{})
	for _, v := range detailMap {
		detail, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		note, _ := detail["note"].(map[string]interface{})
		if u := extractXhsVideoURLFromCard(note); u != "" {
			return u
		}
	}
	return ""
}

func extractXhsVideoURLFromCard(card map[string]interface{}) string {
	if card == nil {
		return ""
	}
	video, ok := card["video"].(map[string]interface{})
	if !ok {
		return ""
	}
	// 优先带签名的 stream URL（可直接播放）；origin_video_key 裸链易 403
	if u := extractXhsStreamURL(video); u != "" {
		return u
	}
	if consumer, ok := video["consumer"].(map[string]interface{}); ok {
		if key, _ := consumer["origin_video_key"].(string); strings.TrimSpace(key) != "" {
			return buildXhsCDNURL(key)
		}
	}
	return firstNonEmptyString(
		nestedString(video, "url"),
		nestedString(video, "media", "url"),
	)
}

func extractXhsStreamURL(video map[string]interface{}) string {
	media, ok := video["media"].(map[string]interface{})
	if !ok {
		return ""
	}
	stream, ok := media["stream"].(map[string]interface{})
	if !ok {
		return ""
	}
	var bestURL string
	var bestSize int64
	for _, codec := range []string{"h264", "h265", "av1"} {
		arr, ok := stream[codec].([]interface{})
		if !ok {
			continue
		}
		for _, row := range arr {
			m, ok := row.(map[string]interface{})
			if !ok {
				continue
			}
			for _, key := range []string{"master_url", "backup_urls", "backup_url", "url"} {
				if key == "backup_urls" {
					if list, ok := m[key].([]interface{}); ok {
						for _, item := range list {
							raw, ok := item.(string)
							if !ok {
								continue
							}
							if u := normalizeMediaURL(raw); u != "" {
								return u
							}
						}
					}
					continue
				}
				raw, _ := m[key].(string)
				if u := normalizeMediaURL(raw); u != "" {
					size := int64FromAny(m["size"])
					if size >= bestSize {
						bestSize = size
						bestURL = u
					}
				}
			}
		}
	}
	return bestURL
}

func buildXhsCDNURL(key string) string {
	key = strings.TrimPrefix(strings.TrimSpace(key), "/")
	if key == "" {
		return ""
	}
	return xhsVideoCDN + "/" + key
}

func normalizeMediaURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.ReplaceAll(u, `\u002F`, `/`)
	u = strings.ReplaceAll(u, `\/`, `/`)
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	return ""
}

func nestedString(m map[string]interface{}, keys ...string) string {
	cur := interface{}(m)
	for _, k := range keys {
		obj, ok := cur.(map[string]interface{})
		if !ok {
			return ""
		}
		cur = obj[k]
	}
	s, _ := cur.(string)
	return strings.TrimSpace(s)
}

func firstNonEmptyString(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseXhsCountString(v interface{}) int64 {
	switch t := v.(type) {
	case string:
		return parseXhsCountText(t)
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	default:
		return 0
	}
}

func parseXhsCountText(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	mult := float64(1)
	if strings.HasSuffix(s, "万") || strings.HasSuffix(s, "w") || strings.HasSuffix(s, "W") {
		mult = 10000
		s = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(s, "万"), "w"), "W")
	} else if strings.HasSuffix(s, "k") || strings.HasSuffix(s, "K") {
		mult = 1000
		s = strings.TrimSuffix(strings.TrimSuffix(s, "k"), "K")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(f * mult)
}

func pickXhsMetric(play, digg int64, metric string) (int64, string) {
	switch metric {
	case "play":
		return play, "play"
	case "digg":
		return digg, "digg"
	default:
		if play > 0 {
			return play, "play"
		}
		return digg, "digg"
	}
}

func matchXhsCategory(cat XhsOfficialCategory, title string) bool {
	title = strings.ToLower(title)
	for _, m := range cat.Matchers {
		if strings.Contains(title, strings.ToLower(m)) {
			return true
		}
	}
	return strings.Contains(title, strings.ToLower(cat.Keyword))
}

func trimXhsTop(list []XhsCrawlVideo, limit int) []XhsCrawlVideo {
	sort.Slice(list, func(i, j int) bool {
		return list[i].EffectiveStat > list[j].EffectiveStat
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func trimXhsTitle(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > 120 {
		return string([]rune(s)[:120]) + "…"
	}
	return s
}

func xhsNewSearchID() string {
	b := make([]byte, 11)
	_, _ = rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))[:22]
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
