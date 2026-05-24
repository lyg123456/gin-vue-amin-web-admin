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

const wechatDefaultReferer = "https://channels.weixin.qq.com/post/list"

// WechatOfficialCategory 微信视频号助手内容垂类（发表/标签体系，与平台展示分类一致）
type WechatOfficialCategory struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Keyword  string   `json:"keyword"`
	Matchers []string `json:"-"`
}

// OfficialWechatCategories 官方垂类
var OfficialWechatCategories = []WechatOfficialCategory{
	{ID: 1, Name: "生活", Keyword: "生活", Matchers: []string{"生活", "日常", "vlog", "家居"}},
	{ID: 2, Name: "美食", Keyword: "美食", Matchers: []string{"美食", "吃货", "做饭", "餐厅", "菜谱"}},
	{ID: 3, Name: "旅行", Keyword: "旅行", Matchers: []string{"旅行", "旅游", "探险", "户外", "自驾"}},
	{ID: 4, Name: "时尚", Keyword: "时尚", Matchers: []string{"时尚", "穿搭", "潮流"}},
	{ID: 5, Name: "母婴", Keyword: "母婴", Matchers: []string{"母婴", "育儿", "亲子", "宝宝"}},
	{ID: 6, Name: "教育", Keyword: "教育", Matchers: []string{"教育", "学习", "课程", "知识分享"}},
	{ID: 7, Name: "科技", Keyword: "科技", Matchers: []string{"科技", "数码", "AI", "手机", "互联网"}},
	{ID: 8, Name: "财经", Keyword: "财经", Matchers: []string{"财经", "金融", "理财", "股票", "经济"}},
	{ID: 9, Name: "汽车", Keyword: "汽车", Matchers: []string{"汽车", "车友", "试驾", "新能源车"}},
	{ID: 10, Name: "游戏", Keyword: "游戏", Matchers: []string{"游戏", "电竞", "王者", "原神"}},
	{ID: 11, Name: "音乐", Keyword: "音乐", Matchers: []string{"音乐", "唱歌", "歌曲"}},
	{ID: 12, Name: "知识", Keyword: "知识", Matchers: []string{"知识", "科普", "干货"}},
	{ID: 13, Name: "搞笑", Keyword: "搞笑", Matchers: []string{"搞笑", "幽默", "段子"}},
	{ID: 14, Name: "情感", Keyword: "情感", Matchers: []string{"情感", "恋爱", "婚姻"}},
	{ID: 15, Name: "美妆", Keyword: "美妆", Matchers: []string{"美妆", "护肤", "化妆"}},
	{ID: 16, Name: "运动", Keyword: "运动", Matchers: []string{"运动", "健身", "体育", "篮球", "足球"}},
	{ID: 17, Name: "三农", Keyword: "三农", Matchers: []string{"三农", "农村", "农业", "乡村"}},
	{ID: 18, Name: "职场", Keyword: "职场", Matchers: []string{"职场", "办公", "打工人"}},
	{ID: 19, Name: "影视", Keyword: "影视", Matchers: []string{"影视", "电影", "电视剧", "综艺", "娱乐"}},
	{ID: 20, Name: "时事", Keyword: "时事", Matchers: []string{"时事", "新闻", "热点", "资讯"}},
}

// WechatCrawlVideo 视频号作品
type WechatCrawlVideo struct {
	ExportID      string `json:"exportId"`
	Title         string `json:"title"`
	AuthorName    string `json:"authorName"`
	CoverURL      string `json:"coverUrl"`
	VideoURL      string `json:"videoUrl"`
	PlayCount     int64  `json:"playCount"`
	DiggCount     int64  `json:"diggCount"`
	EffectiveStat int64  `json:"effectiveStat"`
	StatSource    string `json:"statSource"`
	Source        string `json:"source"`
}

// WechatCategoryCrawlResult 单垂类结果
type WechatCategoryCrawlResult struct {
	CategoryID   int                `json:"categoryId"`
	CategoryName string             `json:"categoryName"`
	Videos       []WechatCrawlVideo `json:"videos"`
	Source       string             `json:"source,omitempty"`
	Error        string             `json:"error,omitempty"`
}

// WechatIndustryCrawlReq 抓取请求（字段与抖音工具一致）
type WechatIndustryCrawlReq struct {
	Cookie        string `json:"cookie"`
	WxChannelBase string `json:"wxChannelBase"` // 可选，如 http://127.0.0.1:2026（wx_channel 本地代理）
	CategoryIDs   []int  `json:"categoryIds"`
	MinPlayCount  int64  `json:"minPlayCount"`
	LimitPerCat   int    `json:"limitPerCat"`
	Metric        string `json:"metric"`
}

// WechatCookieVerifyResult Cookie 检测
type WechatCookieVerifyResult struct {
	OK              bool   `json:"ok"`
	HasSession      bool   `json:"hasSession"`
	AssistantOK     bool   `json:"assistantOk"`
	WxChannelReady  bool   `json:"wxChannelReady"`
	Nickname        string `json:"nickname,omitempty"`
	Message         string `json:"message"`
	SamplePostCount int    `json:"samplePostCount"`
}

type wechatSession struct {
	cookie        string
	wxChannelBase string
	client        *http.Client
}

type OfficeWechatCrawlService struct{}

func (s *OfficeWechatCrawlService) ListCategories() []WechatOfficialCategory {
	out := make([]WechatOfficialCategory, len(OfficialWechatCategories))
	copy(out, OfficialWechatCategories)
	return out
}

func (s *OfficeWechatCrawlService) VerifyCookie(cookieRaw string) WechatCookieVerifyResult {
	return s.VerifyCookieWithOptions(cookieRaw, "")
}

func (s *OfficeWechatCrawlService) VerifyCookieWithOptions(cookieRaw, wxChannelBase string) WechatCookieVerifyResult {
	sess, err := newWechatSession(cookieRaw, wxChannelBase)
	if err != nil {
		return WechatCookieVerifyResult{Message: err.Error()}
	}
	res := WechatCookieVerifyResult{HasSession: sess.hasLogin()}
	if sess.wxChannelEnabled() {
		ready, wxMsg := sess.checkWxChannelReady()
		if ready {
			res.WxChannelReady = true
			res.OK = true
			res.Message = wxMsg
			return res
		}
		res.Message = wxMsg
	}
	if !res.HasSession && !sess.wxChannelEnabled() {
		res.Message = "请填写 wx_channel 代理地址，或登录视频号助手后粘贴 Cookie"
		return res
	}
	if !res.HasSession && sess.wxChannelEnabled() {
		return res
	}
	nick, ok, err := sess.verifyAssistant()
	if err != nil {
		res.Message = err.Error()
	}
	res.AssistantOK = ok
	res.Nickname = nick

	items, listErr := sess.fetchPostListPage(1, 5)
	res.SamplePostCount = len(items)
	if listErr == nil && len(items) > 0 {
		res.OK = true
		res.AssistantOK = true
		if nick != "" {
			res.Message = fmt.Sprintf("助手 Cookie 有效（%s）；抓取他人视频请优先使用 wx_channel", nick)
		} else {
			res.Message = "助手 Cookie 有效；抓取他人视频请优先使用 wx_channel"
		}
		return res
	}
	if ok {
		res.Message = "已登录视频号助手，但作品列表为空或未拉取到数据"
		return res
	}
	if err != nil {
		return res
	}
	if listErr != nil {
		res.Message = listErr.Error()
		return res
	}
	res.Message = "Cookie 无法访问视频号助手：请在 channels.weixin.qq.com 登录后，从「内容管理」页复制 Cookie（需含 sessionid、wxuin）"
	return res
}

func (s *OfficeWechatCrawlService) CrawlByCookie(req WechatIndustryCrawlReq) ([]WechatCategoryCrawlResult, string, error) {
	sess, err := newWechatSession(req.Cookie, req.WxChannelBase)
	if err != nil {
		return nil, "", err
	}
	if !sess.wxChannelEnabled() && !sess.hasLogin() {
		return nil, "", fmt.Errorf("请粘贴 Cookie，或配置 wx_channel 本地代理地址")
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

	if sess.wxChannelEnabled() {
		if ready, msg := sess.checkWxChannelReady(); !ready {
			return nil, "", fmt.Errorf(msg)
		}
	} else if _, ok, err := sess.verifyAssistant(); err != nil {
		return nil, "", err
	} else if !ok {
		return nil, "", fmt.Errorf("Cookie 已失效，请重新登录视频号助手，或配置 wx_channel 代理抓取他人视频")
	}

	note := "仅抓取其他创作者公开作品（播放量/点赞来自作品数据）。微信无抖音式全站热榜，需通过「搜索视频号/搜索视频」发现他人内容。"
	if sess.wxChannelEnabled() {
		note += " 当前经 wx_channel 本地代理调用浏览器内视频号搜索 API。"
	} else {
		note += " 助手后台 contact/search 已下线，强烈建议配置 wx_channel（本机 http://127.0.0.1:2026）。"
	}

	results := make([]WechatCategoryCrawlResult, 0, len(cats))
	for i, cat := range cats {
		if i > 0 {
			time.Sleep(400 * time.Millisecond)
		}
		videos, source, err := s.crawlCategory(sess, cat, minStat, limit, metric)
		row := WechatCategoryCrawlResult{
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

func (s *OfficeWechatCrawlService) resolveCategories(ids []int) []WechatOfficialCategory {
	if len(ids) == 0 {
		return nil
	}
	set := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	var out []WechatOfficialCategory
	for _, c := range OfficialWechatCategories {
		if _, ok := set[c.ID]; ok {
			out = append(out, c)
		}
	}
	return out
}

func (s *OfficeWechatCrawlService) crawlCategory(
	sess *wechatSession,
	cat WechatOfficialCategory,
	minStat int64,
	limit int,
	metric string,
) ([]WechatCrawlVideo, string, error) {
	var pool []WechatCrawlVideo
	seen := make(map[string]struct{})
	keywords := categorySearchKeywords(cat)

	for _, kw := range keywords {
		// 优先：按关键词直接搜视频（他人作品）
		if sess.wxChannelEnabled() {
			videos, _ := sess.searchOthersVideosWxChannel(kw, 2, metric)
			for _, v := range videos {
				appendWechatVideoIfMatch(&pool, seen, cat, v, minStat)
			}
		}
		// 搜账号 → 拉取该账号作品列表
		accountVideos, err := sess.searchOthersByAccounts(kw, 12, 6, metric)
		if err != nil && sess.wxChannelEnabled() {
			continue
		}
		for _, v := range accountVideos {
			appendWechatVideoIfMatch(&pool, seen, cat, v, minStat)
		}
		if len(pool) >= limit*3 {
			break
		}
		time.Sleep(350 * time.Millisecond)
	}

	out := trimWechatTop(pool, limit)
	if len(out) == 0 {
		hint := "未找到其他创作者播放/点赞≥%d 的作品，可降低热度下限或更换垂类"
		if !sess.wxChannelEnabled() {
			hint += "；微信已关闭助手后台搜索，请在本机运行 wx_channel 并填写代理地址（默认 http://127.0.0.1:2026），在浏览器打开视频号页面后再抓取"
		}
		return nil, "", fmt.Errorf(hint, minStat)
	}
	return out, "others", nil
}

func categorySearchKeywords(cat WechatOfficialCategory) []string {
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
		if i >= 4 {
			break
		}
		add(m)
	}
	return out
}

func appendWechatVideoIfMatch(pool *[]WechatCrawlVideo, seen map[string]struct{}, cat WechatOfficialCategory, v WechatCrawlVideo, minStat int64) {
	if v.ExportID == "" {
		return
	}
	if !matchWechatCategory(cat, v.Title) {
		return
	}
	if v.EffectiveStat < minStat {
		return
	}
	if _, ok := seen[v.ExportID]; ok {
		return
	}
	seen[v.ExportID] = struct{}{}
	v.Source = "others"
	*pool = append(*pool, v)
}

func newWechatSession(cookieRaw, wxChannelBase string) (*wechatSession, error) {
	cookie := normalizeWechatCookie(cookieRaw)
	wxChannelBase = strings.TrimSpace(wxChannelBase)
	wxChannelBase = strings.TrimRight(wxChannelBase, "/")
	if cookie == "" && wxChannelBase == "" {
		return nil, fmt.Errorf("请粘贴 Cookie 或填写 wx_channel 代理地址")
	}
	return &wechatSession{
		cookie:        cookie,
		wxChannelBase: wxChannelBase,
		client:        &http.Client{Timeout: 90 * time.Second},
	}, nil
}

func (s *wechatSession) wxChannelEnabled() bool {
	return s.wxChannelBase != ""
}

func normalizeWechatCookie(raw string) string {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "cookie") {
		if i := strings.Index(raw, ":"); i >= 0 {
			raw = strings.TrimSpace(raw[i+1:])
		}
	}
	raw = strings.ReplaceAll(raw, "\n", "")
	raw = strings.ReplaceAll(raw, "\r", "")
	return strings.TrimSpace(raw)
}

func (s *wechatSession) hasLogin() bool {
	c := strings.ToLower(s.cookie)
	return strings.Contains(c, "sessionid") ||
		strings.Contains(c, "wxuin") ||
		strings.Contains(c, "finder_") ||
		strings.Contains(c, "ticket=")
}

func (s *wechatSession) verifyAssistant() (nickname string, ok bool, err error) {
	// auth_finder_list 已下线，改用 auth_data（与视频号助手页面一致）
	body, err := s.postForm(
		"mmfinderassistant-bin/auth/auth_data",
		map[string]interface{}{"timestamp": time.Now().UnixMilli()},
		wechatDefaultReferer,
	)
	if err != nil {
		return "", false, err
	}
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return "", false, fmt.Errorf("解析登录状态失败")
	}
	code := int64FromAny(root["errCode"])
	if code != 0 {
		msg, _ := root["errMsg"].(string)
		if code == 300330 {
			return "", false, fmt.Errorf("Cookie 已失效，请在视频号助手重新登录后复制 Cookie")
		}
		return "", false, fmt.Errorf("%s", msg)
	}
	data, _ := root["data"].(map[string]interface{})
	fu, _ := data["finderUser"].(map[string]interface{})
	if fu == nil {
		return "", false, nil
	}
	nickname, _ = fu["nickname"].(string)
	if nickname == "" {
		nickname, _ = fu["finderNickname"].(string)
	}
	return nickname, nickname != "" || fu["feedsCount"] != nil, nil
}

func (s *wechatSession) fetchAllOwnPosts(maxPages int) ([]map[string]interface{}, error) {
	var all []map[string]interface{}
	for page := 1; page <= maxPages; page++ {
		items, err := s.fetchPostListPage(page, 20)
		if err != nil {
			return all, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		if len(items) < 20 {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	return all, nil
}

func (s *wechatSession) fetchPostListPage(page, pageSize int) ([]map[string]interface{}, error) {
	body, err := s.postForm(
		"mmfinderassistant-bin/post/post_list",
		map[string]interface{}{
			"currentPage": page,
			"pageSize":    pageSize,
			"timestamp":   time.Now().UnixMilli(),
		},
		wechatDefaultReferer,
	)
	if err != nil {
		return nil, err
	}
	return parseWechatPostListBody(body)
}

func (s *wechatSession) searchOthersByAccounts(keyword string, maxAccounts, pagesPerAccount int, metric string) ([]WechatCrawlVideo, error) {
	contacts, err := s.searchContacts(keyword, 20)
	if err != nil {
		return nil, err
	}
	if len(contacts) == 0 {
		return nil, nil
	}
	if maxAccounts > len(contacts) {
		maxAccounts = len(contacts)
	}
	var out []WechatCrawlVideo
	for i := 0; i < maxAccounts; i++ {
		username := contactUsername(contacts[i])
		nickname, _ := contacts[i]["nickname"].(string)
		if username == "" {
			continue
		}
		feeds, err := s.fetchContactFeeds(username, pagesPerAccount)
		if err != nil {
			continue
		}
		for _, raw := range feeds {
			v := mapWechatFeedObject(raw, nickname)
			if v == nil {
				continue
			}
			s.enrichFeedStatsWxChannel(v, raw)
			v.Source = "account_feed"
			v.EffectiveStat, v.StatSource = pickWechatMetric(v.PlayCount, v.DiggCount, metric)
			out = append(out, *v)
		}
		time.Sleep(300 * time.Millisecond)
	}
	return out, nil
}

func contactUsername(c map[string]interface{}) string {
	for _, key := range []string{"username", "finderUsername", "finder_username", "encryptUsername"} {
		if u, ok := c[key].(string); ok && u != "" {
			return u
		}
	}
	return ""
}

func (s *wechatSession) searchContacts(keyword string, pageSize int) ([]map[string]interface{}, error) {
	if s.wxChannelEnabled() {
		list, err := s.wxChannelSearchContacts(keyword, 1, pageSize)
		if err == nil && len(list) > 0 {
			return list, nil
		}
		if err != nil {
			return nil, err
		}
	}
	bodyPayload := map[string]interface{}{
		"keyword":     keyword,
		"currentPage": 1,
		"pageSize":    pageSize,
		"timestamp":   time.Now().UnixMilli(),
		"scene":       7,
		"reqScene":    7,
	}
	body, err := s.postForm("mmfinderassistant-bin/contact/search", bodyPayload, "https://channels.weixin.qq.com/platform/post/list")
	if err != nil {
		return nil, err
	}
	if deprecated, msg := wechatBodyDeprecated(body); deprecated {
		return nil, fmt.Errorf("%s", msg)
	}
	list := parseWechatContactSearch(body)
	if len(list) > 0 {
		return list, nil
	}
	return nil, nil
}

func (s *wechatSession) fetchContactFeeds(finderUsername string, maxPages int) ([]map[string]interface{}, error) {
	if s.wxChannelEnabled() {
		return s.wxChannelFetchFeeds(finderUsername, maxPages)
	}
	paths := []string{
		"mmfinderassistant-bin/contact/feed_list",
		"mmfinderassistant-bin/feed/feed_list",
		"mmfinderassistant-bin/mmfinder/feed/feed_list",
	}
	lastBuffer := ""
	var all []map[string]interface{}
	for page := 0; page < maxPages; page++ {
		var got []map[string]interface{}
		var err error
		for _, p := range paths {
			payload := map[string]interface{}{
				"finderUsername": finderUsername,
				"username":       finderUsername,
				"pageSize":       20,
				"timestamp":      time.Now().UnixMilli(),
				"lastBuffer":     lastBuffer,
			}
			body, e := s.postForm(p, payload, wechatDefaultReferer)
			if e != nil {
				err = e
				continue
			}
			if deprecated, _ := wechatBodyDeprecated(body); deprecated {
				err = fmt.Errorf("助手接口已下线")
				continue
			}
			items, buf, e2 := parseWechatFeedListBody(body)
			if e2 != nil {
				err = e2
				continue
			}
			if len(items) > 0 {
				got = items
				lastBuffer = buf
				err = nil
				break
			}
		}
		if err != nil || len(got) == 0 {
			break
		}
		all = append(all, got...)
		if lastBuffer == "" {
			break
		}
	}
	return all, nil
}

func (s *wechatSession) postForm(apiPath string, payload map[string]interface{}, referer string) ([]byte, error) {
	fullURL := apiPath
	if !strings.HasPrefix(apiPath, "http") {
		fullURL = "https://channels.weixin.qq.com/cgi-bin/" + strings.TrimPrefix(apiPath, "/")
	}
	if referer == "" {
		referer = wechatDefaultReferer
	}
	form := wechatPayloadToForm(payload)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Origin", "https://channels.weixin.qq.com")
	req.Header.Set("Cookie", s.cookie)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	// 视频号助手部分接口返回 201 Created，按 2xx 均视为成功
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var root map[string]interface{}
	if json.Unmarshal(body, &root) == nil {
		if msg, _ := root["errMsg"].(string); strings.Contains(msg, "Cannot POST") {
			return nil, fmt.Errorf("视频号接口已变更(%s)，请更新服务端", apiPath)
		}
		if int64FromAny(root["errCode"]) == 300330 {
			return nil, fmt.Errorf("Cookie 已失效，请在 channels.weixin.qq.com 重新登录后复制 Cookie")
		}
	}
	return body, nil
}

func wechatPayloadToForm(payload map[string]interface{}) url.Values {
	form := url.Values{}
	for k, v := range payload {
		if v == nil {
			form.Set(k, "null")
			continue
		}
		switch x := v.(type) {
		case string:
			form.Set(k, x)
		case int:
			form.Set(k, strconv.Itoa(x))
		case int64:
			form.Set(k, strconv.FormatInt(x, 10))
		case float64:
			form.Set(k, strconv.FormatInt(int64(x), 10))
		case bool:
			form.Set(k, strconv.FormatBool(x))
		default:
			b, _ := json.Marshal(v)
			form.Set(k, string(b))
		}
	}
	return form
}

func parseWechatPostListBody(body []byte) ([]map[string]interface{}, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	if int64FromAny(root["errCode"]) != 0 {
		msg, _ := root["errMsg"].(string)
		return nil, fmt.Errorf("%s", msg)
	}
	data, _ := root["data"].(map[string]interface{})
	list, _ := data["list"].([]interface{})
	var out []map[string]interface{}
	for _, it := range list {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func parseWechatContactSearch(body []byte) []map[string]interface{} {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return nil
	}
	if int64FromAny(root["errCode"]) != 0 {
		return nil
	}
	data, _ := root["data"].(map[string]interface{})
	var out []map[string]interface{}
	for _, key := range []string{"infoList", "list", "contactList"} {
		arr, ok := data[key].([]interface{})
		if !ok {
			continue
		}
		for _, it := range arr {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			if c, ok := m["contact"].(map[string]interface{}); ok {
				out = append(out, c)
			} else {
				out = append(out, m)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

func parseWechatFeedListBody(body []byte) ([]map[string]interface{}, string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, "", err
	}
	if int64FromAny(root["errCode"]) != 0 {
		msg, _ := root["errMsg"].(string)
		return nil, "", fmt.Errorf("%s", msg)
	}
	data, _ := root["data"].(map[string]interface{})
	lastBuffer, _ := data["lastBuffer"].(string)
	var items []interface{}
	for _, key := range []string{"object", "list", "feedList"} {
		if arr, ok := data[key].([]interface{}); ok && len(arr) > 0 {
			items = arr
			break
		}
	}
	var out []map[string]interface{}
	for _, it := range items {
		if m, ok := it.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out, lastBuffer, nil
}

func mapWechatPostItem(raw map[string]interface{}, author string) *WechatCrawlVideo {
	exportID := stringifyID(raw["exportId"])
	if exportID == "" {
		exportID = stringifyID(raw["objectId"])
	}
	if exportID == "" {
		return nil
	}
	title := wechatExtractDesc(raw)
	play, digg := wechatExtractStats(raw)
	cover, videoURL := wechatExtractMedia(raw)
	if videoURL == "" {
		videoURL = "https://channels.weixin.qq.com/platform/post/list"
	}
	return &WechatCrawlVideo{
		ExportID:   exportID,
		Title:      title,
		AuthorName: author,
		CoverURL:   cover,
		VideoURL:   videoURL,
		PlayCount:  play,
		DiggCount:  digg,
	}
}

func mapWechatFeedObject(raw map[string]interface{}, author string) *WechatCrawlVideo {
	exportID := stringifyID(raw["id"])
	if exportID == "" {
		exportID = stringifyID(raw["exportId"])
	}
	if exportID == "" {
		return nil
	}
	desc, _ := raw["objectDesc"].(map[string]interface{})
	if desc == nil {
		return mapWechatPostItem(raw, author)
	}
	play, digg := wechatExtractStats(raw)
	title := ""
	if d, ok := desc["description"].(string); ok {
		title = strings.TrimSpace(d)
	}
	cover, videoURL := wechatExtractMediaFromDesc(desc)
	if videoURL == "" {
		nonce := stringifyID(raw["objectNonceId"])
		if nonce != "" {
			videoURL = fmt.Sprintf("https://channels.weixin.qq.com/web/pages/feed?oid=%s&nid=%s", exportID, nonce)
		}
	}
	return &WechatCrawlVideo{
		ExportID:   exportID,
		Title:      title,
		AuthorName: author,
		CoverURL:   cover,
		VideoURL:   videoURL,
		PlayCount:  play,
		DiggCount:  digg,
	}
}

func wechatExtractDesc(raw map[string]interface{}) string {
	desc, _ := raw["desc"].(map[string]interface{})
	if desc == nil {
		return ""
	}
	if d, ok := desc["description"].(string); ok {
		return trimWechatTitle(d)
	}
	return ""
}

func wechatExtractMedia(raw map[string]interface{}) (cover, videoURL string) {
	desc, _ := raw["desc"].(map[string]interface{})
	if desc != nil {
		return wechatExtractMediaFromDesc(desc)
	}
	return "", ""
}

func wechatExtractMediaFromDesc(desc map[string]interface{}) (cover, videoURL string) {
	media, _ := desc["media"].([]interface{})
	if len(media) == 0 {
		return "", ""
	}
	m0, _ := media[0].(map[string]interface{})
	if m0 == nil {
		return "", ""
	}
	if u, ok := m0["thumbUrl"].(string); ok {
		cover = u
	}
	if u, ok := m0["fullUrl"].(string); ok {
		videoURL = u
	}
	if cover == "" {
		if u, ok := m0["coverUrl"].(string); ok {
			cover = u
		}
	}
	return cover, videoURL
}

func trimWechatTitle(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > 120 {
		return string([]rune(s)[:120]) + "…"
	}
	return s
}

func matchWechatCategory(cat WechatOfficialCategory, title string) bool {
	title = strings.ToLower(title)
	for _, m := range cat.Matchers {
		if strings.Contains(title, strings.ToLower(m)) {
			return true
		}
	}
	return strings.Contains(title, strings.ToLower(cat.Keyword))
}

func enrichWechatVideos(rawPosts []map[string]interface{}, metric, source string) []WechatCrawlVideo {
	var out []WechatCrawlVideo
	for _, raw := range rawPosts {
		if v := mapWechatPostItem(raw, ""); v != nil {
			v.Source = source
			v.EffectiveStat, v.StatSource = pickWechatMetric(v.PlayCount, v.DiggCount, metric)
			out = append(out, *v)
		}
	}
	return out
}

func pickWechatMetric(play, digg int64, metric string) (int64, string) {
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

func wechatBodyDeprecated(body []byte) (bool, string) {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return false, ""
	}
	msg, _ := root["errMsg"].(string)
	if strings.Contains(msg, "Cannot POST") {
		return true, "微信已关闭视频号助手后台「搜索账号」接口，请配置 wx_channel 本地代理抓取他人视频"
	}
	return false, ""
}

func wechatExtractStats(raw map[string]interface{}) (play, digg int64) {
	play = int64FromAny(raw["readCount"])
	if play == 0 {
		play = int64FromAny(raw["viewCount"])
	}
	if play == 0 {
		play = int64FromAny(raw["playCount"])
	}
	digg = int64FromAny(raw["likeCount"])
	if digg == 0 {
		digg = int64FromAny(raw["diggCount"])
	}
	if digg == 0 {
		digg = int64FromAny(raw["favCount"])
	}
	if ii, ok := raw["interactionInfo"].(map[string]interface{}); ok {
		if digg == 0 {
			digg = int64FromAny(ii["likeCount"])
		}
		if play == 0 {
			play = int64FromAny(ii["readCount"])
		}
	}
	if st, ok := raw["stat"].(map[string]interface{}); ok {
		if play == 0 {
			play = int64FromAny(st["readCount"])
		}
		if digg == 0 {
			digg = int64FromAny(st["likeCount"])
		}
	}
	return play, digg
}

func (s *wechatSession) enrichFeedStatsWxChannel(v *WechatCrawlVideo, raw map[string]interface{}) {
	if v == nil || !s.wxChannelEnabled() {
		return
	}
	if v.PlayCount > 0 && v.DiggCount > 0 {
		return
	}
	nonce := stringifyID(raw["objectNonceId"])
	if nonce == "" {
		return
	}
	q := url.Values{}
	q.Set("objectId", v.ExportID)
	q.Set("nonceId", nonce)
	body, err := s.wxChannelGET("/api/channels/feed/profile", q)
	if err != nil {
		return
	}
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return
	}
	data, _ := root["data"].(map[string]interface{})
	obj, _ := data["object"].(map[string]interface{})
	if obj == nil {
		return
	}
	play, digg := wechatExtractStats(obj)
	if play > 0 {
		v.PlayCount = play
	}
	if digg > 0 {
		v.DiggCount = digg
	}
}

func formatWxChannelConnErr(base string, err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "actively refused") ||
		strings.Contains(msg, "connectex") {
		return fmt.Sprintf(
			"无法连接 wx_channel（%s）：本机未启动 wx_channel 或端口不对。请先下载运行 wx_channel.exe（默认 API 为 http://127.0.0.1:2026，代理端口 2025）；若用 -p 改端口则 API=代理端口+1。注意：须在与「本后台服务」同一台电脑上运行；后台若在远程服务器，127.0.0.1 指向服务器而非你电脑。",
			base,
		)
	}
	return "无法连接 wx_channel（" + base + "）：" + msg
}

func (s *wechatSession) checkWxChannelReady() (bool, string) {
	body, err := s.wxChannelGET("/api/channels/status", nil)
	if err != nil {
		return false, formatWxChannelConnErr(s.wxChannelBase, err)
	}
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return false, "wx_channel 状态解析失败"
	}
	data, _ := root["data"].(map[string]interface{})
	if data == nil {
		data = root
	}
	searchReady := int64FromAny(data["search_ready_clients"])
	feedReady := int64FromAny(data["feed_ready_clients"])
	clients := int64FromAny(data["clients"])
	if clients == 0 {
		if n, ok := data["client_list"].([]interface{}); ok {
			clients = int64(len(n))
		}
	}
	ready := searchReady > 0 || feedReady > 0
	if !ready {
		if list, ok := data["client_list"].([]interface{}); ok {
			for _, it := range list {
				m, ok := it.(map[string]interface{})
				if !ok {
					continue
				}
				if m["supports_search"] == true || m["supports_feed"] == true {
					ready = true
					break
				}
			}
		}
	}
	if ready {
		return true, fmt.Sprintf("wx_channel 已就绪（%d 个页面已连接），可搜索他人视频", clients)
	}
	return false, "wx_channel 服务已启动（端口 2026），但尚无浏览器页面连接。请用 Chrome/Edge 打开 https://channels.weixin.qq.com 的视频号页并保持不关；若日志提示「注入失败」，请右键 wx_channel.exe → 以管理员身份运行后重开浏览器"
}

func (s *wechatSession) wxChannelGET(path string, query url.Values) ([]byte, error) {
	u := s.wxChannelBase + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return nil, err
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("wx_channel HTTP %d", resp.StatusCode)
	}
	return body, nil
}

func (s *wechatSession) wxChannelPOSTJSON(path string, payload interface{}) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, s.wxChannelBase+path, strings.NewReader(string(raw)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("wx_channel HTTP %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (s *wechatSession) wxChannelSearchContacts(keyword string, searchType, pageSize int) ([]map[string]interface{}, error) {
	q := url.Values{}
	q.Set("keyword", keyword)
	if searchType > 0 {
		q.Set("type", strconv.Itoa(searchType))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	body, err := s.wxChannelGET("/api/channels/contact/search", q)
	if err != nil {
		body, err = s.wxChannelPOSTJSON("/api/channels/contact/search", map[string]interface{}{
			"keyword": keyword, "type": searchType, "page_size": pageSize,
		})
	}
	if err != nil {
		return nil, err
	}
	return parseWxChannelSearchPayload(body), nil
}

func parseWxChannelSearchPayload(body []byte) []map[string]interface{} {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return nil
	}
	data := root
	if d, ok := root["data"].(map[string]interface{}); ok {
		data = d
	}
	var out []map[string]interface{}
	for _, key := range []string{"infoList", "list", "contactList"} {
		arr, ok := data[key].([]interface{})
		if !ok {
			continue
		}
		for _, it := range arr {
			m, ok := it.(map[string]interface{})
			if !ok {
				continue
			}
			if c, ok := m["contact"].(map[string]interface{}); ok {
				out = append(out, c)
			} else if m["username"] != nil || m["nickname"] != nil {
				out = append(out, m)
			}
		}
	}
	return out
}

func (s *wechatSession) searchOthersVideosWxChannel(keyword string, maxPages int, metric string) ([]WechatCrawlVideo, error) {
	var out []WechatCrawlVideo
	marker := ""
	for page := 0; page < maxPages; page++ {
		q := url.Values{}
		q.Set("keyword", keyword)
		q.Set("type", "3")
		if marker != "" {
			q.Set("next_marker", marker)
		}
		body, err := s.wxChannelPOSTJSON("/api/channels/contact/search", map[string]interface{}{
			"keyword": keyword, "type": 3, "next_marker": marker,
		})
		if err != nil {
			body, err = s.wxChannelGET("/api/channels/contact/search", q)
		}
		if err != nil {
			return out, err
		}
		videos, next := parseWxChannelVideoSearchPayload(body, metric)
		out = append(out, videos...)
		marker = next
		if marker == "" || len(videos) == 0 {
			break
		}
	}
	return out, nil
}

func parseWxChannelVideoSearchPayload(body []byte, metric string) ([]WechatCrawlVideo, string) {
	var root map[string]interface{}
	if json.Unmarshal(body, &root) != nil {
		return nil, ""
	}
	data := root
	if d, ok := root["data"].(map[string]interface{}); ok {
		data = d
	}
	next, _ := data["lastBuffer"].(string)
	if next == "" {
		next, _ = data["last_buffer"].(string)
	}
	var items []interface{}
	for _, key := range []string{"object", "feedList", "list", "objects"} {
		if arr, ok := data[key].([]interface{}); ok && len(arr) > 0 {
			items = arr
			break
		}
	}
	var out []WechatCrawlVideo
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		author, _ := m["nickname"].(string)
		if author == "" {
			if c, ok := m["contact"].(map[string]interface{}); ok {
				author, _ = c["nickname"].(string)
			}
		}
		v := mapWechatFeedObject(m, author)
		if v == nil {
			continue
		}
		v.Source = "video_search"
		v.EffectiveStat, v.StatSource = pickWechatMetric(v.PlayCount, v.DiggCount, metric)
		if v.VideoURL == "" {
			nonce := stringifyID(m["objectNonceId"])
			if nonce != "" {
				v.VideoURL = fmt.Sprintf("https://channels.weixin.qq.com/web/pages/feed?oid=%s&nid=%s", v.ExportID, nonce)
			}
		}
		out = append(out, *v)
	}
	return out, next
}

func (s *wechatSession) wxChannelFetchFeeds(username string, maxPages int) ([]map[string]interface{}, error) {
	var all []map[string]interface{}
	marker := ""
	for page := 0; page < maxPages; page++ {
		q := url.Values{}
		q.Set("username", username)
		if marker != "" {
			q.Set("next_marker", marker)
		}
		body, err := s.wxChannelGET("/api/channels/contact/feed/list", q)
		if err != nil {
			return all, err
		}
		items, next, err := parseWechatFeedListBody(body)
		if err != nil {
			return all, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		marker = next
		if marker == "" {
			break
		}
	}
	return all, nil
}

func trimWechatTop(list []WechatCrawlVideo, limit int) []WechatCrawlVideo {
	sort.Slice(list, func(i, j int) bool {
		return list[i].EffectiveStat > list[j].EffectiveStat
	})
	if len(list) > limit {
		list = list[:limit]
	}
	return list
}

