// Package saascity provides a zero-dependency Go client for SaaSCity's keyless
// Model Context Protocol (MCP) server, NLWeb search endpoint, and Maker API.
//
// SaaSCity (https://saascity.io) is the curated launch directory for SaaS and AI
// tools on a live isometric city map, backed by a verified dataset of 1,600+
// startup, SaaS, and AI directories with Ahrefs Domain Rating (DR), dofollow/nofollow
// link verification, and pricing.
//
// # Keyless Endpoints
//
// Read-only launch readiness checks, directory filtering, badge checks, and backlink
// verification call https://saascity.io/api/mcp over JSON-RPC 2.0 (MCP protocol
// version 2025-11-25) with zero API keys or registration required:
//
//   - SaaSCity Homepage: https://saascity.io
//   - MCP Server Documentation: https://saascity.io/mcp
//   - CLI & SDK Integrations: https://saascity.io/integrations/cli
//   - Free SaaS Launch & SEO Tools: https://saascity.io/free-tools
//   - Best SaaS Directories Benchmark: https://saascity.io/best-saas-directories
//
// # Quick Start
//
//	client := saascity.NewClient()
//	audit, err := client.CheckLaunchReadiness(ctx, "https://acme.dev")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Score: %d/100 (%s)\n", audit.Score, audit.Verdict)
package saascity
