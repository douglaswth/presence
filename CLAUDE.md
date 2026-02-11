# CLAUDE.md

## Build & Test
- `go test ./...` — run all tests
- `go test -race ./...` — run with race detection
- `go test -coverprofile=coverage.out ./...` — coverage report
- Go 1.24 — no `tc := tc` needed in test loops

## Testing Conventions
- **Table-driven tests**: `cases` slice, `tc` variable, `t.Parallel()` in every subtest
- **Mock framework**: Clue Mock Generator (`cmg`), mocks live in `*/mocks/` directories
  - `Add*()` queues expectations (FIFO), `Set*()` replaces
  - Always verify with `assert.False(mock.HasMore(), "missing expected ... calls")`
- **Setup pattern**: optional `setup func(t *testing.T, ...)` field in test table, nil-checked before calling
- **Assertions**: `github.com/stretchr/testify/assert`; use `assert := assert.New(t)` in subtests
- **Context**: use `log.Context(context.Background(), log.WithDebug())` when testing code that calls `goa.design/clue/log`
- **Test fixtures**: YAML files in `tests/` directory
- **Naming**: `TestFunctionName` or `TestType_Method`

## Project Structure
- Top-level package `presence`: config parsing, detector logic
- `ifttt/`: IFTTT webhook client
- `neighbors/`: ARP presence detection, state machine
- `wrap/`: thin wrappers around `net` stdlib (for mocking)
- `cmd/presence/`: CLI entrypoint
