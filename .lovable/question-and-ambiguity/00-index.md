# Ambiguity & Inference Log — No-Questions Mode

**Mode active**: AI proceeds with best-inference for the next 40 tasks. No clarifying questions are asked. Each ambiguity is logged here as a numbered file for later review.

**Resumption trigger**: User says "ask question" → resume normal clarifying-question flow.

## Index

| # | File | Task | Inference made |
|---|------|------|----------------|
| 01 | [01-json-schema-docs-scope.md](01-json-schema-docs-scope.md) | Generate JSON schema docs for each JSON output | Narrow scope: only stablejson-backed outputs (today: `startup-list --json`); JSON Schema 2020-12 + `propertyOrder` extension; hand-written; contract test guards drift; remaining 20 outputs tracked in `_TODO.md` |
| 02 | [02-cmd-test-helper-duplicates.md](02-cmd-test-helper-duplicates.md) | (discovered during 01) Pre-existing duplicate helpers in `gitmap/cmd/` test files | Left existing files untouched; logged for separate cleanup task |
| 03 | [03-clone-from-scope.md](03-clone-from-scope.md) | "Add a `gitmap clone` that reads JSON/CSV" — but `gitmap clone` already exists | Added new sibling subcommand `gitmap clone-from <file>` (alias `cf`) instead of mutating existing `gitmap clone` |

## How to read each entry

Each `xx-brief-title.md` file contains:
1. **Original task** — verbatim user request + reference to the original spec/prompt
2. **Ambiguity** — the specific point of confusion
3. **Options considered** — every reasonable interpretation with pros/cons
4. **Recommendation** — best option with rationale
5. **Decision taken** — what the AI actually implemented (so user can confirm or override)

## Counter

Tasks consumed: 1 / 40 (entry 02 was discovered during entry 01, not a separate user task)
