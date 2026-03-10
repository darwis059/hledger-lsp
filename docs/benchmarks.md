# hledger-lsp Benchmark Results

Benchmarks run on Apple M4 Pro, macOS, Go 1.26.

## NFR Targets

| NFR | Target | Measured | Status |
| --- | --- | --- | --- |
| NFR-1.1 | Completion < 100ms | ~4.2ms | Pass |
| NFR-1.2 | Parsing 10k lines < 500ms | ~11ms | Pass |
| NFR-1.3 | Incremental updates < 50ms | ~2.7ms | Pass |
| NFR-1.4 | Memory < 200MB | ~31MB | Pass |

All NFR targets are validated by automated tests in `internal/benchmark/nfr_test.go`.

## Parser Benchmarks

| Benchmark | Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| Lexer_Small | 10 | 4,784 | 68 | 17 |
| Lexer_Medium | 100 | 46,774 | 644 | 161 |
| Lexer_Large | 1,000 | 468,776 | 6,404 | 1,601 |
| Lexer_XLarge | 10,000 | 4,743,745 | 64,004 | 16,001 |
| Parser_Small | 10 | 10,625 | 24,208 | 153 |
| Parser_Medium | 100 | 105,585 | 245,986 | 1,485 |
| Parser_Large | 1,000 | 1,028,232 | 2,416,411 | 14,805 |
| Parser_XLarge | 10,000 | 10,411,240 | 24,128,694 | 148,005 |

## Analyzer Benchmarks

| Benchmark | Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| Analyze_Small | 10 | 23,778 | 37,004 | 407 |
| Analyze_Medium | 100 | 218,812 | 332,751 | 2,961 |
| Analyze_Large | 1,000 | 2,343,398 | 3,837,194 | 27,935 |
| CheckBalance | 1 tx | 696 | 1,577 | 18 |
| CheckBalance_MultiCommodity | 1 tx | 600 | 1,360 | 16 |
| CollectAccounts | 1,000 | 36,842 | 3,568 | 67 |
| CollectPayees | 1,000 | 72,857 | 143,944 | 31 |
| CollectCommodities | 1,000 | 24,763 | 112 | 3 |
| CollectTags | 1,000 | 15,591 | 0 | 0 |

## Workspace Index Benchmarks

| Benchmark | Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| BuildFileIndex_Small | 10 | 28,482 | 43,428 | 487 |
| BuildFileIndex_Medium | 100 | 270,820 | 397,875 | 3,908 |
| BuildFileIndex_Large | 1,000 | 2,750,533 | 4,182,682 | 37,817 |
| BuildFileIndex_XLarge | 10,000 | 27,715,848 | 41,002,139 | 376,532 |
| UpdateFile | any | ~16 | 0 | 0 |
| IndexSnapshot | any | ~403,880 | 1,600,520 | 5,986 |

## Include Loader Benchmarks

| Benchmark | Files/Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| Load_Small | 1 file / 10 tx | 25,278 | 27,368 | 163 |
| Load_Medium | 1 file / 100 tx | 130,278 | 271,418 | 1,495 |
| Load_Large | 1 file / 1000 tx | 1,201,359 | 2,646,643 | 14,815 |
| LoadFromContent_Small | 1 file / 10 tx | 11,211 | 24,320 | 155 |
| LoadFromContent_Large | 1 file / 1000 tx | 1,140,531 | 2,416,521 | 14,807 |
| IncludeTree_5Files | 5 files / 100 tx | 17,249 | 4,635 | 43 |
| IncludeTree_10Files | 10 files / 200 tx | 21,293 | 8,122 | 68 |
| IncludeTree_20Files | 20 files / 400 tx | 28,248 | 15,081 | 112 |

## Incremental Update Benchmarks

These benchmarks measure the full incremental update cycle when a document changes:

| Benchmark | Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| DidChange_Incremental_Small | 10 | 51,971 | 779,589 | 9 |
| DidChange_Incremental_Medium | 100 | 26,939 | 321,605 | 9 |
| DidChange_Incremental_Large | 1,000 | 57,886 | 274,313 | 9 |
| PublishDiagnostics_Small | 10 | 45,967 | 98,383 | 764 |
| PublishDiagnostics_Medium | 100 | 394,201 | 883,791 | 6,203 |
| PublishDiagnostics_Large | 1,000 | 4,115,476 | 9,129,770 | 59,980 |

**Components of incremental update:**

1. `DidChange` (sync): Apply text change, update workspace index, invalidate cache (~27-58us)
2. `PublishDiagnostics` (async): Parse, analyze, publish diagnostics (~46us - 4.1ms)

Full cycle for 1000 transactions: ~4.2ms (well under NFR-1.3 target of 50ms)

## Server Benchmarks

| Benchmark | Transactions | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| Completion_Account_Small | 10 | 35,073 | 80,335 | 619 |
| Completion_Account_Medium | 100 | 297,439 | 652,700 | 4,507 |
| Completion_Account_Large | 1,000 | 3,548,146 | 6,894,203 | 42,816 |
| Completion_Payee | 1,000 | 4,024,686 | 8,385,182 | 50,479 |
| Completion_Commodity | 1,000 | 3,587,262 | 6,808,426 | 42,756 |
| ApplyChange_Small | 10 | 605 | 2,336 | 4 |
| ApplyChange_Large | 1,000 | 54,746 | 229,440 | 4 |

## Running Benchmarks

```bash
# All benchmarks
go test ./... -bench=. -benchmem

# Specific package
go test ./internal/workspace/... -bench=. -benchmem

# With count for statistical significance
go test ./internal/parser/... -bench=. -benchmem -count=5

# NFR validation tests
go test ./internal/benchmark/... -v -run TestNFR

# With profiling
go test ./internal/parser/... -bench=BenchmarkParser_XLarge -cpuprofile=cpu.prof -memprofile=mem.prof
go tool pprof -http=:8080 cpu.prof
```

## Key Observations

1. **Parser scaling**: Linear with transaction count (~1.0us per transaction)
2. **Memory efficiency**: ~4.1KB per transaction for full index
3. **Include tree**: Minimal overhead for multi-file journals (~17-28us for 5-20 files)
4. **Incremental updates**: ~4.2ms for 1000 transactions (full cycle including diagnostics)
5. **Completion latency**: ~3.5ms for 1000 transactions
6. **Workspace UpdateFile**: Sub-microsecond with zero allocations (deferred indexing)
7. **Analyzer**: ~2.3ms for 1000 transactions; balance check sub-microsecond per transaction
8. **CollectTags**: Zero allocations — fully stack-based iteration
