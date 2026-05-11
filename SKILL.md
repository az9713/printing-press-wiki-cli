---
name: pp-wikipedia
description: "Every Wikipedia reading pattern as a composable command — with structured output, offline cache, and an... Trigger phrases: `look up on Wikipedia`, `what does Wikipedia say about`, `Wikipedia summary of`, `search Wikipedia for`, `what happened on this day`, `get a random Wikipedia article`, `use wikipedia-pp-cli`."
author: "az9713"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - wikipedia-pp-cli
---

# Wikipedia — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `wikipedia-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer:
   ```bash
   npx -y @mvanhorn/printing-press install wikipedia --cli-only
   ```
2. Verify: `wikipedia-pp-cli --version`
3. Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on `$PATH`.

If the `npx` install fails before this CLI has a public-library category, install Node or use the category-specific Go fallback after publish.

If `--version` reports "command not found" after install, the install step did not put the binary on `$PATH`. Do not proceed with skill commands until verification succeeds.

This CLI covers the Wikipedia REST API in five focused commands: search, summary, page, random, and on-this-day. Every command speaks --json with --select field filtering so AI agents never have to parse paragraphs. A local SQLite cache makes repeated lookups instant.

## When to Use This CLI

Use this CLI when an agent needs factual encyclopedia content without a web search. Ideal for: grounding claims against Wikipedia sources, enriching entity data with descriptions and coordinates, generating calendar-aware historical context, or building a local article knowledge base via sync.

## When Not to Use This CLI

Do not activate this CLI for requests that require creating, updating, deleting, publishing, commenting, upvoting, inviting, ordering, sending messages, booking, purchasing, or changing remote state. This printed CLI exposes read-only commands for inspection, export, sync, and analysis.

## Unique Capabilities

These capabilities aren't available in any other tool for this API.

### Unique Wikipedia surfaces
- **`on-this-day`** — Get curated historical events, births, deaths, and holidays for any calendar date.

  _Use this when an agent needs calendar-aware historical context without a web search._

  ```bash
  wikipedia-pp-cli on-this-day --month 7 --day 20 --type selected --agent
  ```
- **`feed featured`** — Get Wikipedia's featured article for any date — the highest-quality content on the site.

  _Use when you need a high-quality, peer-reviewed reference article on any topic for a given date._

  ```bash
  wikipedia-pp-cli feed featured 2026 01 15 --agent
  ```

### Agent-native plumbing
- **`summary --agent`** — Single --agent flag activates JSON output, compact fields, and deterministic exit codes across every command.

  _Use --agent in any pipeline or agent loop that must parse output programmatically._

  ```bash
  wikipedia-pp-cli summary "Alan Turing" --agent --select title,description
  ```
- **`summary`** — Exit code 3 on disambiguation pages so callers know to clarify the query, not just retry.

  _Use exit code 3 to trigger a clarification loop rather than silently returning ambiguous content._

  ```bash
  wikipedia-pp-cli summary Mercury; echo $?  # 3 for disambiguation
  ```
- **`summary --select`** — Narrow JSON output to exactly the fields you need across every command.

  _Use --select to avoid sending tens of KB of Wikipedia HTML through your agent context window._

  ```bash
  wikipedia-pp-cli summary "Python" --json --select title,description,thumbnail.source
  ```

### Local state that compounds
- **`sync`** — Sync article summaries to a local SQLite store for offline re-queries and custom SQL.

  _Use sync when an agent needs to work with a curated set of articles repeatedly without hitting the API._

  ```bash
  wikipedia-pp-cli sync --full
  ```

## Command Reference

**feed** — Manage feed

- `wikipedia-pp-cli feed featured` — Returns Wikipedia's featured article, image, and news for a given date.
- `wikipedia-pp-cli feed onthisday` — Returns historical events, births, deaths, and holidays for a given calendar date.

**page** — Manage page

- `wikipedia-pp-cli page html` — Returns the full Parsoid HTML of the given article.
- `wikipedia-pp-cli page random` — Returns the summary of a randomly selected Wikipedia article.
- `wikipedia-pp-cli page summary` — Returns a summary of the given article including title, description, extract, and thumbnail. Returns exit code 3 for...


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
wikipedia-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

## Recipes


### Agent: look up an entity and extract just the description

```bash
wikipedia-pp-cli summary "Python (programming language)" --agent --select description
```

Returns a single-field JSON object — minimal context consumption

### What happened on this day in history

```bash
wikipedia-pp-cli on-this-day --month 7 --day 20 --type selected --agent --select events.text,events.year
```

Curated events with --select narrowing the nested array fields

### Search then summarize the top result

```bash
wikipedia-pp-cli search "CRISPR gene editing" --limit 1 --json --select pages.0.title
```

Get the top title, then pipe to summary for full context

### Cache a topic set for offline use

```bash
wikipedia-pp-cli sync --full
```

Populates local SQLite; follow with cache search or sql for offline queries

### Query cached articles with SQL

```bash
wikipedia-pp-cli sql "SELECT title, description FROM articles WHERE description LIKE '%physicist%' LIMIT 10"
```

Use when you need filtered article sets without hitting the Wikipedia API

## Auth Setup

No authentication required.

Run `wikipedia-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  wikipedia-pp-cli feed featured 2026 05 10 --agent --select tfa.title,tfa.description
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Read-only** — do not use this CLI for create, update, delete, publish, comment, upvote, invite, order, send, or other mutating requests

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal — piped/agent consumers get pure JSON on stdout.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
wikipedia-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
wikipedia-pp-cli feedback --stdin < notes.txt
wikipedia-pp-cli feedback list --json --limit 10
```

Entries are stored locally at `~/.wikipedia-pp-cli/feedback.jsonl`. They are never POSTed unless `WIKIPEDIA_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `WIKIPEDIA_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
wikipedia-pp-cli profile save briefing --json
wikipedia-pp-cli --profile briefing feed featured 2026 05 10
wikipedia-pp-cli profile list --json
wikipedia-pp-cli profile show briefing
wikipedia-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `wikipedia-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

Install the MCP binary from this CLI's published public-library entry or pre-built release, then register it:

```bash
claude mcp add wikipedia-pp-mcp -- wikipedia-pp-mcp
```

Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which wikipedia-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   wikipedia-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `wikipedia-pp-cli <command> --help`.
