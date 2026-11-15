# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `qarz add` to record a loan, with an optional lending date, note and repeatable `--meta key=value` details.
- `qarz repay` to record repayments, refusing any that exceed the balance and marking a loan paid when repayments reach the principal.
- `qarz list` with `--overdue`, `--paid` and `--all` filters, and `qarz show` for one loan and its repayment history.
- `qarz --version`.

[Unreleased]: https://github.com/KhizerRehan/qarz-cli/commits/main
