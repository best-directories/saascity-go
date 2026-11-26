package saascity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production SaaSCity base URL (https://saascity.io).
	DefaultBaseURL = "https://saascity.io"
	// MCPProtocolVersion is the stateless Streamable HTTP MCP version supported by SaaSCity.
	MCPProtocolVersion = "2025-11-25"
	// Version is the saascity-go SDK version.
	Version = "1.0.0"
)

// Client communicates with SaaSCity's keyless MCP server (https://saascity.io/api/mcp),
// NLWeb search endpoint (https://saascity.io/ask), and authenticated Maker API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the SaaSCity base URL (useful for testing).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	}
}

// WithAPIKey sets a SaaSCity Maker API key (scm_live_...) for Maker API endpoints.
func WithAPIKey(apiKey string) Option {
	return func(c *Client) {
		c.APIKey = strings.TrimSpace(apiKey)
	}
}

// WithHTTPClient provides a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.HTTPClient = hc
		}
	}
}

// NewClient constructs a SaaSCity Client targeting https://saascity.io.
func NewClient(opts ...Option) *Client {
	c := &Client{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "saascity-go/" + Version + " (+https://saascity.io/mcp)",
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	return c
}

type rpcRequest struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Method  string     `json:"method"`
	Params  toolParams `json:"params"`
}

type toolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type rpcResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      int        `json:"id"`
	Result  *mcpResult `json:"result,omitempty"`
	Error   *rpcError  `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpResult struct {
	Content           []mcpContent    `json:"content,omitempty"`
	StructuredContent json.RawMessage `json:"structuredContent,omitempty"`
	IsError           bool            `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResponse holds both the human-readable text output and typed JSON structuredContent
// returned by https://saascity.io/api/mcp.
type ToolResponse struct {
	Text       string          `json:"text"`
	Structured json.RawMessage `json:"structuredContent,omitempty"`
}

// CallTool invokes any keyless SaaSCity MCP tool by name on POST /api/mcp.
// Full tool documentation is available at https://saascity.io/mcp.
func (c *Client) CallTool(ctx context.Context, toolName string, arguments map[string]any) (*ToolResponse, error) {
	if arguments == nil {
		arguments = map[string]any{}
	}
	bodyBytes, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: toolParams{
			Name:      toolName,
			Arguments: arguments,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal mcp request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/mcp", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build mcp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("MCP-Protocol-Version", MCPProtocolVersion)
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post mcp request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read mcp response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mcp HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var parsed rpcResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode mcp response: %w", err)
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("mcp error [%d]: %s", parsed.Error.Code, parsed.Error.Message)
	}
	if parsed.Result == nil {
		return nil, fmt.Errorf("mcp response missing result object")
	}

	var texts []string
	for _, item := range parsed.Result.Content {
		if item.Type == "text" && item.Text != "" {
			texts = append(texts, item.Text)
		}
	}
	joined := strings.Join(texts, "\n")
	if parsed.Result.IsError {
		if joined == "" {
			joined = "mcp tool returned error"
		}
		return nil, fmt.Errorf("%s", joined)
	}

	return &ToolResponse{
		Text:       joined,
		Structured: parsed.Result.StructuredContent,
	}, nil
}

// AuditCheck describes a single check in a 16-point SaaSCity homepage readiness audit.
type AuditCheck struct {
	ID       string `json:"id,omitempty"`
	Label    string `json:"label"`
	Status   string `json:"status"`
	Detail   string `json:"detail"`
	Critical bool   `json:"critical,omitempty"`
}

// LaunchReadinessReport contains the 0-100 score, verdict, and individual SEO/OG/technical checks.
type LaunchReadinessReport struct {
	URL           string       `json:"url"`
	FinalURL      string       `json:"finalUrl,omitempty"`
	Score         int          `json:"score"`
	Verdict       string       `json:"verdict"`
	HTTPStatus    int          `json:"httpStatus,omitempty"`
	ResponseMs    int          `json:"responseMs,omitempty"`
	SaaSCityBadge bool         `json:"saascityBadge,omitempty"`
	Checks        []AuditCheck `json:"checks"`
}

// CheckLaunchReadiness runs SaaSCity's 16-point launch readiness audit against targetURL.
func (c *Client) CheckLaunchReadiness(ctx context.Context, targetURL string) (*LaunchReadinessReport, error) {
	res, err := c.CallTool(ctx, "check_launch_readiness", map[string]any{"url": targetURL})
	if err != nil {
		return nil, err
	}
	var report LaunchReadinessReport
	if len(res.Structured) > 0 {
		if err := json.Unmarshal(res.Structured, &report); err != nil {
			return nil, fmt.Errorf("decode readiness report: %w", err)
		}
	}
	return &report, nil
}

// DirectoryFilter filters SaaSCity's 1,600+ verified launch directories dataset.
type DirectoryFilter struct {
	Query                string
	Category             string
	MinDR                int
	DofollowOnly         bool
	FreeOrFreemiumOnly   bool
	ExcludeBadgeExchange bool
	Limit                int
	Offset               int
}

// DirectoryEntry represents a startup, SaaS, or AI launch directory in SaaSCity's dataset.
type DirectoryEntry struct {
	Name          string `json:"name"`
	Domain        string `json:"domain"`
	DR            int    `json:"dr"`
	LinkType      string `json:"linkType"`
	Pricing       string `json:"pricing"`
	Category      string `json:"category"`
	BadgeExchange bool   `json:"badgeExchange"`
	SubmitURL     string `json:"submitUrl"`
	Note          string `json:"note,omitempty"`
}

// DirectorySearchResult wraps paginated directory matches from find_launch_directories.
type DirectorySearchResult struct {
	Total         int              `json:"total"`
	Offset        int              `json:"offset"`
	Directories   []DirectoryEntry `json:"directories"`
	DataUpdatedAt string           `json:"dataUpdatedAt,omitempty"`
	Attribution   string           `json:"attribution,omitempty"`
}

// FindLaunchDirectories filters 1,600+ launch directories by category, DR, link type, and pricing.
func (c *Client) FindLaunchDirectories(ctx context.Context, filter DirectoryFilter) (*DirectorySearchResult, error) {
	args := map[string]any{}
	if filter.Query != "" {
		args["query"] = filter.Query
	}
	if filter.Category != "" {
		args["category"] = filter.Category
	}
	if filter.MinDR > 0 {
		args["min_dr"] = filter.MinDR
	}
	if filter.DofollowOnly {
		args["link_type"] = "dofollow"
	}
	if filter.FreeOrFreemiumOnly {
		args["pricing"] = "free_or_freemium"
	}
	if filter.ExcludeBadgeExchange {
		args["exclude_badge_exchange"] = true
	}
	if filter.Limit > 0 {
		args["limit"] = filter.Limit
	}
	if filter.Offset > 0 {
		args["offset"] = filter.Offset
	}

	res, err := c.CallTool(ctx, "find_launch_directories", args)
	if err != nil {
		return nil, err
	}
	var out DirectorySearchResult
	if len(res.Structured) > 0 {
		if err := json.Unmarshal(res.Structured, &out); err != nil {
			return nil, fmt.Errorf("decode directory list: %w", err)
		}
	}
	return &out, nil
}

// BadgeCheckResult reports whether the Featured on SaaSCity badge is live on a URL.
type BadgeCheckResult struct {
	Found        bool   `json:"found"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
	BadgeSnippet string `json:"badgeSnippet,omitempty"`
}

// CheckBadge checks whether targetURL displays a verified SaaSCity launch badge.
func (c *Client) CheckBadge(ctx context.Context, targetURL string) (*BadgeCheckResult, error) {
	res, err := c.CallTool(ctx, "check_badge", map[string]any{"url": targetURL})
	if err != nil {
		return nil, err
	}
	var out BadgeCheckResult
	if len(res.Structured) > 0 {
		if err := json.Unmarshal(res.Structured, &out); err != nil {
			return nil, fmt.Errorf("decode badge check: %w", err)
		}
	}
	return &out, nil
}

// NLWebItem represents a natural-language search result from GET https://saascity.io/ask.
type NLWebItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Score       float64 `json:"score,omitempty"`
}

// AskNLWeb queries SaaSCity's NLWeb natural-language discovery endpoint (/ask).
func (c *Client) AskNLWeb(ctx context.Context, query string) ([]NLWebItem, error) {
	u, err := url.Parse(c.BaseURL + "/ask")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("query", query)
	q.Set("streaming", "false")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nlweb HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var wrapper struct {
		Results []NLWebItem `json:"results"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("decode nlweb response: %w", err)
	}
	return wrapper.Results, nil
}
