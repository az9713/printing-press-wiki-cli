# Wikipedia CLI

**Every Wikipedia reading pattern as a composable command — with structured output, offline cache, and an agent-native surface no other Wikipedia tool has.**

wikipedia-pp-cli wraps the Wikipedia REST API in five focused commands: search, summary, page, random, and on-this-day. Every command speaks --json with --select field filtering so AI agents never have to parse paragraphs. A local SQLite cache makes repeated lookups instant.

Learn more at [Wikipedia](https://www.mediawiki.org/wiki/Wikimedia_REST_API).

> **Built with [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)** — an open-source generator that turns an OpenAPI spec into a production-quality Go CLI with offline search, SQLite caching, structured JSON output, and an MCP server surface. See [`generate-wikipedia-cli-git-bash-windows.md`](generate-wikipedia-cli-git-bash-windows.md) for a step-by-step walkthrough of how this CLI was generated on Windows.

## Install

The recommended path installs both the `wikipedia-pp-cli` binary and the `pp-wikipedia` agent skill in one shot:

```bash
npx -y @mvanhorn/printing-press install wikipedia
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press install wikipedia --cli-only
```


### Without Node

The generated install path is category-agnostic until this CLI is published. If `npx` is not available before publish, install Node or use the category-specific Go fallback from the public-library entry after publish.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/wikipedia-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-wikipedia --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-wikipedia --force
```

## Install for OpenClaw

Tell your OpenClaw agent (copy this):

```
Install the pp-wikipedia skill from https://github.com/mvanhorn/printing-press-library/tree/main/cli-skills/pp-wikipedia. The skill defines how its required CLI can be installed.
```

## Quick Start

```bash
# Structured summary for agents — exits 0 on success, 2 on not-found, 3 on disambiguation
wikipedia-pp-cli summary "Alan Turing" --agent


# Search with structured results — pipe to jq or --select title,description
wikipedia-pp-cli search "quantum computing" --limit 5 --json


# Moon landing day — curated historical events as structured JSON
wikipedia-pp-cli on-this-day --month 7 --day 20 --type selected --agent


# Random article title + description — no API key required
wikipedia-pp-cli random --compact


# Full article as plain text — first 2000 chars
wikipedia-pp-cli article "Python (programming language)" --format text --limit 2000

```

## Unique Features

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

## Usage

Run `wikipedia-pp-cli --help` for the full command reference and flag list.

## Commands

### feed

Manage feed

- **`wikipedia-pp-cli feed featured`** - Returns Wikipedia's featured article, image, and news for a given date.
- **`wikipedia-pp-cli feed onthisday`** - Returns historical events, births, deaths, and holidays for a given calendar date.

### page

Manage page

- **`wikipedia-pp-cli page html`** - Returns the full Parsoid HTML of the given article.
- **`wikipedia-pp-cli page random`** - Returns the summary of a randomly selected Wikipedia article.
- **`wikipedia-pp-cli page summary`** - Returns a summary of the given article including title, description, extract, and thumbnail. Returns exit code 3 for disambiguation pages.


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
wikipedia-pp-cli feed featured 2026 05 10

# JSON for scripting and agents
wikipedia-pp-cli feed featured 2026 05 10 --json

# Filter to specific fields
wikipedia-pp-cli feed featured 2026 05 10 --json --select id,name,status

# Dry run — show the request without sending
wikipedia-pp-cli feed featured 2026 05 10 --dry-run

# Agent mode — JSON + compact + no prompts in one flag
wikipedia-pp-cli feed featured 2026 05 10 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `5` API error, `7` rate limited, `10` config error.

## Use with Claude Code

Install the focused skill — it auto-installs the CLI on first invocation:

```bash
npx skills add mvanhorn/printing-press-library/cli-skills/pp-wikipedia -g
```

Then invoke `/pp-wikipedia <query>` in Claude Code. The skill is the most efficient path — Claude Code drives the CLI directly without an MCP server in the middle.

<details>
<summary>Use as an MCP server in Claude Code (advanced)</summary>

If you'd rather register this CLI as an MCP server in Claude Code, install the MCP binary first:


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Then register it:

```bash
claude mcp add wikipedia wikipedia-pp-mcp
```

</details>

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/wikipedia-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "wikipedia": {
      "command": "wikipedia-pp-mcp"
    }
  }
}
```

</details>

## Health Check

```bash
wikipedia-pp-cli doctor
```

Verifies configuration and connectivity to the API.

## Configuration

Config file: `~/.config/wikipedia-pp-cli/config.toml`

Static request headers can be configured under `headers`; per-command header overrides take precedence.

## Troubleshooting
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific

- **exit 3 on summary** — disambiguation page — add qualifier: summary "Mercury (planet)"
- **exit 2 on summary** — article not found — check spelling, try search first
- **slow first request** — Wikipedia CDN cache miss — second request is instant
- **HTML in text output** — use --format text not --format html for page command

---

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**wtf_wikipedia**](https://github.com/spencermountain/wtf_wikipedia) — JavaScript (2800 stars)
- [**wikipedia (PyPI)**](https://github.com/goldsmith/Wikipedia) — Python (2700 stars)
- [**wikit**](https://github.com/KorySchneider/wikit) — JavaScript (1200 stars)
- [**wikipedia (npm)**](https://github.com/dopecodez/Wikipedia) — JavaScript (890 stars)
- [**wikipedia-api (PyPI)**](https://github.com/martin-majlis/Wikipedia-API) — Python (500 stars)

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
