# Contributing to qarz

Thanks for taking the time to help. Bug reports, fixes, documentation and ideas are all welcome.

## Before you start

- Read the [README](README.md) for what qarz does and, just as important, what it deliberately does not do.
- Read the [GLOSSARY](GLOSSARY.md). Code, flags, messages and docs use its words (loan, borrower, principal, repayment, balance) and avoid the ones it lists under _Avoid_.
- For anything larger than a small fix, open an issue first so we can agree on the shape before you write code.

## Design rules that are not up for negotiation

These are what make qarz trustworthy as a ledger. A change that breaks one of them will not be merged.

1. **The ledger is the only source of truth, and it stays on your machine.** qarz makes no network calls.
2. **Money is integer minor units (paisa).** No floating point anywhere money is handled.
3. **Balances are derived, never stored.**
4. **Nothing is deleted.** Loans and repayments accumulate; a loan becomes paid only through repayments that sum to its principal.
5. **No interest, markup or fees.** A qarz is interest-free by definition.

## Development setup

You need Go (see the `go` line in [go.mod](go.mod) for the minimum version) and git.

```sh
git clone https://github.com/KhizerRehan/qarz-cli.git
cd qarz-cli
go build -o qarz ./cmd/qarz
go test ./...
```

To try your build without touching your real ledger, point `HOME` at a scratch directory:

```sh
HOME=$(mktemp -d) ./qarz add --to Test --amount 1000 --due 2027-01-01
```

## Tests

Tests sit at two seams: the command line, driven against a temporary ledger (`internal/cli`), the ledger's storage and migrations (`internal/store`), and the domain arithmetic as table tests (`internal/loan`).

Command-line output is checked against golden files in `internal/cli/testdata/`. When you change output on purpose, regenerate them and review the diff before committing:

```sh
go test ./internal/cli -update
git diff internal/cli/testdata
```

Every behaviour change needs a test. Every bug fix needs a test that fails without the fix.

## Code style

- Run `gofmt` (or `go fmt ./...`) before committing. CI also runs `go vet` and `golangci-lint`.
- Keep packages under `internal/`. qarz is a command, not a library.
- Error messages are lowercase, say what failed, and where possible tell the user what to do next (`run 'qarz list --all'`).
- Prefer clear names over comments.

## Commits and pull requests

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(repay): accept a backdated repayment date
fix(list): sort overdue loans by days late
docs: explain how to back up the ledger
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`. Keep the subject imperative, lowercase and under 50 characters.

For a pull request:

1. Fork the repository and create a branch from `main`.
2. Make your change with tests, and update the README if user-facing behaviour changed.
3. Add a line under `Unreleased` in [CHANGELOG.md](CHANGELOG.md).
4. Make sure `go test ./...` passes.
5. Open the pull request and fill in the template.

## Reporting security issues

Do not open a public issue. See [SECURITY.md](SECURITY.md).

## Code of conduct

Everyone taking part is expected to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
