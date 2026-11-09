# 01: Ledger skeleton — record a loan and list it

**What to build:** A lender can record a loan in one command and see it listed. `qarz add --to Ali --amount 50000 --due 2026-12-15` writes the loan and prints its number; `qarz list` shows every open loan with borrower, amount, and due date. No calendar involvement of any kind — this ticket delivers a usable ledger on its own.

The full schema lands here, including the columns later tickets need (`qarz_uid`, `voided_at`), so that no later ticket has to migrate a column across the whole codebase. Foreign keys and a busy timeout are enabled in the connection string, because SQLite ignores `REFERENCES` without the pragma.

**Blocked by:** None (can start immediately).

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] `qarz add` accepts a borrower, an amount in whole rupees, a due date, and optional date of lending, note, and repeated key-value details
- [ ] Date of lending defaults to today when not given
- [ ] Amounts are stored as integer minor units; the lender never sees or types minor units
- [ ] Each loan gets a small sequential number, assigned by the database, and a separate random globally-unique tag
- [ ] The loan number is printed on success
- [ ] A due date earlier than the date of lending is rejected with a non-zero exit
- [ ] A non-positive amount is rejected with a non-zero exit
- [ ] `qarz list` prints open loans; output is covered by a golden file
- [ ] The ledger is a single file under the lender's home directory, created on first use
- [ ] Tests drive the root command against a temporary database and assert on output, exit status, and resulting rows
