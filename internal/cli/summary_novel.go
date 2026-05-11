// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: top-level summary command with disambiguation exit code.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSummaryCmd(flags *rootFlags) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "summary <title>",
		Short: "Get a Wikipedia article summary",
		Long: `Fetches the summary of a Wikipedia article: title, description, plain-text extract, and thumbnail.

Exit codes:
  0  success
  2  article not found
  3  disambiguation page — refine the title (e.g. "Mercury (planet)")
  5  API error`,
		Example: strings.Trim(`
  wikipedia-pp-cli summary "Alan Turing"
  wikipedia-pp-cli summary "Python (programming language)" --json
  wikipedia-pp-cli summary "Mount Everest" --agent --select title,description,coordinates
  wikipedia-pp-cli summary "Mercury"   # exits 3 — disambiguation
  wikipedia-pp-cli summary "Paris" --lang fr`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only":       "true",
			"pp:typed-exit-codes": "0,2,3,5",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				return nil
			}

			title := strings.Join(args, " ")
			c, err := flags.newClient()
			if err != nil {
				return err
			}

			// Use language-prefixed base URL when non-English
			path := "/page/summary/" + encodePathSegment(title)
			if lang != "en" {
				// Swap base URL to the correct language wiki
				c.BaseURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", lang)
			}

			data, err := c.Get(path, nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}

			// Detect disambiguation pages: type == "disambiguation"
			var article struct {
				Type        string `json:"type"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Extract     string `json:"extract"`
			}
			if jsonErr := json.Unmarshal(data, &article); jsonErr == nil {
				if article.Type == "disambiguation" {
					if !flags.quiet {
						fmt.Fprintf(cmd.ErrOrStderr(), "Disambiguation: %q matches multiple articles. Try a more specific title.\n", title)
						fmt.Fprintf(cmd.ErrOrStderr(), "  Example: %s summary %q\n", cmd.Root().Name(), article.Title+" (topic)")
					}
					if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
						_ = printOutputWithFlags(cmd.OutOrStdout(), data, flags)
					}
					return &cliError{code: 3, err: fmt.Errorf("disambiguation page: %s", title)}
				}
				if article.Type == "no-extract" || article.Extract == "" {
					if !flags.quiet {
						fmt.Fprintf(cmd.ErrOrStderr(), "No extract available for %q.\n", title)
					}
				}
			}

			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				filtered := data
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				} else if flags.compact {
					type compactSummary struct {
						Title       string `json:"title"`
						Description string `json:"description"`
						Extract     string `json:"extract"`
					}
					cs := compactSummary{
						Title:       article.Title,
						Description: article.Description,
						Extract:     article.Extract,
					}
					if len(cs.Extract) > 500 {
						cs.Extract = cs.Extract[:497] + "..."
					}
					d, _ := json.Marshal(cs)
					filtered = d
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(json.RawMessage(filtered))
			}

			// Human output
			fmt.Fprintf(cmd.OutOrStdout(), "== %s ==\n", article.Title)
			if article.Description != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n", article.Description)
			}
			if article.Extract != "" {
				extract := article.Extract
				if len(extract) > 2000 {
					extract = extract[:1997] + "..."
				}
				fmt.Fprintln(cmd.OutOrStdout(), extract)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "en", "Wikipedia language code (e.g. fr, de, es)")
	return cmd
}

// encodePathSegment encodes a path segment, replacing spaces with underscores per Wikipedia convention.
func encodePathSegment(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}
