package content

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type OfficeTempEmailService struct{}

// TempMailboxResult 创建临时邮箱结果
type TempMailboxResult struct {
	Mailbox  string `json:"mailbox"`
	Login    string `json:"login"`
	Domain   string `json:"domain"`
	Token    string `json:"token,omitempty"` // mail.tm 鉴权
	Provider string `json:"provider"`
}

func (s *OfficeTempEmailService) provider() string {
	p := strings.ToLower(strings.TrimSpace(global.GVA_CONFIG.OfficeTools.TempEmailProvider))
	if p == "" {
		return "mailtm"
	}
	return p
}

func (s *OfficeTempEmailService) mailtmBase() string {
	base := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.MailTmAPI)
	if base == "" {
		base = "https://api.mail.tm"
	}
	return strings.TrimRight(base, "/")
}

func (s *OfficeTempEmailService) secmailBase() string {
	base := strings.TrimSpace(global.GVA_CONFIG.OfficeTools.TempEmailAPI)
	if base == "" {
		base = "https://www.1secmail.com/api/v1/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base
}

func randomHex(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func (s *OfficeTempEmailService) httpClient() *http.Client {
	return &http.Client{Timeout: 25 * time.Second}
}

func (s *OfficeTempEmailService) httpJSON(method, rawURL string, body interface{}, bearer string, dest interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GVA-OfficeTools/1.0")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		n := len(data)
		if n > 300 {
			n = 300
		}
		return fmt.Errorf("临时邮箱上游 HTTP %d: %s", resp.StatusCode, string(data[:n]))
	}
	if dest == nil {
		return nil
	}
	return json.Unmarshal(data, dest)
}

// CreateMailbox 申请临时邮箱
func (s *OfficeTempEmailService) CreateMailbox() (*TempMailboxResult, error) {
	switch s.provider() {
	case "1secmail", "secmail":
		return s.createSecMail()
	default:
		return s.createMailTm()
	}
}

func (s *OfficeTempEmailService) mailtmBases() []string {
	primary := s.mailtmBase()
	bases := []string{primary}
	if primary != "https://api.mail.gw" {
		bases = append(bases, "https://api.mail.gw")
	}
	return bases
}

func (s *OfficeTempEmailService) createMailTm() (*TempMailboxResult, error) {
	var lastErr error
	for _, base := range s.mailtmBases() {
		r, err := s.createMailTmAt(base)
		if err == nil {
			return r, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("无法连接临时邮箱服务")
}

type mailTmDomainItem struct {
	Domain   string `json:"domain"`
	IsActive bool   `json:"isActive"`
}

func parseMailTmDomainsJSON(data []byte) ([]mailTmDomainItem, error) {
	var list []mailTmDomainItem
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}
	var wrapped struct {
		Members []mailTmDomainItem `json:"hydra:member"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("解析域名列表失败: %w", err)
	}
	return wrapped.Members, nil
}

func (s *OfficeTempEmailService) fetchMailTmDomains(base string) ([]mailTmDomainItem, error) {
	req, err := http.NewRequest(http.MethodGet, base+"/domains", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GVA-OfficeTools/1.0")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		n := len(data)
		if n > 300 {
			n = 300
		}
		return nil, fmt.Errorf("临时邮箱上游 HTTP %d: %s", resp.StatusCode, string(data[:n]))
	}
	return parseMailTmDomainsJSON(data)
}

func (s *OfficeTempEmailService) createMailTmAt(base string) (*TempMailboxResult, error) {
	domains, err := s.fetchMailTmDomains(base)
	if err != nil {
		return nil, err
	}
	var domain string
	for _, d := range domains {
		if d.IsActive && d.Domain != "" {
			domain = d.Domain
			break
		}
	}
	if domain == "" {
		return nil, fmt.Errorf("未获取到可用邮箱域名")
	}
	login := randomHex(10)
	password := randomHex(16)
	address := login + "@" + domain
	if err := s.httpJSON(http.MethodPost, base+"/accounts", map[string]string{
		"address":  address,
		"password": password,
	}, "", nil); err != nil {
		return nil, fmt.Errorf("创建邮箱失败: %w", err)
	}
	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := s.httpJSON(http.MethodPost, base+"/token", map[string]string{
		"address":  address,
		"password": password,
	}, "", &tokenResp); err != nil {
		return nil, fmt.Errorf("获取令牌失败: %w", err)
	}
	if tokenResp.Token == "" {
		return nil, fmt.Errorf("上游未返回 token")
	}
	return &TempMailboxResult{
		Mailbox:  address,
		Login:    login,
		Domain:   domain,
		Token:    tokenResp.Token,
		Provider: "mailtm",
	}, nil
}

func (s *OfficeTempEmailService) createSecMail() (*TempMailboxResult, error) {
	u := s.secmailBase() + "?action=genRandomMailbox&count=1"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GVA-OfficeTools/1.0)")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		n := len(body)
		if n > 300 {
			n = 300
		}
		return nil, fmt.Errorf("临时邮箱上游 HTTP %d: %s", resp.StatusCode, string(body[:n]))
	}
	var list []string
	if err := json.Unmarshal(body, &list); err != nil || len(list) == 0 {
		return nil, fmt.Errorf("上游未返回邮箱地址")
	}
	addr := list[0]
	parts := strings.SplitN(addr, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("邮箱地址格式异常")
	}
	return &TempMailboxResult{
		Mailbox:  addr,
		Login:    parts[0],
		Domain:   parts[1],
		Provider: "1secmail",
	}, nil
}

// ListMessages 收件箱（mail.tm 需 token；1secmail 需 login+domain）
func (s *OfficeTempEmailService) ListMessages(token, login, domain string) ([]map[string]interface{}, error) {
	if strings.TrimSpace(token) != "" {
		return s.listMailTmMessages(token)
	}
	return s.listSecMailMessages(login, domain)
}

func (s *OfficeTempEmailService) listMailTmMessages(token string) ([]map[string]interface{}, error) {
	base := s.mailtmBase()
	return s.listMailTmMessagesAt(base, token)
}

type mailTmMessageItem struct {
	ID      string `json:"id"`
	From    struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	} `json:"from"`
	Subject   string `json:"subject"`
	Intro     string `json:"intro"`
	CreatedAt string `json:"createdAt"`
}

func parseMailTmMessagesJSON(data []byte) ([]mailTmMessageItem, error) {
	var list []mailTmMessageItem
	if err := json.Unmarshal(data, &list); err == nil {
		return list, nil
	}
	var wrapped struct {
		Members []mailTmMessageItem `json:"hydra:member"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("解析邮件列表失败: %w", err)
	}
	return wrapped.Members, nil
}

func (s *OfficeTempEmailService) listMailTmMessagesAt(base, token string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, base+"/messages", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GVA-OfficeTools/1.0")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("临时邮箱上游 HTTP %d", resp.StatusCode)
	}
	members, err := parseMailTmMessagesJSON(data)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(members))
	for _, m := range members {
		from := m.From.Address
		if m.From.Name != "" {
			from = fmt.Sprintf("%s <%s>", m.From.Name, m.From.Address)
		}
		out = append(out, map[string]interface{}{
			"id":      m.ID,
			"from":    from,
			"subject": m.Subject,
			"date":    m.CreatedAt,
			"intro":   m.Intro,
		})
	}
	return out, nil
}

func (s *OfficeTempEmailService) listSecMailMessages(login, domain string) ([]map[string]interface{}, error) {
	login = strings.TrimSpace(login)
	domain = strings.TrimSpace(domain)
	if login == "" || domain == "" {
		return nil, fmt.Errorf("login 与 domain 不能为空")
	}
	q := url.Values{}
	q.Set("action", "getMessages")
	q.Set("login", login)
	q.Set("domain", domain)
	u := s.secmailBase() + "?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GVA-OfficeTools/1.0)")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("临时邮箱上游 HTTP %d", resp.StatusCode)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ReadMessage 读信
func (s *OfficeTempEmailService) ReadMessage(token, login, domain, id string) (map[string]interface{}, error) {
	if strings.TrimSpace(token) != "" {
		return s.readMailTmMessage(token, id)
	}
	return s.readSecMailMessage(login, domain, id)
}

func (s *OfficeTempEmailService) readMailTmMessage(token, id string) (map[string]interface{}, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id 不能为空")
	}
	return s.readMailTmMessageAt(s.mailtmBase(), token, id)
}

func (s *OfficeTempEmailService) readMailTmMessageAt(base, token, id string) (map[string]interface{}, error) {
	var m struct {
		From struct {
			Address string `json:"address"`
			Name    string `json:"name"`
		} `json:"from"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
		HTML    []string `json:"html"`
		Intro   string   `json:"intro"`
	}
	if err := s.httpJSON(http.MethodGet, base+"/messages/"+id, nil, token, &m); err != nil {
		return nil, err
	}
	from := m.From.Address
	if m.From.Name != "" {
		from = fmt.Sprintf("%s <%s>", m.From.Name, m.From.Address)
	}
	htmlBody := strings.Join(m.HTML, "")
	if htmlBody == "" {
		htmlBody = m.Intro
	}
	return map[string]interface{}{
		"from":     from,
		"subject":  m.Subject,
		"textBody": m.Text,
		"htmlBody": htmlBody,
	}, nil
}

func (s *OfficeTempEmailService) readSecMailMessage(login, domain, id string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("action", "readMessage")
	q.Set("login", strings.TrimSpace(login))
	q.Set("domain", strings.TrimSpace(domain))
	q.Set("id", strings.TrimSpace(id))
	u := s.secmailBase() + "?" + q.Encode()
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GVA-OfficeTools/1.0)")
	resp, err := s.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("临时邮箱上游 HTTP %d", resp.StatusCode)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
