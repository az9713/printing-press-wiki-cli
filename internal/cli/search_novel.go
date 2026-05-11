// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: search Wikipedia via Core REST API.

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

const wikiSearchBaseURL = "https://api.wikimedia.org/core/v1/wikipedia"

type searchResult struct {
	Pages []searchPage `json:"pages"`
}

type searchPage struct {
	ID           int64            `json:"id"`
	Key          string           `json:"key"`
	Title        string           `json:"title"`
	Excerpt      string           `json:"excerpt"`
	MatchedTitle *string          `json:"matched_title"`
	Description  string           `json:"description"`
	Thumbnail    *searchThumbnail `json:"thumbnail"`
}

type searchThumbnail struct {
	MimeType string `json:"mimetype"`
	Size     *int64 `json:"size"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration *int64 `json:"duration"`
	URL      string `json:"url"`
}

func newSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var lang string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Wikipedia articles by keyword",
		Long:  "Searches Wikipedia using the Core REST API full-text search. Returns article titles, descriptions, and excerpts.",
		Example: strings.Trim(`
  wikipedia-pp-cli search "quantum computing" --limit 5
  wikipedia-pp-cli search "Alan Turing" --json
  wikipedia-pp-cli search "Python programming" --limit 3 --agent --select pages.0.title,pages.0.description
  wikipedia-pp-cli search "mars" --lang fr`, "\n"),
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

			query := strings.Join(args, " ")
			if query == "" {
				return usageErr(fmt.Errorf("search query is required"))
			}
			if limit < 1 || limit > 100 {
				return usageErr(fmt.Errorf("--limit must be between 1 and 100"))
			}

			apiURL := fmt.Sprintf("%s/%s/search/page", wikiSearchBaseURL, url.PathEscape(lang))
			req, err := http.NewRequestWithContext(cmd.Context(), "GET", apiURL, nil)
			if err != nil {
				return apiErr(fmt.Errorf("building search request: %w", err))
			}
			q := req.URL.Query()
			q.Set("q", query)
			q.Set("limit", fmt.Sprintf("%d", limit))
			req.URL.RawQuery = q.Encode()
			req.Header.Set("User-Agent", "wikipedia-pp-cli/1.0.0 (https://github.com/mvanhorn/cli-printing-press)")

			httpClient := &http.Client{}
			resp, err := httpClient.Do(req)
			if err != nil {
				return apiErr(fmt.Errorf("search request failed: %w", err))
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return apiErr(fmt.Errorf("reading search response: %w", err))
			}

			if resp.StatusCode == 404 {
				return notFoundErr(fmt.Errorf("no results for %q", query))
			}
			if resp.StatusCode >= 400 {
				return apiErr(fmt.Errorf("search returned HTTP %d: %s", resp.StatusCode, string(body)))
			}

			var result searchResult
			if err := json.Unmarshal(body, &result); err != nil {
				return apiErr(fmt.Errorf("parsing search response: %w", err))
			}
			// The Wikipedia search API embeds HTML highlight markers in excerpt fields
			// (e.g., <span class="searchmatch">term</span>). Strip them at parse time so
			// all output modes — JSON, table, plain — return clean text.
			for i := range result.Pages {
				result.Pages[i].Excerpt = stripHTMLTags(result.Pages[i].Excerpt)
				if result.Pages[i].Thumbnail != nil {
					// Normalize protocol-relative URLs to HTTPS for agent-safe output.
					if strings.HasPrefix(result.Pages[i].Thumbnail.URL, "//") {
						result.Pages[i].Thumbnail.URL = "https:" + result.Pages[i].Thumbnail.URL
					}
				}
			}

			if len(result.Pages) == 0 {
				if flags.quiet {
					return nil
				}
				fmt.Fprintln(cmd.ErrOrStderr(), "No results found.")
				return &cliError{code: 2, err: fmt.Errorf("no results for %q", query)}
			}

			data, err := json.Marshal(result)
			if err != nil {
				return err
			}

			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				filtered := json.RawMessage(data)
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				} else if flags.compact {
					// compact: just titles and descriptions
					type compact struct {
						Title       string `json:"title"`
						Description string `json:"description"`
						Key         string `json:"key"`
					}
					var out []compact
					for _, p := range result.Pages {
						out = append(out, compact{Title: p.Title, Description: p.Description, Key: p.Key})
					}
					data2, _ := json.Marshal(out)
					filtered = data2
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(json.RawMessage(filtered))
			}

			// Human output: numbered list
			tw := fmt.Sprintf("Search results for %q (%d found):\n\n", query, len(result.Pages))
			fmt.Fprint(cmd.OutOrStdout(), tw)
			for i, p := range result.Pages {
				fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, p.Title)
				if p.Description != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", p.Description)
				}
				if p.Excerpt != "" {
					excerpt := stripHTMLTags(p.Excerpt)
					if len(excerpt) > 120 {
						excerpt = excerpt[:117] + "..."
					}
					fmt.Fprintf(cmd.OutOrStdout(), "   %s\n", excerpt)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results (1-100)")
	cmd.Flags().StringVar(&lang, "lang", "en", "Wikipedia language code (default: en)")
	return cmd
}

// stripHTMLTags removes basic HTML tags from a string.
func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}
