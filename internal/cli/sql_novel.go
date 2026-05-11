// Copyright 2026 az9713. Licensed under Apache-2.0. See LICENSE.
// Hand-authored novel feature: sql command for querying the local SQLite cache.

package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"wikipedia-pp-cli/internal/store"
)

func newSQLCmd(flags *rootFlags) *cobra.Command {
	var dbPath string

	cmd := &cobra.Command{
		Use:   "sql <query>",
		Short: "Run a read-only SQL query against the local Wikipedia cache",
		Long: `Execute a SELECT query against the local SQLite database populated by 'wikipedia-pp-cli sync'.

Only SELECT statements are allowed. The database schema mirrors synced Wikipedia article summaries.
Run 'wikipedia-pp-cli sync' first to populate the database.`,
		Example: strings.Trim(`
  wikipedia-pp-cli sql "SELECT title, description FROM articles LIMIT 10"
  wikipedia-pp-cli sql "SELECT title FROM articles WHERE description LIKE '%physicist%'"
  wikipedia-pp-cli sql "SELECT count(*) FROM articles" --json
  wikipedia-pp-cli sql "SELECT title, description FROM articles ORDER BY title LIMIT 20" --json`, "\n"),
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
			trimmed := strings.TrimSpace(strings.ToUpper(query))
			if !strings.HasPrefix(trimmed, "SELECT") {
				return usageErr(fmt.Errorf("only SELECT statements are allowed"))
			}

			if dbPath == "" {
				home, _ := os.UserHomeDir()
				dbPath = filepath.Join(home, ".local", "share", "wikipedia-pp-cli", "data.db")
			}

			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return usageErr(fmt.Errorf("database not found at %s — run 'wikipedia-pp-cli sync' first", dbPath))
			}

			db, err := store.OpenReadOnly(dbPath)
			if err != nil {
				return apiErr(fmt.Errorf("opening database: %w", err))
			}
			defer db.Close()

			rows, err := db.DB().QueryContext(cmd.Context(), query)
			if err != nil {
				return usageErr(fmt.Errorf("query error: %w", err))
			}
			defer rows.Close()

			cols, err := rows.Columns()
			if err != nil {
				return err
			}

			var results []map[string]any
			for rows.Next() {
				vals := make([]any, len(cols))
				valPtrs := make([]any, len(cols))
				for i := range vals {
					valPtrs[i] = &vals[i]
				}
				if err := rows.Scan(valPtrs...); err != nil {
					return err
				}
				row := make(map[string]any, len(cols))
				for i, col := range cols {
					v := vals[i]
					if b, ok := v.([]byte); ok {
						v = string(b)
					}
					row[col] = v
				}
				results = append(results, row)
			}
			if err := rows.Err(); err != nil {
				return err
			}

			if len(results) == 0 {
				if flags.quiet {
					return nil
				}
				fmt.Fprintln(cmd.OutOrStdout(), "No rows returned.")
				return nil
			}

			if flags.asJSON || !isTerminal(cmd.OutOrStdout()) {
				data, _ := json.Marshal(results)
				if flags.selectFields != "" {
					data = filterFields(data, flags.selectFields)
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(json.RawMessage(data))
			}

			// Human table output
			headers := cols
			var tableRows [][]string
			for _, r := range results {
				row := make([]string, len(cols))
				for i, col := range cols {
					v := r[col]
					if v == nil {
						row[i] = ""
					} else {
						s := fmt.Sprintf("%v", v)
						if len(s) > 80 {
							s = s[:77] + "..."
						}
						row[i] = s
					}
				}
				tableRows = append(tableRows, row)
			}
			return flags.printTable(cmd, headers, tableRows)
		},
	}

	// Ensure sql.ErrNoRows is imported (it's used implicitly via rows.Err())
	_ = sql.ErrNoRows

	cmd.Flags().StringVar(&dbPath, "db", "", "Database path (default: ~/.local/share/wikipedia-pp-cli/data.db)")
	return cmd
}
