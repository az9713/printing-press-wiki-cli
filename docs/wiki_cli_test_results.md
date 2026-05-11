# Wikipedia CLI — Smoke Test Results

Built with [Printing Press](https://github.com/mvanhorn/cli-printing-press) · version `1.0.0`

---

## Build

```bash
go build -o wikipedia-pp-cli.exe ./cmd/wikipedia-pp-cli/
```

Build succeeded with no errors.

---

## `--version`

```bash
$ ./wikipedia-pp-cli.exe --version
wikipedia-pp-cli 1.0.0
```

---

## `--help`

```bash
$ ./wikipedia-pp-cli.exe --help
Wikipedia CLI — Every Wikipedia reading pattern as a composable command — with structured output, offline cache, and an agent-native su…

Highlights (not in the official API docs):
  • on-this-day   Get curated historical events, births, deaths, and holidays for any calendar date.
  • summary --agent   Single --agent flag activates JSON output, compact fields, and deterministic exit codes across every command.
  • summary   Exit code 3 on disambiguation pages so callers know to clarify the query, not just retry.
  • summary --select   Narrow JSON output to exactly the fields you need across every command.
  • sync   Sync article summaries to a local SQLite store for offline re-queries and custom SQL.
  • feed featured   Get Wikipedia's featured article for any date — the highest-quality content on the site.

Agent mode: add --agent to any command for JSON output + non-interactive mode.
Health check: run 'wikipedia-pp-cli doctor' to verify auth and connectivity.
See README.md or the bundled SKILL.md for recipes.

Usage:
  wikipedia-pp-cli [command]

Available Commands:
  agent-context Emit structured JSON describing this CLI for agents
  article       Get a Wikipedia article in text, html, or json format
  completion    Generate the autocompletion script for the specified shell
  doctor        Check CLI health
  feed          Manage feed
  feedback      Record feedback about this CLI (local by default; upstream opt-in)
  help          Help about any command
  import        Import data from JSONL file via API create/upsert calls
  on-this-day   Get historical events for a calendar date
  page          Manage page
  profile       Named sets of flags saved for reuse
  random        Get a random Wikipedia article summary
  search        Search Wikipedia articles by keyword
  sql           Run a read-only SQL query against the local Wikipedia cache
  summary       Get a Wikipedia article summary
  sync          Sync API data to local SQLite for offline search and analysis
  version       Print version
  which         Find the command that implements a capability
  workflow      Compound workflows that combine multiple API operations

Flags:
      --agent                Set all agent-friendly defaults (--json --compact --no-input --no-color --yes)
      --compact              Return only key fields (id, name, status, timestamps) for minimal token usage
      --config string        Config file path
      --csv                  Output as CSV (table and array responses)
      --data-source string   Data source for read commands: auto (live with local fallback), live (API only), local (synced data only) (default "auto")
      --deliver string       Route output to a sink: stdout (default), file:<path>, webhook:<url>
      --dry-run              Show request without sending
  -h, --help                 help for wikipedia-pp-cli
      --human-friendly       Enable colored output and rich formatting
      --idempotent           Treat already-existing create results as a successful no-op
      --json                 Output as JSON
      --no-cache             Bypass response cache
      --no-color             Disable colored output
      --no-input             Disable all interactive prompts (for CI/agents)
      --plain                Output as plain tab-separated text
      --profile string       Apply values from a saved profile (see 'wikipedia-pp-cli profile list')
      --quiet                Bare output, one value per line
      --rate-limit float     Max requests per second (0 to disable)
      --select string        Comma-separated fields to include in output (e.g. --select id,name,status)
      --timeout duration     Request timeout (default 30s)
  -v, --version              version for wikipedia-pp-cli
      --yes                  Skip confirmation prompts (for agents and scripts)

Use "wikipedia-pp-cli [command] --help" for more information about a command.
```

---

## `doctor`

```bash
$ ./wikipedia-pp-cli.exe doctor
  OK Config: ok
  OK Auth: not required
  OK API: reachable
  base_url: https://en.wikipedia.org/api/rest_v1
  version: 1.0.0
  INFO Cache: unknown
    schema_version: 1
    stale_after: 6h0m0s
    hint: sync_state is empty; run 'wikipedia-pp-cli sync' to hydrate.
```

All checks passed. No auth required (public API).

---

## `summary` — JSON output

```bash
$ ./wikipedia-pp-cli.exe summary "Alan Turing" --json
```

```json
{
  "type": "standard",
  "title": "Alan Turing",
  "displaytitle": "<span lang=\"en\" dir=\"ltr\"><span class=\"mw-page-title-main\">Alan Turing</span></span>",
  "namespace": {
    "id": 0,
    "text": ""
  },
  "wikibase_item": "Q7251",
  "titles": {
    "canonical": "Alan_Turing",
    "normalized": "Alan Turing",
    "display": "<span lang=\"en\" dir=\"ltr\"><span class=\"mw-page-title-main\">Alan Turing</span></span>"
  },
  "pageid": 1208,
  "thumbnail": {
    "source": "https://upload.wikimedia.org/wikipedia/commons/thumb/c/ce/Alan_turing_header.jpg/330px-Alan_turing_header.jpg",
    "width": 330,
    "height": 440
  },
  "originalimage": {
    "source": "https://upload.wikimedia.org/wikipedia/commons/c/ce/Alan_turing_header.jpg",
    "width": 752,
    "height": 1002
  },
  "lang": "en",
  "dir": "ltr",
  "revision": "1352875803",
  "tid": "f4fd19cc-4982-11f1-8a64-79c78e586ce6",
  "timestamp": "2026-05-06T19:37:03Z",
  "description": "English computer scientist (1912–1954)",
  "description_source": "local",
  "content_urls": {
    "desktop": {
      "page": "https://en.wikipedia.org/wiki/Alan_Turing",
      "revisions": "https://en.wikipedia.org/wiki/Alan_Turing?action=history",
      "edit": "https://en.wikipedia.org/wiki/Alan_Turing?action=edit",
      "talk": "https://en.wikipedia.org/wiki/Talk:Alan_Turing"
    },
    "mobile": {
      "page": "https://en.wikipedia.org/wiki/Alan_Turing",
      "revisions": "https://en.wikipedia.org/wiki/Special:History/Alan_Turing",
      "edit": "https://en.wikipedia.org/wiki/Alan_Turing?action=edit",
      "talk": "https://en.wikipedia.org/wiki/Talk:Alan_Turing"
    }
  },
  "extract": "Alan Mathison Turing was an English mathematician, computer scientist, logician, cryptanalyst, philosopher and theoretical biologist. He was highly influential in the development of theoretical computer science, providing a formalisation of the concepts of algorithm and computation with the Turing machine, which can be considered a model of a general-purpose computer. Turing is widely considered to be the father of theoretical computer science.",
  "extract_html": "<p><b>Alan Mathison Turing</b> was an English mathematician, computer scientist, logician, cryptanalyst, philosopher and theoretical biologist. He was highly influential in the development of theoretical computer science, providing a formalisation of the concepts of algorithm and computation with the Turing machine, which can be considered a model of a general-purpose computer. Turing is widely considered to be the father of theoretical computer science.</p>"
}
```

---

## `summary` — agent mode with `--select`

```bash
$ ./wikipedia-pp-cli.exe summary "Alan Turing" --agent --select title,description
```

```json
{
  "description": "English computer scientist (1912–1954)",
  "title": "Alan Turing"
}
```

`--select` without a value correctly returns an error:

```bash
$ ./wikipedia-pp-cli.exe summary "Alan Turing" --agent --select
Error: flag needs an argument: --select
```

---

## `summary` — disambiguation exit code

```bash
$ ./wikipedia-pp-cli.exe summary "Mercury"
Disambiguation: "Mercury" matches multiple articles. Try a more specific title.
  Example: wikipedia-pp-cli summary "Mercury (topic)"
Error: disambiguation page: Mercury

$ echo "Exit code: $?"
Exit code: 3
```

Exit code `3` on disambiguation pages — correct typed exit code behavior.

---

## `search`

```bash
$ ./wikipedia-pp-cli.exe search "quantum computing" --limit 5
Search results for "quantum computing" (5 found):

1. Quantum computing
   Computer hardware technology that uses quantum mechanics
   quantum computing (abbreviated 'n.quantum computing') is an unconventional process of computing that uses n...

2. Superconducting quantum computing
   Quantum computing implementation
   quantum computing is a branch of quantum computing and solid-state physics that implements superconducting electronic...

3. Timeline of quantum computing and communication
   timeline of quantum computing and communication. Erwin Schrödinger publishes a theorem setting the basis for quantum...

4. Trapped-ion quantum computer
   Proposed quantum computer implementation
   quantum computing began to accelerate worldwide. In 2021, researchers from the University of Innsbru...

5. Institute for Quantum Computing
   Research institute at the University of Waterloo in Ontario, Canada
   The Institute for Quantum Computing (IQC) is an affiliate scientific research institute of the University of Waterloo...
```

---

## `search` — JSON output

```bash
$ ./wikipedia-pp-cli.exe search "black holes" --limit 3 --json
```

```json
{
  "pages": [
    {
      "id": 4650,
      "key": "Black_hole",
      "title": "Black hole",
      "excerpt": "The first widely accepted black hole was Cygnus X-1, identified by several researchers independently in 1971. Black holes typically form when very massive",
      "matched_title": null,
      "description": "Compact astronomical body",
      "thumbnail": {
        "mimetype": "image/jpeg",
        "size": null,
        "width": 60,
        "height": 60,
        "duration": null,
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4f/Black_hole_-_Messier_87_crop_max_res.jpg/60px-Black_hole_-_Messier_87_crop_max_res.jpg"
      }
    },
    {
      "id": 215706,
      "key": "Supermassive_black_hole",
      "title": "Supermassive black hole",
      "excerpt": " and the black hole at the center of Messier 87, a giant elliptical galaxy. Supermassive black holes are classically defined as black holes with a mass",
      "matched_title": null,
      "description": "Largest type of black hole",
      "thumbnail": {
        "mimetype": "image/jpeg",
        "size": null,
        "width": 60,
        "height": 35,
        "duration": null,
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/c/cf/Black_hole_-_Messier_87.jpg/60px-Black_hole_-_Messier_87.jpg"
      }
    },
    {
      "id": 41096027,
      "key": "List_of_most_massive_black_holes",
      "title": "List of most massive black holes",
      "excerpt": "black holes so far discovered (and probable candidates), measured in units of solar masses (M☉), approximately 2×1030 kilograms. A supermassive black",
      "matched_title": null,
      "description": "",
      "thumbnail": {
        "mimetype": "image/jpeg",
        "size": null,
        "width": 60,
        "height": 60,
        "duration": null,
        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/4/4f/Black_hole_-_Messier_87_crop_max_res.jpg/60px-Black_hole_-_Messier_87_crop_max_res.jpg"
      }
    }
  ]
}
```

---

## `random`

```bash
$ ./wikipedia-pp-cli.exe random
== Lautrach ==
Municipality in Bavaria, Germany

Lautrach is a municipality in the district of Unterallgäu in Bavaria, Germany.
```

---

## `on-this-day`

Wrong invocation (flag attached to wrong subcommand):

```bash
$ ./wikipedia-pp-cli.exe random on-this-day --month 5 --day 11 --type selected
Error: unknown flag: --month
```

Correct invocation:

```bash
$ ./wikipedia-pp-cli.exe on-this-day --month 5 --day 11 --type selected
== On This Day: May 11 ==

-- SELECTED --
  2022: Myanmar civil war: Government troops killed 37 unarmed civilians in Mondaingbin.
  2022: Palestinian-American journalist Shireen Abu Akleh was shot and killed while reporting on an Israel Defense Forces raid on the Jenin refugee camp.
  2013: Two car bombs by unknown perpetrators exploded in Reyhanlı, Turkey, resulting in 52 killed and 140 injured.
  2011: An earthquake registering Mw 5.1, the worst to hit the region for more than 50 years, struck near Lorca, Spain.
  2010: David Cameron took office as Prime Minister of the United Kingdom as the Conservatives and Liberal Democrats formed the country's first coalition government since the Second World War.
  2010: Gordon Brown resigned as Prime Minister of the United Kingdom and Leader of the Labour Party after failing to strike a coalition agreement with the Liberal Democrats.
  1998: India began the Pokhran-II nuclear-weapons test, its first since the Smiling Buddha test 24 years earlier.
  1997: Deep Blue defeated Garry Kasparov in six games to become the first chess computer to win a match against a world champion.
  1981: Andrew Lloyd Webber's Cats opened at the New London Theatre.
  1970: Lubbock, Texas, was struck by a tornado that left 26 people dead.
  1963: African Americans rioted in Birmingham, Alabama, in response to two bombings, perceiving local police to be complicit with the perpetrators.
  1928: After a week-long standoff punctuated by military clashes, Japanese forces captured the city of Jinan, Shandong in China.
  1910: Glacier National Park was established in the U.S. state of Montana.
  1894: In response to a 28-percent wage cut, 4,000 Pullman Palace Car Company workers went on strike in Illinois, bringing rail traffic west of Chicago to a halt.
  1889: Bandits attacked a U.S. Army paymaster's escort in the Arizona Territory, stealing more than $28,000.
  1880: A land dispute between the Southern Pacific Railroad and settlers in Hanford, California, turned deadly when a gun battle broke out, leaving seven dead.
  1858: Minnesota was carved out of the eastern half of the Minnesota Territory and admitted as the 32nd U.S. state.
  1820: HMS Beagle, the ship that would take Charles Darwin on his voyage, was launched.
  1813: William Lawson, Gregory Blaxland and William Wentworth departed westward from Sydney on an expedition to become the first confirmed Europeans to cross the Blue Mountains.
  1812: Spencer Perceval was shot in the lobby of the House of Commons, becoming the only British prime minister to be assassinated.
  1745: War of the Austrian Succession: French forces defeated those of the Pragmatic Allies at the Battle of Fontenoy in the Austrian Netherlands in present-day Belgium.
  868: A copy of the Diamond Sutra was printed in Tang-dynasty China, making it the world's oldest dated printed book.
```

---

## Summary

| Command | Result |
|---------|--------|
| `build` | ✅ Pass |
| `--version` | ✅ Pass |
| `--help` | ✅ Pass |
| `doctor` | ✅ Pass — API reachable, no auth required |
| `summary --json` | ✅ Pass |
| `summary --agent --select` | ✅ Pass |
| `summary` disambiguation | ✅ Pass — exit code 3 |
| `search --limit` | ✅ Pass |
| `search --json` | ✅ Pass |
| `random` | ✅ Pass |
| `on-this-day` | ✅ Pass |
| Wrong subcommand flag | ✅ Correct error returned |
