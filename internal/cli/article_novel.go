// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: top-level "article" command with --format text|html|json.
// Registered as "article" to avoid conflicting with the generated "page" group command.

package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newArticleCmd(flags *rootFlags) *cobra.Command {
	var format string
	var lang string
	var limit int

	cmd := &cobra.Command{
		Use:     "article <title>",
		Short:   "Get a Wikipedia article in text, html, or json format",
		Aliases: []string{"a"},
		Long: `Fetches a full Wikipedia article in multiple formats.

Formats:
  text  Plain text (via Action API, no HTML). Good for reading in a terminal.
  html  Full Parsoid HTML. Pipe to a browser or html renderer.
  json  Structured JSON (summary + extract + links). Good for agents.`,
		Example: strings.Trim(`
  wikipedia-pp-cli article "Python (programming language)" --format text
  wikipedia-pp-cli article "Mount Everest" --format json --agent --select title,description,extract
  wikipedia-pp-cli article "Tokyo" --format html | open -f -a Safari
  wikipedia-pp-cli article "Alan Turing" --format text --limit 3000`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}

			title := strings.Join(args, " ")
			format = strings.ToLower(format)
			allowed := []string{"text", "html", "json"}
			valid := false
			for _, f := range allowed {
				if format == f {
					valid = true
					break
				}
			}
			if !valid {
				return usageErr(fmt.Errorf("--format must be one of: %s", strings.Join(allowed, ", ")))
			}

			switch format {
			case "json":
				return articleJSON(cmd, flags, title, lang)
			case "html":
				return articleHTML(cmd, flags, title, lang)
			case "text":
				return articleText(cmd, flags, title, lang, limit)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Output format: text, html, or json")
	cmd.Flags().StringVar(&lang, "lang", "en", "Wikipedia language code (e.g. fr, de)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Truncate text output to N characters (0 = no limit)")
	return cmd
}

// articleJSON returns the summary JSON (title, description, extract, thumbnail, URLs).
func articleJSON(cmd *cobra.Command, flags *rootFlags, title, lang string) error {
	c, err := flags.newClient()
	if err != nil {
		return err
	}
	if lang != "en" {
		c.BaseURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", lang)
	}

	data, err := c.Get("/page/summary/"+encodePathSegment(title), nil)
	if err != nil {
		return classifyAPIError(err, flags)
	}

	filtered := data
	if flags.selectFields != "" {
		filtered = filterFields(filtered, flags.selectFields)
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(json.RawMessage(filtered))
}

// articleHTML fetches the full Parsoid HTML from rest_v1.
func articleHTML(cmd *cobra.Command, flags *rootFlags, title, lang string) error {
	c, err := flags.newClient()
	if err != nil {
		return err
	}
	if lang != "en" {
		c.BaseURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", lang)
	}

	// rest_v1/page/html returns text/html, not JSON. Use a raw HTTP call.
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/html/%s", lang, url.PathEscape(encodePathSegment(title)))
	req, err := http.NewRequestWithContext(cmd.Context(), "GET", apiURL, nil)
	if err != nil {
		return apiErr(fmt.Errorf("building HTML request: %w", err))
	}
	req.Header.Set("User-Agent", "wikipedia-pp-cli/1.0.0")
	_ = c // silence unused warning

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return apiErr(fmt.Errorf("HTML request failed: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return notFoundErr(fmt.Errorf("article not found: %s", title))
	}
	if resp.StatusCode >= 400 {
		return apiErr(fmt.Errorf("HTML request returned HTTP %d", resp.StatusCode))
	}

	_, err = io.Copy(cmd.OutOrStdout(), resp.Body)
	return err
}

// articleText fetches plain text via the MediaWiki Action API (explaintext=true).
func articleText(cmd *cobra.Command, flags *rootFlags, title, lang string, limitChars int) error {
	apiURL := fmt.Sprintf("https://%s.wikipedia.org/w/api.php", lang)
	req, err := http.NewRequestWithContext(cmd.Context(), "GET", apiURL, nil)
	if err != nil {
		return apiErr(fmt.Errorf("building text request: %w", err))
	}
	q := req.URL.Query()
	q.Set("action", "query")
	q.Set("prop", "extracts")
	q.Set("titles", title)
	q.Set("format", "json")
	q.Set("explaintext", "true")
	q.Set("formatversion", "2")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "wikipedia-pp-cli/1.0.0")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return apiErr(fmt.Errorf("text request failed: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return apiErr(fmt.Errorf("reading text response: %w", err))
	}

	// Parse Action API response: {"query":{"pages":[{"extract":"..."}]}}
	var apiResp struct {
		Query struct {
			Pages []struct {
				Title   string `json:"title"`
				Missing bool   `json:"missing"`
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return apiErr(fmt.Errorf("parsing text response: %w", err))
	}

	if len(apiResp.Query.Pages) == 0 {
		return notFoundErr(fmt.Errorf("article not found: %s", title))
	}

	page := apiResp.Query.Pages[0]
	if page.Missing {
		return notFoundErr(fmt.Errorf("article not found: %s", title))
	}
	if page.Extract == "" {
		return apiErr(fmt.Errorf("no text content available for: %s", title))
	}

	text := page.Extract
	if limitChars > 0 && len(text) > limitChars {
		text = text[:limitChars] + "\n[truncated]"
	}

	fmt.Fprintln(cmd.OutOrStdout(), text)
	return nil
}
