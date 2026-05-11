// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: on-this-day top-level command with fixed flag interface.

package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newOnThisDayCmd(flags *rootFlags) *cobra.Command {
	var month, day int
	var eventType string
	var lang string

	now := time.Now()

	cmd := &cobra.Command{
		Use:   "on-this-day",
		Short: "Get historical events for a calendar date",
		Long: `Returns curated historical events, notable births, deaths, and holidays for a given calendar date.
Uses Wikipedia's On This Day feed (rest_v1/feed/onthisday).

Types:
  selected  Curated significant events (default)
  events    All historical events
  births    Notable births
  deaths    Notable deaths
  holidays  Holidays and observances
  all       All categories combined`,
		Example: strings.Trim(`
  wikipedia-pp-cli on-this-day
  wikipedia-pp-cli on-this-day --month 7 --day 20 --type selected
  wikipedia-pp-cli on-this-day --month 11 --day 9 --type events --json
  wikipedia-pp-cli on-this-day --agent --select selected.0.text,selected.0.year
  wikipedia-pp-cli on-this-day --type births --month 3 --day 14`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return nil
			}

			// Validate type
			allowed := []string{"all", "selected", "births", "deaths", "events", "holidays"}
			validType := false
			for _, v := range allowed {
				if eventType == v {
					validType = true
					break
				}
			}
			if !validType {
				return usageErr(fmt.Errorf("--type must be one of: %s", strings.Join(allowed, ", ")))
			}

			if month < 1 || month > 12 {
				return usageErr(fmt.Errorf("--month must be between 1 and 12"))
			}
			if day < 1 || day > 31 {
				return usageErr(fmt.Errorf("--day must be between 1 and 31"))
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if lang != "en" {
				c.BaseURL = fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1", lang)
			}

			path := fmt.Sprintf("/feed/onthisday/%s/%02d/%02d", eventType, month, day)
			data, err := c.Get(path, nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}

			// Strip HTML from displaytitle fields in the pages array of each event.
			// The Wikipedia on-this-day API embeds HTML spans in displaytitle
			// (e.g., <span lang="en">…</span>); clean them so all output modes
			// return plain text.
			data = stripOnThisDayDisplayTitles(data)

			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				filtered := data
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				} else if flags.compact {
					// compact: return only the requested type's events with text+year
					type event struct {
						Year int    `json:"year"`
						Text string `json:"text"`
					}
					var raw map[string]json.RawMessage
					if jsonErr := json.Unmarshal(data, &raw); jsonErr == nil {
						if eventType == "all" {
							// return all as-is but compact per field
							filtered = data
						} else if section, ok := raw[eventType]; ok {
							var events []event
							_ = json.Unmarshal(section, &events)
							d, _ := json.Marshal(events)
							filtered = d
						}
					}
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(json.RawMessage(filtered))
			}

			// Human output
			fmt.Fprintf(cmd.OutOrStdout(), "== On This Day: %s %d ==\n\n", monthName(month), day)
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return err
			}

			type event struct {
				Year  int    `json:"year"`
				Text  string `json:"text"`
				Pages []struct {
					Title string `json:"title"`
				} `json:"pages"`
			}

			sections := []string{eventType}
			if eventType == "all" {
				sections = []string{"selected", "events", "births", "deaths", "holidays"}
			}

			for _, section := range sections {
				raw2, ok := raw[section]
				if !ok {
					continue
				}
				var events []event
				if err := json.Unmarshal(raw2, &events); err != nil || len(events) == 0 {
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "-- %s --\n", strings.ToUpper(section))
				for _, e := range events {
					year := ""
					if e.Year != 0 {
						year = fmt.Sprintf("%d: ", e.Year)
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %s%s\n", year, e.Text)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&month, "month", int(now.Month()), "Month (1-12, defaults to current month)")
	cmd.Flags().IntVar(&day, "day", now.Day(), "Day (1-31, defaults to today)")
	cmd.Flags().StringVar(&eventType, "type", "selected", "Event type: selected, events, births, deaths, holidays, all")
	cmd.Flags().StringVar(&lang, "lang", "en", "Wikipedia language code")

	return cmd
}

// stripOnThisDayDisplayTitles walks the on-this-day JSON structure and strips
// HTML from every displaytitle field inside pages arrays. The API consistently
// wraps display titles in <span> tags that are meaningful only in a browser.
func stripOnThisDayDisplayTitles(data json.RawMessage) json.RawMessage {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return data
	}
	type page struct {
		DisplayTitle string          `json:"displaytitle"`
		Rest         json.RawMessage `json:"-"`
	}
	type event struct {
		Pages []json.RawMessage `json:"pages"`
		Rest  json.RawMessage   `json:"-"`
	}
	changed := false
	for section, sectionData := range raw {
		var events []map[string]json.RawMessage
		if err := json.Unmarshal(sectionData, &events); err != nil {
			continue
		}
		for i, ev := range events {
			pagesRaw, ok := ev["pages"]
			if !ok {
				continue
			}
			var pages []map[string]json.RawMessage
			if err := json.Unmarshal(pagesRaw, &pages); err != nil {
				continue
			}
			for j, pg := range pages {
				dtRaw, hasDT := pg["displaytitle"]
				if !hasDT {
					continue
				}
				var dt string
				if err := json.Unmarshal(dtRaw, &dt); err != nil {
					continue
				}
				clean := stripHTMLTags(dt)
				if clean == dt {
					continue
				}
				cleaned, err := json.Marshal(clean)
				if err != nil {
					continue
				}
				pages[j]["displaytitle"] = cleaned
				changed = true
			}
			if changed {
				pagesBytes, err := json.Marshal(pages)
				if err == nil {
					ev["pages"] = pagesBytes
					events[i] = ev
				}
			}
		}
		if changed {
			sectionBytes, err := json.Marshal(events)
			if err == nil {
				raw[section] = sectionBytes
			}
		}
	}
	if !changed {
		return data
	}
	result, err := json.Marshal(raw)
	if err != nil {
		return data
	}
	return result
}

func monthName(m int) string {
	months := [...]string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	if m < 1 || m > 12 {
		return "Unknown"
	}
	return months[m]
}
