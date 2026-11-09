# 11: Void and restore

**What to build:** A lender can make a loan go away without erasing it. `qarz void 7` removes its calendar event and drops it from the default listing, while the loan and every repayment against it stay in the database and remain visible when asked for directly. `qarz restore 7` reverses it. Nothing in the tool destroys a row, so voiding is safe enough not to need confirmation ceremony.

This covers both the loan entered by mistake and the one written off, without either pretending the money came back.

**Blocked by:** 05.

**Status:** ready-for-agent

- [ ] `qarz void <id>` hides the loan from the default listing and removes its calendar event
- [ ] A voided loan keeps every repayment recorded against it
- [ ] `qarz show` still displays a voided loan, marked as void, with the date it was voided
- [ ] `qarz restore <id>` returns the loan to its derived status and recreates its event
- [ ] `qarz list --void` lists voided loans and `--all` includes them alongside the rest
- [ ] Voiding an already-voided loan, or restoring a live one, reports the situation without a non-zero exit
- [ ] Repair removes the events of voided loans and does not recreate them
