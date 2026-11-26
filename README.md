# saascity-go

[![Go Reference](https://pkg.go.dev/badge/github.com/best-directories/saascity-go.svg)](https://pkg.go.dev/github.com/best-directories/saascity-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

Zero-dependency Go SDK and CLI for **[SaaSCity](https://saascity.io)** — the curated SaaS & AI launch directory on a live isometric city map.

Calls SaaSCity's keyless Streamable HTTP **Model Context Protocol (MCP)** endpoint (`https://saascity.io/api/mcp`) and **NLWeb** discovery endpoint (`https://saascity.io/ask`) using only the Go standard library (`net/http`, `encoding/json`).

- **Website:** [https://saascity.io](https://saascity.io)
- **MCP Server Docs:** [https://saascity.io/mcp](https://saascity.io/mcp)
- **CLI & Integrations:** [https://saascity.io/integrations/cli](https://saascity.io/integrations/cli)
- **Free Launch & SEO Tools:** [https://saascity.io/free-tools](https://saascity.io/free-tools)
- **Best SaaS Directories Dataset:** [https://saascity.io/best-saas-directories](https://saascity.io/best-saas-directories)

## Install

```bash
go get github.com/best-directories/saascity-go@v1.0.0
```

Or install the standalone Go CLI:

```bash
go install github.com/best-directories/saascity-go/cmd/saascity@v1.0.0
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	saascity "github.com/best-directories/saascity-go"
)

func main() {
	ctx := context.Background()
	client := saascity.NewClient()

	// 1. Run a 16-point homepage launch readiness audit (no API key required)
	report, err := client.CheckLaunchReadiness(ctx, "https://acme.dev")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Readiness: %d/100 (%s)\n", report.Score, report.Verdict)

	// 2. Filter 1,600+ startup, SaaS & AI directories by DR, dofollow & pricing
	dirs, err := client.FindLaunchDirectories(ctx, saascity.DirectoryFilter{
		Category:           "SaaS Tools",
		DofollowOnly:       true,
		FreeOrFreemiumOnly: true,
		MinDR:              40,
		Limit:              10,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range dirs.Directories {
		fmt.Printf("- %s (%s) DR %d [%s]\n", d.Name, d.Domain, d.DR, d.SubmitURL)
	}
}
```

## License

MIT — see [LICENSE](LICENSE).
