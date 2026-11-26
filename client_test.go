package saascity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckLaunchReadiness(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mcp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("MCP-Protocol-Version") != MCPProtocolVersion {
			t.Fatalf("missing MCP-Protocol-Version header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "Score: 94/100"}},
				"structuredContent": map[string]any{
					"url":     "https://saascity.io",
					"score":   94,
					"verdict": "ready",
					"checks": []map[string]any{
						{"label": "HTTPS", "status": "pass", "detail": "TLS active"},
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	got, err := c.CheckLaunchReadiness(context.Background(), "https://saascity.io")
	if err != nil {
		t.Fatalf("CheckLaunchReadiness error: %v", err)
	}
	if got.Score != 94 || got.Verdict != "ready" || len(got.Checks) != 1 {
		t.Fatalf("unexpected report: %+v", got)
	}
}

func TestFindLaunchDirectories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result": map[string]any{
				"content": []map[string]any{{"type": "text", "text": "Found 1 directory"}},
				"structuredContent": map[string]any{
					"total":  1,
					"offset": 0,
					"directories": []map[string]any{
						{
							"name":      "SaaSCity",
							"domain":    "saascity.io",
							"dr":        71,
							"linkType":  "dofollow",
							"pricing":   "freemium",
							"category":  "Launchpad",
							"submitUrl": "https://saascity.io/submit",
						},
					},
				},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	res, err := c.FindLaunchDirectories(context.Background(), DirectoryFilter{
		DofollowOnly:       true,
		FreeOrFreemiumOnly: true,
		MinDR:              40,
	})
	if err != nil {
		t.Fatalf("FindLaunchDirectories error: %v", err)
	}
	if res.Total != 1 || res.Directories[0].Domain != "saascity.io" {
		t.Fatalf("unexpected directories: %+v", res)
	}
}
