# hledger-lsp Project Overview

## Purpose and Context

hledger-lsp is an LSP server for hledger journal files written in Go. It provides IDE-like features (completion, diagnostics, formatting, etc.) to any LSP-compatible editor, eliminating the dependency on the VS Code extension and making hledger convenient in Neovim, Emacs, Helix, and others.

Target audience:

- hledger users in Neovim/Emacs/Helix and other editors;
- teams that need consistent validation and formatting rules.

## Implemented Features (per README/tasks)

- Completion: accounts, payees, commodities.
- Diagnostics: balance, syntax, basic checks.
- Formatting: amount alignment and indentation (full, range, onType).
- Hover: balances and transaction details.
- Semantic tokens: syntax highlighting.
- Document symbols: transactions, directives, include.
- Include: path resolution, cycle detection.
- CLI code actions: running hledger reports.

## Architecture and Modules

Key directories:

- `cmd/` — server entry point.
- `internal/parser` — lexer/parser and AST.
- `internal/analyzer` — semantics, balance, indexing.
- `internal/server` — LSP handlers.
- `internal/formatter` — formatting.
- `internal/include` — include resolution.
- `internal/cli` — hledger CLI wrapper.
- `internal/workspace` — cross-file data aggregation.

Processing pipeline: parsing → analysis → diagnostics/highlighting/hover/completion.

## Symbol Index Model and Data Sources

Currently indexed:

- accounts: `AccountIndex` with `All` and `ByPrefix` for prefix-based completion;
- payees: unique values from transactions;
- commodities: directives and amounts/costs in postings;
- tags: tags from AST (tags from comments are not yet extracted).

Data sources:

- AST from `internal/parser` (account/commodity directives, transactions, and postings);
- include tree via `internal/include.Loader` and `ResolvedJournal` (Primary + Files).

Analysis separation:

- `Analyze` processes a single `Journal`;
- `AnalyzeResolved` aggregates symbols across the include tree, merging data from all files.

Workspace caches:

- declared accounts/commodities and number formats from commodity directives;
- used for fast checks and formatting without full recomputation.

Limitations:

- no transaction index or relationship tracking for references/duplicates (groundwork for tasks 2–4).

## Gaps and Incomplete Features

- Duplicate transaction diagnostics — deferred.

## Risks and Current Mitigations (per PRD)

Risks:

- performance on large files;
- hledger format complexity;
- CLI compatibility.

Mitigations already in place:

- benchmarks and incremental updates;
- hand-written parser and tests;
- include cycle detection;
- lint/CI and defensive checks.

## Improvement Suggestions

- Go to Definition/Find References: symbol index in `internal/analyzer`/`internal/workspace` + new LSP handlers in `internal/server`.
- Completion for tags/dates/templates: extend completion provider and context rules.
- Duplicate transactions: detection via transaction hashing in the analyzer.
- Workspace-wide completion: shared index across the include tree.
- Performance: expand benchmarks, introduce config limits (max results, file size, include depth).
