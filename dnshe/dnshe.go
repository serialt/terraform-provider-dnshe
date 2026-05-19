package dnshe

import (
	"fmt"
	"strconv"
	"time"

	"github.com/krabt/req"
)

const DefaultBaseURL = "https://api005.dnshe.com/index.php?m=domain_hub"

// APIError 统一错误结构体
type APIError struct {
	Success   bool                   `json:"success"`
	ErrorCode string                 `json:"error_code"`
	Message   string                 `json:"message"`
	ErrorStr  string                 `json:"error"`
	Details   map[string]interface{} `json:"details"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[API Error] Code: %s, Message: %s", e.ErrorCode, e.Message)
}

// Client DNSHE 客户端
type Client struct {
	client *req.Client
}

// NewClient 初始化客户端
func NewClient(url, apiKey, apiSecret string) *Client {
	if url == "" {
		url = DefaultBaseURL
	}
	c := req.NewClient().
		SetBaseURL(url).
		SetTimeout(30*time.Second).
		// 统一注入安全认证 Header
		SetCommonHeader("X-API-Key", apiKey).
		SetCommonHeader("X-API-Secret", apiSecret).
		// 当 HTTP 状态码 >= 400 时，自动将 Body 解析进 APIError，并绑定到结果的 Error 中
		SetCommonErrorResult(&APIError{}).
		// 开启自动包装，如果返回错误，会自动转换成我们定义的 APIError
		SetCommonContentType("application/json")

	// 可选：如果需要调试，可以打开此开关输出极度详细的 HTTP 请求抓包日志
	// c.DevMode()

	return &Client{client: c}
}

// doRequest 封装 req/v3 的请求管道
func (c *Client) doRequest(method, endpoint, action string, queryParams map[string]string, body interface{}, result interface{}) error {
	reqObj := c.client.R().
		SetQueryParam("endpoint", endpoint)

	if action != "" {
		reqObj.SetQueryParam("action", action)
	}
	if queryParams != nil {
		reqObj.SetQueryParams(queryParams)
	}
	if body != nil {
		reqObj.SetBody(body)
	}
	if result != nil {
		reqObj.SetResult(result)
	}

	// 修正这里：req/v3 使用 Send(method, url) 执行通用请求
	resp, err := reqObj.Send(method, "")
	if err != nil {
		return err
	}

	// 如果触发了错误状态码（通过 SetCommonErrorResult 绑定）
	if resp.IsError() {
		return resp.ErrorResult().(*APIError)
	}

	return nil
}

// ==========================================
// 1. 子域名管理模型与方法
// ==========================================

type ListSubdomainsParams struct {
	Page         int    `url:"page,omitempty"`
	PerPage      int    `url:"per_page,omitempty"`
	IncludeTotal bool   `url:"include_total,omitempty"`
	Search       string `url:"search,omitempty"`
	RootDomain   string `url:"rootdomain,omitempty"`
	Status       string `url:"status,omitempty"`
	CreatedFrom  string `url:"created_from,omitempty"`
	CreatedTo    string `url:"created_to,omitempty"`
	SortBy       string `url:"sort_by,omitempty"`
	SortDir      string `url:"sort_dir,omitempty"`
	Fields       string `url:"fields,omitempty"`
}

type Subdomain struct {
	ID                int    `json:"id"`
	Subdomain         string `json:"subdomain"`
	RootDomain        string `json:"rootdomain"`
	FullDomain        string `json:"full_domain"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	ExpiresAt         string `json:"expires_at,omitempty"`
	NeverExpires      int    `json:"never_expires,omitempty"`
	CloudflareZoneID  string `json:"cloudflare_zone_id,omitempty"`
	ProviderAccountID int    `json:"provider_account_id,omitempty"`
}

type Pagination struct {
	Page     int  `json:"page"`
	PerPage  int  `json:"per_page"`
	HasMore  bool `json:"has_more"`
	NextPage int  `json:"next_page"`
	PrevPage int  `json:"prev_page"`
	Total    int  `json:"total"`
}

type ListSubdomainsResponse struct {
	Success    bool        `json:"success"`
	Count      int         `json:"count"`
	Subdomains []Subdomain `json:"subdomains"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func (c *Client) ListSubdomains(params ListSubdomainsParams) (*ListSubdomainsResponse, error) {
	q := make(map[string]string)
	if params.Page > 0 {
		q["page"] = strconv.Itoa(params.Page)
	}
	if params.PerPage > 0 {
		q["per_page"] = strconv.Itoa(params.PerPage)
	}
	if params.IncludeTotal {
		q["include_total"] = "1"
	}
	if params.Search != "" {
		q["search"] = params.Search
	}
	if params.RootDomain != "" {
		q["rootdomain"] = params.RootDomain
	}
	if params.Status != "" {
		q["status"] = params.Status
	}
	if params.CreatedFrom != "" {
		q["created_from"] = params.CreatedFrom
	}
	if params.CreatedTo != "" {
		q["created_to"] = params.CreatedTo
	}
	if params.SortBy != "" {
		q["sort_by"] = params.SortBy
	}
	if params.SortDir != "" {
		q["sort_dir"] = params.SortDir
	}
	if params.Fields != "" {
		q["fields"] = params.Fields
	}

	var resp ListSubdomainsResponse
	err := c.doRequest("GET", "subdomains", "list", q, nil, &resp)
	return &resp, err
}

type RegisterSubdomainRequest struct {
	Subdomain  string `json:"subdomain"`
	RootDomain string `json:"rootdomain"`
}

type RegisterSubdomainResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	SubdomainID int    `json:"subdomain_id"`
	FullDomain  string `json:"full_domain"`
}

func (c *Client) RegisterSubdomain(subdomain, rootdomain string) (*RegisterSubdomainResponse, error) {
	reqData := RegisterSubdomainRequest{Subdomain: subdomain, RootDomain: rootdomain}
	var resp RegisterSubdomainResponse
	err := c.doRequest("POST", "subdomains", "register", nil, reqData, &resp)
	return &resp, err
}

type GetSubdomainResponse struct {
	Success    bool        `json:"success"`
	Subdomain  Subdomain   `json:"subdomain"`
	DNSRecords []DNSRecord `json:"dns_records"`
	DNSCount   int         `json:"dns_count"`
}

func (c *Client) GetSubdomain(subdomainID int) (*GetSubdomainResponse, error) {
	q := map[string]string{"subdomain_id": strconv.Itoa(subdomainID)}
	var resp GetSubdomainResponse
	err := c.doRequest("GET", "subdomains", "get", q, nil, &resp)
	return &resp, err
}

type DeleteSubdomainRequest struct {
	SubdomainID int `json:"subdomain_id"`
}

type DeleteSubdomainResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	SubdomainID       int    `json:"subdomain_id"`
	FullDomain        string `json:"full_domain"`
	DNSRecordsDeleted int    `json:"dns_records_deleted"`
}

func (c *Client) DeleteSubdomain(subdomainID int) (*DeleteSubdomainResponse, error) {
	reqData := DeleteSubdomainRequest{SubdomainID: subdomainID}
	var resp DeleteSubdomainResponse
	err := c.doRequest("POST", "subdomains", "delete", nil, reqData, &resp)
	return &resp, err
}

type RenewSubdomainRequest struct {
	SubdomainID int `json:"subdomain_id"`
}

type RenewSubdomainResponse struct {
	Success           bool    `json:"success"`
	Message           string  `json:"message"`
	SubdomainID       int     `json:"subdomain_id"`
	Subdomain         string  `json:"subdomain"`
	PreviousExpiresAt string  `json:"previous_expires_at"`
	NewExpiresAt      string  `json:"new_expires_at"`
	RenewedAt         string  `json:"renewed_at"`
	NeverExpires      int     `json:"never_expires"`
	Status            string  `json:"status"`
	RemainingDays     int     `json:"remaining_days"`
	ChargedAmount     float64 `json:"charged_amount"`
}

func (c *Client) RenewSubdomain(subdomainID int) (*RenewSubdomainResponse, error) {
	reqData := RenewSubdomainRequest{SubdomainID: subdomainID}
	var resp RenewSubdomainResponse
	err := c.doRequest("POST", "subdomains", "renew", nil, reqData, &resp)
	return &resp, err
}

// ==========================================
// 2. DNS记录管理模型与方法
// ==========================================

type DNSRecord struct {
	ID        int    `json:"id,omitempty"`
	RecordID  string `json:"record_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Content   string `json:"content,omitempty"`
	TTL       int    `json:"ttl,omitempty"`
	Priority  *int   `json:"priority,omitempty"`
	Line      string `json:"line,omitempty"`
	Proxied   bool   `json:"proxied,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type ListDNSRecordsResponse struct {
	Success bool        `json:"success"`
	Count   int         `json:"count"`
	Records []DNSRecord `json:"records"`
}

func (c *Client) ListDNSRecords(subdomainID int) (*ListDNSRecordsResponse, error) {
	q := map[string]string{"subdomain_id": strconv.Itoa(subdomainID)}
	var resp ListDNSRecordsResponse
	err := c.doRequest("GET", "dns_records", "list", q, nil, &resp)
	return &resp, err
}

type CreateDNSRecordRequest struct {
	SubdomainID  int    `json:"subdomain_id"`
	Type         string `json:"type"`
	Name         string `json:"name,omitempty"`
	Content      string `json:"content,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
	Priority     *int   `json:"priority,omitempty"`
	Line         string `json:"line,omitempty"`
	RecordWeight *int   `json:"record_weight,omitempty"`
	Weight       *int   `json:"weight,omitempty"`
	RecordPort   *int   `json:"record_port,omitempty"`
	Port         *int   `json:"port,omitempty"`
	RecordTarget string `json:"record_target,omitempty"`
	Target       string `json:"target,omitempty"`
	CAAFlag      *int   `json:"caa_flag,omitempty"`
	CAATag       string `json:"caa_tag,omitempty"`
	CAAValue     string `json:"caa_value,omitempty"`
}

type CreateDNSRecordResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ID       int    `json:"id"`
	RecordID string `json:"record_id"`
}

func (c *Client) CreateDNSRecord(reqData CreateDNSRecordRequest) (*CreateDNSRecordResponse, error) {
	var resp CreateDNSRecordResponse
	err := c.doRequest("POST", "dns_records", "create", nil, reqData, &resp)
	return &resp, err
}

type UpdateDNSRecordRequest struct {
	ID           int    `json:"id,omitempty"`
	RecordID     string `json:"record_id,omitempty"`
	Type         string `json:"type,omitempty"`
	Name         string `json:"name,omitempty"`
	Content      string `json:"content,omitempty"`
	TTL          int    `json:"ttl,omitempty"`
	Priority     *int   `json:"priority,omitempty"`
	Line         string `json:"line,omitempty"`
	RecordWeight *int   `json:"record_weight,omitempty"`
	Weight       *int   `json:"weight,omitempty"`
	RecordPort   *int   `json:"record_port,omitempty"`
	Port         *int   `json:"port,omitempty"`
	RecordTarget string `json:"record_target,omitempty"`
	Target       string `json:"target,omitempty"`
	CAAFlag      *int   `json:"caa_flag,omitempty"`
	CAATag       string `json:"caa_tag,omitempty"`
	CAAValue     string `json:"caa_value,omitempty"`
}

type UpdateDNSRecordResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ID       int    `json:"id"`
	RecordID string `json:"record_id"`
}

func (c *Client) UpdateDNSRecord(reqData UpdateDNSRecordRequest) (*UpdateDNSRecordResponse, error) {
	var resp UpdateDNSRecordResponse
	err := c.doRequest("POST", "dns_records", "update", nil, reqData, &resp)
	return &resp, err
}

type DeleteDNSRecordRequest struct {
	ID       int    `json:"id,omitempty"`
	RecordID string `json:"record_id,omitempty"`
}

type DeleteDNSRecordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *Client) DeleteDNSRecord(id int, recordID string) (*DeleteDNSRecordResponse, error) {
	reqData := DeleteDNSRecordRequest{}
	if id > 0 {
		reqData.ID = id
	}
	if recordID != "" {
		reqData.RecordID = recordID
	}
	var resp DeleteDNSRecordResponse
	err := c.doRequest("POST", "dns_records", "delete", nil, reqData, &resp)
	return &resp, err
}

// ==========================================
// 3. API 密钥管理模型与方法
// ==========================================

type APIKeyInfo struct {
	ID           int    `json:"id"`
	KeyName      string `json:"key_name"`
	APIKey       string `json:"api_key"`
	Status       string `json:"status"`
	RequestCount int    `json:"request_count"`
	LastUsedAt   string `json:"last_used_at"`
	CreatedAt    string `json:"created_at"`
}

type ListAPIKeysResponse struct {
	Success bool         `json:"success"`
	Count   int          `json:"count"`
	Keys    []APIKeyInfo `json:"keys"`
}

func (c *Client) ListAPIKeys() (*ListAPIKeysResponse, error) {
	var resp ListAPIKeysResponse
	err := c.doRequest("GET", "keys", "list", nil, nil, &resp)
	return &resp, err
}

type CreateAPIKeyRequest struct {
	KeyName     string `json:"key_name"`
	IPWhitelist string `json:"ip_whitelist,omitempty"`
}

type CreateAPIKeyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Warning   string `json:"warning"`
}

func (c *Client) CreateAPIKey(keyName, ipWhitelist string) (*CreateAPIKeyResponse, error) {
	reqData := CreateAPIKeyRequest{KeyName: keyName, IPWhitelist: ipWhitelist}
	var resp CreateAPIKeyResponse
	err := c.doRequest("POST", "keys", "create", nil, reqData, &resp)
	return &resp, err
}

type DeleteAPIKeyRequest struct {
	KeyID int `json:"key_id"`
}

type DeleteAPIKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (c *Client) DeleteAPIKey(keyID int) (*DeleteAPIKeyResponse, error) {
	reqData := DeleteAPIKeyRequest{KeyID: keyID}
	var resp DeleteAPIKeyResponse
	err := c.doRequest("POST", "keys", "delete", nil, reqData, &resp)
	return &resp, err
}

type RegenerateAPIKeyRequest struct {
	KeyID int `json:"key_id"`
}

type RegenerateAPIKeyResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
	Warning   string `json:"warning"`
}

func (c *Client) RegenerateAPIKey(keyID int) (*RegenerateAPIKeyResponse, error) {
	reqData := RegenerateAPIKeyRequest{KeyID: keyID}
	var resp RegenerateAPIKeyResponse
	err := c.doRequest("POST", "keys", "regenerate", nil, reqData, &resp)
	return &resp, err
}

// ==========================================
// 4. 配额管理模型与方法
// ==========================================

type Quota struct {
	Used        int `json:"used"`
	Base        int `json:"base"`
	InviteBonus int `json:"invite_bonus"`
	Total       int `json:"total"`
	Available   int `json:"available"`
}

type GetQuotaResponse struct {
	Success bool  `json:"success"`
	Quota   Quota `json:"quota"`
}

func (c *Client) GetQuota() (*GetQuotaResponse, error) {
	var resp GetQuotaResponse
	err := c.doRequest("GET", "quota", "", nil, nil, &resp)
	return &resp, err
}
