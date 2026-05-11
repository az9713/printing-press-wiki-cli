# Generating the Wikipedia CLI on Windows with Git Bash

A step-by-step walkthrough of how this CLI was built using [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press) on Windows 11 with Git Bash.

---

## What is CLI Printing Press?

[CLI Printing Press](https://github.com/mvanhorn/cli-printing-press) is an open-source generator that takes an API spec (OpenAPI, HAR capture, or docs URL) and produces a fully functional Go CLI with:

- Structured `--json` output and `--select field1,field2` filtering on every command
- A local SQLite cache (`sync` / `sql`) for offline use
- An MCP server so AI agents can call the CLI as a tool
- `--agent` flag that activates JSON + compact + non-interactive mode in one flag
- Typed exit codes (0 = success, 2 = not found, 3 = disambiguation, 5 = API error)
- `doctor`, `profile`, `workflow`, and other production-ready scaffolding

The generator is invoked through a Claude Code skill: `/printing-press <API name>`.

---

## Prerequisites

| Tool | Version used | Install |
|------|-------------|---------|
| Windows 11 | 10.0.26200 | — |
| Git Bash | bundled with Git for Windows | [git-scm.com](https://git-scm.com/downloads/win) |
| Go | 1.26.3 | [go.dev/dl](https://go.dev/dl/) |
| Claude Code | latest | [claude.ai/code](https://claude.ai/code) |
| CLI Printing Press binary | 4.2.2 | `go install github.com/mvanhorn/cli-printing-press/v4/cmd/printing-press@latest` |
| gh CLI | any recent | [cli.github.com](https://cli.github.com/) |

### Install the Printing Press binary

Open Git Bash and run:

```bash
go install github.com/mvanhorn/cli-printing-press/v4/cmd/printing-press@latest
printing-press --version
# printing-press 4.2.2
```

### Install the Claude Code plugin

In Claude Code, run:

```
/install-plugin mvanhorn/cli-printing-press
```

This adds the `/printing-press` skill to your Claude Code session.

---

## Step 1 — Invoke the skill

Open Claude Code from the `cli-printing-press` repo directory (or any directory), then run:

```
/printing-press Wikipedia
```

The skill runs a preflight check, finds the binary, confirms Go is available, and asks for any upfront context.

### Context provided at briefing

For this run, the following context was shared:

```
Target API: Wikipedia / MediaWiki public APIs
No authentication — English Wikipedia first — polite User-Agent

Required commands:
  search QUERY --limit N
  summary TITLE
  article TITLE --format text|html|json
  random
  on-this-day --month M --day D

Agent-native: --json, --compact, --agent, --select, typed exit codes
Optional SQLite cache: sync, sql
No write/edit commands
```

---

## Step 2 — Research phase

The skill searched for:

- Official Wikipedia/Wikimedia API documentation and specs
- Competing CLI tools (`wikit`, `wiki-cli`)
- npm and PyPI SDKs (`wikipedia`, `wtf_wikipedia`, `wikipedia-api`)
- Community MCP servers for Wikipedia
- Rate limits and User-Agent requirements

**Key finding:** Wikipedia's legacy REST API publishes an OpenAPI 3.0.1 spec at:

```
https://en.wikipedia.org/api/rest_v1/?spec
```

This spec covers `page/summary`, `page/html`, `page/random/summary`, `feed/onthisday`, and `feed/featured` — but requires a proper `User-Agent` header. The generator's HTTP client gets a 403 fetching it directly, so the spec was downloaded locally with `curl`:

```bash
curl -s -A "wikipedia-pp-cli/1.0 (contact)" \
  "https://en.wikipedia.org/api/rest_v1/?spec" \
  -o wikipedia-openapi.json
```

The official spec was missing two endpoints (`page/random/summary` and `feed/onthisday`). A focused spec was hand-authored as `wikipedia-focused.yaml` covering exactly the five required commands plus `feed/featured`.

---

## Step 3 — Absorb gate (feature planning)

Before generating any code, the skill catalogued every feature from every competing tool and defined novel features that no existing tool has. The approved manifest:

**Absorbed (match + beat):** 10 features from 7 tools — summary, search, random, full article (text/html/json), language flag, thumbnail access, content URLs, coordinates.

**Novel (only possible with this CLI):**

| Feature | Command | Why unique |
|---------|---------|-----------|
| On-this-day feed | `on-this-day` | Wikipedia's `feed/onthisday` endpoint; no other CLI exposes it |
| Agent-native flag | `--agent` | Combines `--json --compact --no-input` in one flag |
| Disambiguation exit | `summary "Mercury"` → exit 3 | Distinguishes "not found" (exit 2) from "disambiguation" (exit 3) |
| Field selection | `--select title,description` | Works on every command; reduces agent context cost |
| SQLite cache | `sync` + `sql` | Offline re-queries; no competitor has local storage |
| Featured article | `feed featured` | Wikipedia's `feed/featured` endpoint; no other CLI exposes it |

---

## Step 4 — Generation

```bash
printing-press generate \
  --spec wikipedia-focused.yaml \
  --output ./wikipedia-pp-cli \
  --research-dir ./runs/20260510-212353 \
  --force --lenient --validate
```

The generator:

1. Parses the OpenAPI spec
2. Emits the full Go module: `cmd/`, `internal/cli/`, `internal/client/`, `internal/store/`, `internal/mcp/`, `internal/cliutil/`
3. Runs quality gates: `go mod tidy`, `govulncheck`, `go vet`, `go build`
4. Writes `README.md`, `SKILL.md`, `AGENTS.md` from `research.json` narrative

On Windows the `--help` quality gate fails because the generator looks for `wikipedia-pp-cli-validation` without the `.exe` extension. The code itself builds cleanly; the gate failure is a known Windows issue in the generator.

```bash
# Build manually after generation
go build -o ./wikipedia-pp-cli.exe ./cmd/wikipedia-pp-cli/
./wikipedia-pp-cli.exe --help
```

---

## Step 5 — Hand-authored commands (Phase 3)

The generated code scaffolds the spec-derived commands under `page` and `feed` subcommand groups. The user-required top-level interface (`summary TITLE`, `search QUERY`, `random`, `on-this-day`, `article TITLE --format`) was hand-authored as novel commands in `internal/cli/`:

| File | Command | API used |
|------|---------|---------|
| `summary_novel.go` | `summary TITLE` | `rest_v1/page/summary/{title}` — adds disambiguation exit 3 |
| `search_novel.go` | `search QUERY --limit N` | `api.wikimedia.org/core/v1/wikipedia/en/search/page` |
| `random_novel.go` | `random` | `rest_v1/page/random/summary` |
| `onthisday_novel.go` | `on-this-day --month M --day D --type T` | `rest_v1/feed/onthisday/{type}/{mm}/{dd}` |
| `article_novel.go` | `article TITLE --format text\|html\|json` | text: Action API `explaintext=true`; html: rest_v1/page/html; json: rest_v1/page/summary |
| `sql_novel.go` | `sql "SELECT ..."` | Local SQLite via `internal/store` |

All six files were added and wired into `root.go`:

```go
// root.go — novel top-level commands (hand-authored Phase 3)
rootCmd.AddCommand(newSearchCmd(flags))
rootCmd.AddCommand(newSummaryCmd(flags))
rootCmd.AddCommand(newRandomCmd(flags))
rootCmd.AddCommand(newOnThisDayCmd(flags))
rootCmd.AddCommand(newArticleCmd(flags))
rootCmd.AddCommand(newSQLCmd(flags))
```

---

## Step 6 — Shipcheck

```bash
printing-press shipcheck \
  --dir ./wikipedia-pp-cli \
  --spec wikipedia-focused.yaml \
  --research-dir ./runs/20260510-212353
```

Results after one fix loop:

| Leg | Result |
|-----|--------|
| dogfood | PASS |
| verify | PASS — 16/16 commands |
| workflow-verify | PASS |
| verify-skill | PASS (requires `PYTHONIOENCODING=utf-8` on Windows — see below) |
| validate-narrative | PASS — 10/10 examples resolved |
| scorecard | 83/100 — Grade B |

### Windows-specific: verify-skill encoding

On Windows, the `verify-skill` Python subprocess defaults to `cp1252` encoding, which cannot print the `✓` character the tool uses. The direct tool passes when the encoding is set:

```bash
PYTHONIOENCODING=utf-8 printing-press verify-skill --dir ./wikipedia-pp-cli
# === wikipedia-pp-cli ===
#   ✓ All checks passed
```

This is a known bug in the Printing Press generator (the subprocess wrapper should set UTF-8 encoding on Windows).

---

## Step 7 — Live smoke tests

Wikipedia's API requires no authentication, so all five commands were tested live:

```bash
# 1. Summary with agent output
./wikipedia-pp-cli.exe summary "Alan Turing" --agent --select title,description
# {"description":"English computer scientist (1912–1954)","title":"Alan Turing"}

# 2. Disambiguation — exit 3
./wikipedia-pp-cli.exe summary "Mercury" --quiet; echo "exit: $?"
# exit: 3

# 3. Search
./wikipedia-pp-cli.exe search "quantum computing" --limit 3 --json --select pages.title,pages.description

# 4. On this day — moon landing
./wikipedia-pp-cli.exe on-this-day --month 7 --day 20 --type selected --json --select selected.text,selected.year

# 5. Article text
./wikipedia-pp-cli.exe article "Photon" --format text --limit 500

# 6. Doctor
./wikipedia-pp-cli.exe doctor --json
# {"api":"reachable","auth":"not required",...}
```

---

## Step 8 — Push to GitHub

```bash
cd ~/printing-press/library/wikipedia

# Create .gitignore (exclude compiled binaries)
echo "*.exe" >> .gitignore
echo "wikipedia-pp-cli" >> .gitignore

git init
git checkout -b main
git add -A
git commit -m "feat: initial Wikipedia CLI"

# Create repo and push using gh CLI
gh repo create az9713/printing-press-wiki-cli \
  --public \
  --description "Wikipedia CLI — search, summary, random, on-this-day, article" \
  --source . \
  --remote origin \
  --push
```

---

## Repo structure

```
wikipedia-pp-cli/
├── cmd/
│   ├── wikipedia-pp-cli/main.go     ← CLI binary entry point
│   └── wikipedia-pp-mcp/main.go     ← MCP server entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                  ← Cobra tree, rootFlags, --agent/--json/--select
│   │   ├── page.go                  ← Generated: page subcommand group
│   │   ├── page_summary.go          ← Generated: page summary <title>
│   │   ├── page_html.go             ← Generated: page html <title>
│   │   ├── page_random.go           ← Generated: page random
│   │   ├── feed.go                  ← Generated: feed subcommand group
│   │   ├── feed_onthisday.go        ← Generated: feed onthisday <mm> <dd>
│   │   ├── feed_featured.go         ← Generated: feed featured <yyyy> <mm> <dd>
│   │   ├── search_novel.go          ← Hand-authored: search QUERY
│   │   ├── summary_novel.go         ← Hand-authored: summary TITLE + exit 3
│   │   ├── random_novel.go          ← Hand-authored: random
│   │   ├── onthisday_novel.go       ← Hand-authored: on-this-day --month --day
│   │   ├── article_novel.go         ← Hand-authored: article TITLE --format
│   │   ├── sql_novel.go             ← Hand-authored: sql "SELECT ..."
│   │   ├── helpers.go               ← Shared: filterFields, printOutput, exit codes
│   │   ├── doctor.go                ← Health check command
│   │   └── sync.go                  ← Sync to SQLite
│   ├── client/client.go             ← HTTP client: retry, cache, rate limit, User-Agent
│   ├── config/config.go             ← Config loader (~/.config/wikipedia-pp-cli/)
│   ├── store/store.go               ← SQLite store: migrations, DB(), OpenReadOnly()
│   ├── cliutil/                     ← Generator-owned utilities (do not edit)
│   └── mcp/                         ← MCP server: typed tools + Cobra-tree walker
├── go.mod                           ← module wikipedia-pp-cli
├── README.md
├── SKILL.md                         ← Agent skill for Claude Code
└── AGENTS.md                        ← Notes for AI agents working on this codebase
```

---

## Key design decisions

**Why a focused spec instead of the official one?**
The official `rest_v1/?spec` omits `page/random/summary` and `feed/onthisday` — both real endpoints that are well-documented but not in the published spec. A hand-authored focused spec lets the generator scaffold exactly the commands we want without generating 40+ unused endpoints.

**Why hand-author `summary` as a top-level command?**
The generator produces `page summary <title>` (nested under the `page` group). The user requirement was `summary TITLE` at the root level. Rather than restructure the generated code, the novel command wraps the same API call and adds the disambiguation exit code logic (checking `response.type == "disambiguation"` and returning exit 3).

**Why does `search` use `api.wikimedia.org` instead of `rest_v1`?**
The Wikipedia legacy REST API (`en.wikipedia.org/api/rest_v1`) does not include a search endpoint. Search lives at the newer Core REST API (`api.wikimedia.org/core/v1/wikipedia/{lang}/search/page`). Since this is a different base URL from the generated client's `BaseURL`, the `search` command makes a direct `net/http` call rather than routing through the generated client.

**Why exit code 3 for disambiguation?**
The Wikipedia API returns `{"type": "disambiguation"}` in the summary response for pages like "Mercury" that match multiple articles. Most CLI tools ignore this and return partial content. Exit code 3 lets an agent detect disambiguation and ask the user to clarify the title (e.g., "Mercury (planet)") rather than silently returning the wrong article.

---

## Resources

- [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press) — the generator used to build this CLI
- [Wikipedia REST API docs](https://www.mediawiki.org/wiki/Wikimedia_REST_API)
- [MediaWiki Action API](https://www.mediawiki.org/wiki/API:Main_page)
- [Wikimedia Core REST API](https://api.wikimedia.org/wiki/Main_Page)
- [This repository](https://github.com/az9713/printing-press-wiki-cli)
