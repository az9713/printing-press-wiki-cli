// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: top-level random command.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newRandomCmd(flags *rootFlags) *cobra.Command {
	var lang string

	cmd := &cobra.Command{
		Use:   "random",
		Short: "Get a random Wikipedia article summary",
		Long:  "Returns the summary of a randomly selected Wikipedia article. Each invocation returns a different article.",
		Example: strings.Trim(`
  wikipedia-pp-cli random
  wikipedia-pp-cli random --json
  wikipedia-pp-cli random --agent --select title,description
  wikipedia-pp-cli random --compact`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}

			if lang != "en" {
				c.BaseURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", lang)
			}

			data, err := c.Get("/page/random/summary", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}

			var article struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Extract     string `json:"extract"`
			}
			_ = json.Unmarshal(data, &article)

			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				filtered := data
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				} else if flags.compact {
					type compactRandom struct {
						Title       string `json:"title"`
						Description string `json:"description"`
					}
					d, _ := json.Marshal(compactRandom{Title: article.Title, Description: article.Description})
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
				if len(extract) > 1000 {
					extract = extract[:997] + "..."
				}
				fmt.Fprintln(cmd.OutOrStdout(), extract)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "en", "Wikipedia language code (e.g. fr, de, es)")
	return cmd
}
