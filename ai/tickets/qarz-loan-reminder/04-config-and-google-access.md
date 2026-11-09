# 04: Configuration and Google access

**What to build:** A lender runs one guided command and ends up with working Google access. `qarz setup` walks the steps only a human can perform — creating the cloud project, configuring the consent screen, producing a desktop client identifier — pausing and validating between them rather than describing them and hoping. It then completes the OAuth flow in the browser and stores the resulting token locally with owner-only permissions. It also writes the configuration file holding the target calendar, the timezone, and the currency.

No projection happens in this ticket. Success is a token and a config file on disk that a later ticket can pick up.

**Blocked by:** 01.

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] `qarz setup` walks the Google credential steps one at a time, validating what the lender pasted before moving on
- [ ] The OAuth desktop flow completes and the refresh token is persisted
- [ ] The token file is written readable and writable by its owner only
- [ ] A configuration file is created with the calendar, timezone, and currency, defaulting to the primary calendar, Asia/Karachi, and PKR
- [ ] Re-running setup against existing credentials does not destroy a working token
- [ ] Every command reads the currency from configuration rather than hardcoding it
- [ ] A missing or malformed configuration file produces a message naming what to run, not a crash
