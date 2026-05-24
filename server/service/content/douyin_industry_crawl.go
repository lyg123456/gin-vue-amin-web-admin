package content

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// DouyinOfficialCategory 抖音 Web 垂类（搜索筛选项 category_id）
type DouyinOfficialCategory struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Keyword  string   `json:"keyword"`
	Matchers []string `json:"-"` // 推荐流兜底时标题/话题匹配用
}

// OfficialDouyinCategories 官方垂类列表
var OfficialDouyinCategories = []DouyinOfficialCategory{
	{ID: 601, Name: "剧情", Keyword: "剧情", Matchers: []string{"剧情", "短剧", "电视剧"}},
	{ID: 602, Name: "明星", Keyword: "明星", Matchers: []string{"明星", "艺人", "偶像"}},
	{ID: 603, Name: "综艺", Keyword: "综艺", Matchers: []string{"综艺", "节目", "真人秀"}},
	{ID: 604, Name: "电影", Keyword: "电影", Matchers: []string{"电影", "影院", "影评"}},
	{ID: 605, Name: "电视剧", Keyword: "电视剧", Matchers: []string{"电视剧", "剧集", "追剧"}},
	{ID: 606, Name: "音乐", Keyword: "音乐", Matchers: []string{"音乐", "歌曲", "唱歌"}},
	{ID: 607, Name: "二次元", Keyword: "二次元", Matchers: []string{"二次元", "动漫", "cos"}},
	{ID: 608, Name: "游戏", Keyword: "游戏", Matchers: []string{"游戏", "电竞", "王者"}},
	{ID: 609, Name: "社会时政", Keyword: "社会", Matchers: []string{"社会", "新闻", "时政"}},
	{ID: 612, Name: "舞蹈", Keyword: "舞蹈", Matchers: []string{"舞蹈", "跳舞", "编舞"}},
	{ID: 613, Name: "财经", Keyword: "财经", Matchers: []string{"财经", "金融", "股票", "理财", "经济"}},
	{ID: 615, Name: "科技", Keyword: "科技", Matchers: []string{"科技", "数码", "AI", "手机"}},
	{ID: 617, Name: "母婴", Keyword: "母婴", Matchers: []string{"母婴", "育儿", "宝宝"}},
	{ID: 619, Name: "生活家居", Keyword: "生活", Matchers: []string{"生活", "家居", "日常", "vlog"}},
	{ID: 628, Name: "美食", Keyword: "美食", Matchers: []string{"美食", "吃货", "做饭", "餐厅", "菜谱"}},
	{ID: 629, Name: "旅行", Keyword: "旅行", Matchers: []string{"旅行", "旅游", "探险", "户外", "自驾"}},
	{ID: 631, Name: "时尚", Keyword: "时尚", Matchers: []string{"时尚", "穿搭", "美妆"}},
	{ID: 633, Name: "体育", Keyword: "体育", Matchers: []string{"体育", "足球", "篮球", "运动"}},
	{ID: 635, Name: "汽车", Keyword: "汽车", Matchers: []string{"汽车", "车友", "试驾"}},
}

// DouyinCrawlVideo 单条视频
type DouyinCrawlVideo struct {
	AwemeID       string `json:"awemeId"`
	Title         string `json:"title"`
	AuthorName    string `json:"authorName"`
	CoverURL      string `json:"coverUrl"`
	VideoURL      string `json:"videoUrl"`    // 视频页
	PageURL       string `json:"pageUrl"`     // 同 VideoURL
	DownloadURL   string `json:"downloadUrl"` // 播放/下载直链（有时效）
	PlayCount     int64  `json:"playCount"`
	DiggCount     int64  `json:"diggCount"`
	EffectiveStat int64  `json:"effectiveStat"`
	StatSource    string `json:"statSource"` // play | digg
	Source        string `json:"source"`     // search | feed
}

// DouyinCategoryCrawlResult 单垂类结果
type DouyinCategoryCrawlResult struct {
	CategoryID   int                `json:"categoryId"`
	CategoryName string             `json:"categoryName"`
	Videos       []DouyinCrawlVideo `json:"videos"`
	Source       string             `json:"source,omitempty"`
	Error        string             `json:"error,omitempty"`
}

// DouyinIndustryCrawlReq 抓取请求
type DouyinIndustryCrawlReq struct {
	Cookie       string `json:"cookie"`
	CategoryIDs  []int  `json:"categoryIds"`
	MinPlayCount int64  `json:"minPlayCount"`
	LimitPerCat  int    `json:"limitPerCat"`
	// metric: play=仅播放量 | digg=仅点赞 | auto=有播放量用播放否则用点赞（推荐）
	Metric string `json:"metric"`
}

// DouyinCookieVerifyResult Cookie 检测结果
type DouyinCookieVerifyResult struct {
	OK                 bool   `json:"ok"`
	HasSession         bool   `json:"hasSession"`
	FeedOK             bool   `json:"feedOk"`
	SearchOK           bool   `json:"searchOk"`
	SearchNeedVerify   bool   `json:"searchNeedVerify"`
	Message            string `json:"message"`
	SampleFeedCount    int    `json:"sampleFeedCount"`
}

type douyinSession struct {
	cookie   string
	msToken  string
	csrf     string
	client   *http.Client
}

type OfficeDouyinCrawlService struct{}

func (s *OfficeDouyinCrawlService) ListCategories() []DouyinOfficialCategory {
	out := make([]DouyinOfficialCategory, len(OfficialDouyinCategories))
	copy(out, OfficialDouyinCategories)
	return out
}

func (s *OfficeDouyinCrawlService) VerifyCookie(cookieRaw string) DouyinCookieVerifyResult {
	sess, err := newDouyinSession(cookieRaw)
	if err != nil {
		return DouyinCookieVerifyResult{Message: err.Error()}
	}
	res := DouyinCookieVerifyResult{
		HasSession: sess.hasLogin(),
	}
	if !res.HasSession {
		res.Message = "Cookie 中缺少 sessionid，请重新登录抖音后复制完整 Cookie"
		return res
	}
	if err := sess.warmUp(); err != nil {
		res.Message = "预热失败: " + err.Error()
		return res
	}
	feedBody, err := sess.fetchFeed("", 5)
	if err != nil {
		res.Message = "推荐流不可用: " + err.Error()
		return res
	}
	feedItems := parseDouyinAwemeList(feedBody)
	res.SampleFeedCount = len(feedItems)
	res.FeedOK = len(feedItems) > 0

	searchBody, nilType, err := sess.searchPage(OfficialDouyinCategories[14], 0, 5, true)
	if err != nil {
		res.Message = err.Error()
		return res
	}
	if nilType == "verify_check" {
		res.SearchNeedVerify = true
	}
	searchItems := parseDouyinSearchItems(searchBody)
	res.SearchOK = len(searchItems) > 0

	res.OK = res.FeedOK
	if res.OK {
		res.Message = "Cookie 有效，推荐流可抓取"
		if res.SearchNeedVerify && !res.SearchOK {
			res.Message += "；搜索需先在浏览器完成一次人机验证（打开搜索页滑动验证）后将自动走推荐流匹配垂类"
		} else if res.SearchOK {
			res.Message += "；垂类搜索可用"
		}
	}
	return res
}

func (s *OfficeDouyinCrawlService) CrawlByCookie(req DouyinIndustryCrawlReq) ([]DouyinCategoryCrawlResult, string, error) {
	sess, err := newDouyinSession(req.Cookie)
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
	if err := sess.warmUp(); err != nil {
		return nil, "", fmt.Errorf("预热抖音失败: %w", err)
	}

	globalNote := ""
	searchBlocked := false
	results := make([]DouyinCategoryCrawlResult, 0, len(cats))

	for i, cat := range cats {
		if i > 0 {
			time.Sleep(350 * time.Millisecond)
		}
		videos, source, err := s.crawlOneCategory(sess, cat, minStat, limit, metric, &searchBlocked)
		row := DouyinCategoryCrawlResult{
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
	if searchBlocked && globalNote == "" {
		globalNote = "搜索触发人机验证，已用推荐流+关键词匹配垂类（建议在浏览器打开一次抖音搜索并完成验证后重试）"
	}
	return results, globalNote, nil
}

func (s *OfficeDouyinCrawlService) crawlOneCategory(
	sess *douyinSession,
	cat DouyinOfficialCategory,
	minStat int64,
	limit int,
	metric string,
	searchBlocked *bool,
) ([]DouyinCrawlVideo, string, error) {
	if videos, err := sess.searchCategoryVideos(cat, minStat, limit, metric); err == nil && len(videos) > 0 {
		return videos, "search", nil
	} else if err != nil && strings.Contains(err.Error(), "verify_check") {
		*searchBlocked = true
	} else if err != nil && !strings.Contains(err.Error(), "未找到") {
		// 非空结果错误继续走兜底
		if !strings.Contains(err.Error(), "人机验证") {
			_ = err
		}
	}
	videos, err := sess.feedMatchCategory(cat, minStat, limit, metric)
	if err != nil {
		return nil, "", err
	}
	if len(videos) == 0 {
		return nil, "", fmt.Errorf("未找到热度≥%d 的视频（可降低阈值或更新 Cookie）", minStat)
	}
	return videos, "feed", nil
}

func (s *OfficeDouyinCrawlService) resolveCategories(ids []int) []DouyinOfficialCategory {
	if len(ids) == 0 {
		return nil
	}
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var out []DouyinOfficialCategory
	for _, c := range OfficialDouyinCategories {
		if _, ok := idSet[c.ID]; ok {
			out = append(out, c)
		}
	}
	return out
}

func newDouyinSession(cookieRaw string) (*douyinSession, error) {
	cookie := normalizeDouyinCookie(cookieRaw)
	if cookie == "" {
		return nil, fmt.Errorf("请粘贴 Cookie")
	}
	s := &douyinSession{
		cookie: cookie,
		client: &http.Client{Timeout: 50 * time.Second},
	}
	s.msToken = extractCookieValue(cookie, "msToken")
	s.csrf = extractCookieValue(cookie, "passport_csrf_token")
	if s.csrf == "" {
		s.csrf = extractCookieValue(cookie, "passport_csrf_token_default")
	}
	return s, nil
}

func (s *douyinSession) hasLogin() bool {
	return extractCookieValue(s.cookie, "sessionid") != ""
}

func (s *douyinSession) warmUp() error {
	req, err := http.NewRequest(http.MethodGet, "https://www.douyin.com/jingxuan", nil)
	if err != nil {
		return err
	}
	s.setHeaders(req, "https://www.douyin.com/")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *douyinSession) searchCategoryVideos(cat DouyinOfficialCategory, minStat int64, limit int, metric string) ([]DouyinCrawlVideo, error) {
	const pageSize = 20
	var collected []DouyinCrawlVideo
	seen := make(map[string]struct{})

	for offset := 0; offset < 120 && len(collected) < limit*4; offset += pageSize {
		body, nilType, err := s.searchPage(cat, offset, pageSize, true)
		if err != nil {
			return nil, err
		}
		if nilType == "verify_check" {
			return nil, fmt.Errorf("verify_check")
		}
		items := enrichDouyinVideos(parseDouyinSearchItems(body), metric, "search")
		if len(items) == 0 {
			break
		}
		for _, v := range items {
			if v.EffectiveStat < minStat {
				continue
			}
			if _, ok := seen[v.AwemeID]; ok {
				continue
			}
			seen[v.AwemeID] = struct{}{}
			collected = append(collected, v)
		}
		if len(items) < pageSize {
			break
		}
		time.Sleep(280 * time.Millisecond)
	}
	return trimTopVideos(collected, limit), nil
}

func (s *douyinSession) searchPage(cat DouyinOfficialCategory, offset, count int, withFilter bool) ([]byte, string, error) {
	filter := map[string]string{
		"type":            "1",
		"publish_time":    "0",
		"content_type":    "1",
		"sort_type":       "2",
		"filter_duration": "",
		"search_range":    "0",
		"category_id":     strconv.Itoa(cat.ID),
	}
	filterJSON, _ := json.Marshal(filter)

	q := url.Values{}
	q.Set("device_platform", "webapp")
	q.Set("aid", "6383")
	q.Set("channel", "channel_pc_web")
	q.Set("search_channel", "aweme_video_web")
	q.Set("keyword", cat.Keyword)
	q.Set("search_source", "normal_search")
	q.Set("query_correct_type", "1")
	q.Set("offset", strconv.Itoa(offset))
	q.Set("count", strconv.Itoa(count))
	if withFilter {
		q.Set("is_filter_search", "1")
		q.Set("filter_selected", string(filterJSON))
		q.Set("sort_type", "2")
		q.Set("publish_time", "0")
	} else {
		q.Set("sort_type", "1")
	}
	if s.msToken != "" {
		q.Set("msToken", s.msToken)
	}

	u := "https://www.douyin.com/aweme/v1/web/general/search/single/?" + q.Encode()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return nil, "", err
	}
	ref := fmt.Sprintf("https://www.douyin.com/search/%s?type=video", url.PathEscape(cat.Keyword))
	s.setHeaders(req, ref)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("搜索 HTTP %d", resp.StatusCode)
	}

	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, "", fmt.Errorf("解析搜索响应失败")
	}
	if code, _ := root["status_code"].(float64); code != 0 {
		msg, _ := root["status_msg"].(string)
		if msg == "" {
			msg = fmt.Sprintf("status_code=%.0f", code)
		}
		if code == 2483 || strings.Contains(msg, "登录") {
			return nil, "", fmt.Errorf("%s（请重新登录并复制 Cookie）", msg)
		}
		return nil, "", fmt.Errorf(msg)
	}
	nilType := ""
	if info, ok := root["search_nil_info"].(map[string]interface{}); ok {
		nilType, _ = info["search_nil_type"].(string)
	}
	return body, nilType, nil
}

func (s *douyinSession) feedMatchCategory(cat DouyinOfficialCategory, minStat int64, limit int, metric string) ([]DouyinCrawlVideo, error) {
	var pool []DouyinCrawlVideo
	seen := make(map[string]struct{})
	maxCursor := ""

	for page := 0; page < 10 && len(pool) < limit*8; page++ {
		body, err := s.fetchFeed(maxCursor, 20)
		if err != nil {
			return nil, err
		}
		items := enrichDouyinVideos(parseDouyinAwemeList(body), metric, "feed")
		if len(items) == 0 {
			break
		}
		for _, v := range items {
			if !matchCategory(cat, v.Title) {
				continue
			}
			if v.EffectiveStat < minStat {
				continue
			}
			if _, ok := seen[v.AwemeID]; ok {
				continue
			}
			seen[v.AwemeID] = struct{}{}
			pool = append(pool, v)
		}
		next := extractJSONString(body, "max_cursor")
		if next == "" || next == maxCursor {
			break
		}
		maxCursor = next
		time.Sleep(300 * time.Millisecond)
	}
	return trimTopVideos(pool, limit), nil
}

func (s *douyinSession) fetchFeed(maxCursor string, count int) ([]byte, error) {
	q := url.Values{}
	q.Set("device_platform", "webapp")
	q.Set("aid", "6383")
	q.Set("channel", "channel_pc_web")
	q.Set("count", strconv.Itoa(count))
	if maxCursor != "" {
		q.Set("max_cursor", maxCursor)
	}
	if s.msToken != "" {
		q.Set("msToken", s.msToken)
	}
	u := "https://www.douyin.com/aweme/v1/web/tab/feed/?" + q.Encode()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	s.setHeaders(req, "https://www.douyin.com/")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 12<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("推荐流 HTTP %d", resp.StatusCode)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("解析推荐流失败")
	}
	if code, _ := root["status_code"].(float64); code != 0 {
		msg, _ := root["status_msg"].(string)
		if msg == "" {
			msg = "推荐流异常"
		}
		return nil, fmt.Errorf(msg)
	}
	if len(parseDouyinAwemeList(body)) == 0 {
		return nil, fmt.Errorf("推荐流为空，请更新 Cookie")
	}
	return body, nil
}

func (s *douyinSession) setHeaders(req *http.Request, referer string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://www.douyin.com")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Cookie", s.cookie)
	if s.csrf != "" {
		req.Header.Set("x-secsdk-csrf-token", s.csrf)
	}
}

func normalizeDouyinCookie(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "cookie:")
	raw = strings.TrimPrefix(raw, "Cookie:")
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	return strings.TrimSpace(raw)
}

func extractCookieValue(cookie, key string) string {
	prefix := key + "="
	for _, part := range strings.Split(cookie, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, prefix) {
			return strings.TrimPrefix(part, prefix)
		}
	}
	return ""
}

func matchCategory(cat DouyinOfficialCategory, title string) bool {
	title = strings.ToLower(title)
	for _, m := range cat.Matchers {
		if m == "" {
			continue
		}
		if strings.Contains(title, strings.ToLower(m)) {
			return true
		}
	}
	return strings.Contains(title, strings.ToLower(cat.Keyword))
}

func enrichDouyinVideos(list []DouyinCrawlVideo, metric, source string) []DouyinCrawlVideo {
	out := make([]DouyinCrawlVideo, 0, len(list))
	for _, v := range list {
		v.Source = source
		v.EffectiveStat, v.StatSource = pickMetric(v.PlayCount, v.DiggCount, metric)
		out = append(out, v)
	}
	return out
}

func pickMetric(play, digg int64, metric string) (int64, string) {
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

func trimTopVideos(list []DouyinCrawlVideo, limit int) []DouyinCrawlVideo {
	sort.Slice(list, func(i, j int) bool {
		return list[i].EffectiveStat > list[j].EffectiveStat
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

func parseDouyinSearchItems(body []byte) []DouyinCrawlVideo {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return nil
	}
	var rows []interface{}
	switch data := root["data"].(type) {
	case []interface{}:
		rows = data
	case map[string]interface{}:
		if list, ok := data["list"].([]interface{}); ok {
			rows = list
		}
	}
	if len(rows) == 0 {
		if list, ok := root["aweme_list"].([]interface{}); ok {
			rows = list
		}
	}
	var out []DouyinCrawlVideo
	for _, row := range rows {
		m, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		aweme, ok := m["aweme_info"].(map[string]interface{})
		if !ok {
			aweme = m
		}
		if v := mapDouyinAweme(aweme); v != nil {
			out = append(out, *v)
		}
	}
	return out
}

func parseDouyinAwemeList(body []byte) []DouyinCrawlVideo {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return nil
	}
	list, ok := root["aweme_list"].([]interface{})
	if !ok {
		return nil
	}
	var out []DouyinCrawlVideo
	for _, row := range list {
		m, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		if v := mapDouyinAweme(m); v != nil {
			out = append(out, *v)
		}
	}
	return out
}

func mapDouyinAweme(aweme map[string]interface{}) *DouyinCrawlVideo {
	id := stringifyID(aweme["aweme_id"])
	if id == "" {
		return nil
	}
	stats, _ := aweme["statistics"].(map[string]interface{})
	play := int64FromAny(stats["play_count"])
	digg := int64FromAny(stats["digg_count"])
	title, _ := aweme["desc"].(string)
	if title == "" {
		title, _ = aweme["title"].(string)
	}
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > 120 {
		title = string([]rune(title)[:120]) + "…"
	}
	authorName := ""
	if author, ok := aweme["author"].(map[string]interface{}); ok {
		authorName, _ = author["nickname"].(string)
	}
	cover := firstURLFromAwemeCover(aweme)
	pageURL := fmt.Sprintf("https://www.douyin.com/video/%s", id)
	downloadURL := firstDouyinDownloadURL(aweme)
	return &DouyinCrawlVideo{
		AwemeID:     id,
		Title:       title,
		AuthorName:  authorName,
		CoverURL:    cover,
		VideoURL:    pageURL,
		PageURL:     pageURL,
		DownloadURL: downloadURL,
		PlayCount:   play,
		DiggCount:   digg,
	}
}

func firstDouyinDownloadURL(aweme map[string]interface{}) string {
	video, ok := aweme["video"].(map[string]interface{})
	if !ok {
		return ""
	}
	for _, key := range []string{"download_addr", "play_addr"} {
		if addr, ok := video[key].(map[string]interface{}); ok {
			if u := urlFromCoverObj(addr); u != "" {
				return u
			}
		}
	}
	if bitRate, ok := video["bit_rate"].([]interface{}); ok {
		var best string
		var bestBr int64
		for _, row := range bitRate {
			m, ok := row.(map[string]interface{})
			if !ok {
				continue
			}
			u := urlFromCoverObj(m["play_addr"])
			if u == "" {
				continue
			}
			br := int64FromAny(m["bit_rate"])
			if br >= bestBr {
				bestBr = br
				best = u
			}
		}
		if best != "" {
			return best
		}
	}
	return ""
}

func extractJSONString(body []byte, key string) string {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return ""
	}
	v, ok := root[key]
	if !ok {
		return ""
	}
	return stringifyID(v)
}

func firstURLFromAwemeCover(aweme map[string]interface{}) string {
	video, ok := aweme["video"].(map[string]interface{})
	if !ok {
		return ""
	}
	for _, key := range []string{"cover", "origin_cover", "dynamic_cover"} {
		if u := urlFromCoverObj(video[key]); u != "" {
			return u
		}
	}
	return ""
}

func urlFromCoverObj(v interface{}) string {
	m, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	list, ok := m["url_list"].([]interface{})
	if !ok || len(list) == 0 {
		return ""
	}
	s, _ := list[0].(string)
	return s
}

func stringifyID(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(v)
	}
}

func int64FromAny(v interface{}) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	default:
		return 0
	}
}
