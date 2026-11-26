// Command saascity runs SaaSCity homepage launch readiness checks and directory queries from Go.
//
// Website: https://saascity.io
// Docs:    https://saascity.io/mcp
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	saascity "github.com/best-directories/saascity-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: saascity <check|badge|ask> <url-or-query>")
		fmt.Fprintln(os.Stderr, "Docs:  https://saascity.io/mcp")
		os.Exit(2)
	}
	client := saascity.NewClient()
	ctx := context.Background()

	switch os.Args[1] {
	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: saascity check <url>")
			os.Exit(2)
		}
		rep, err := client.CheckLaunchReadiness(ctx, os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(rep)
	case "badge":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: saascity badge <url>")
			os.Exit(2)
		}
		res, err := client.CheckBadge(ctx, os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	case "ask":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: saascity ask <query>")
			os.Exit(2)
		}
		items, err := client.AskNLWeb(ctx, os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(items)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand %q (expected check, badge, or ask)\n", os.Args[1])
		os.Exit(2)
	}
}
