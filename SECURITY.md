# Security Policy

qarz stores personal financial records and a Google OAuth refresh token on your machine, so security reports are taken seriously.

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

- Exposure of the ledger (`~/.qarz/qarz.db`), the OAuth client (`~/.qarz/credentials.json`) or the token (`~/.qarz/token.json`), for example through wrong file permissions
- Flaws in the local OAuth sign-in flow (redirect listener, state validation, token handling)
- qarz reading, changing or deleting calendar events it did not create
- Anything that lets data leave the machine other than the calendar events qarz is meant to write

## What is out of scope

- Someone who already has access to your user account reading files in your home directory
- Vulnerabilities in Google's services themselves
- Leaks caused by sharing your own `credentials.json` or `token.json`

## Keeping your data safe

- qarz asks Google for the `calendar` scope only.
- The files in `~/.qarz/` are written readable by you alone. Keep them that way, and never commit them or share them.
- To revoke qarz's access, remove it at <https://myaccount.google.com/permissions> and delete `~/.qarz/token.json`.
