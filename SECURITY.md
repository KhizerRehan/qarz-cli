# Security Policy

qarz stores personal financial records on your machine, so security reports are taken seriously.

## Supported versions

Only the latest release receives security fixes.

## Reporting a vulnerability

**Please do not report security issues through public GitHub issues, discussions or pull requests.**

Report privately through GitHub's private vulnerability reporting:

1. Go to the [Security tab](https://github.com/KhizerRehan/qarz-cli/security) of this repository.
2. Choose **Report a vulnerability**.

Include as much of the following as you can:

- what the issue is and what an attacker could do with it
- the qarz version (`qarz --version`) and your operating system
- steps to reproduce, or a proof of concept
- any suggested fix

You can expect an acknowledgement within 7 days. Once the issue is confirmed, a fix will be prepared and released, and you will be credited in the release notes unless you ask not to be.

## What is in scope

- Exposure of the ledger (`~/.qarz/qarz.db`) or the settings (`~/.qarz/config.toml`), for example through wrong file permissions
- Anything that lets data leave the machine; qarz makes no network calls

## What is out of scope

- Someone who already has access to your user account reading files in your home directory
- Leaks caused by sharing your own ledger file

## Keeping your data safe

- The `~/.qarz/` directory is created readable by you alone. Keep it that way, and never commit or share the files in it.
- Earlier builds that projected loans onto Google Calendar left `~/.qarz/credentials.json` and `~/.qarz/token.json` behind. qarz no longer reads them: revoke its access at <https://myaccount.google.com/permissions> and delete both files.
