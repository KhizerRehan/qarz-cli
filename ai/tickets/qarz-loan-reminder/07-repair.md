# 07: Repair

**What to build:** A lender whose calendar has drifted can restore it from the ledger in one command. `qarz sync --repair` reconciles in both directions: events whose loan no longer exists or has been voided are removed, loans whose event is missing get one created, and events whose title or date no longer matches the ledger are rewritten. Running it twice is never worse than running it once.

Matching is on the loan's unique tag, not the stored event identifier, so repair recognises its own events even if the stored identifier was lost — and leaves every other event in the calendar alone.

**Blocked by:** 06.

**Status:** ready-for-agent

- [ ] `qarz sync --repair` deletes tagged events whose loan is absent from the ledger
- [ ] It creates events for loans that should have one and do not
- [ ] It rewrites events whose title or date has drifted from the ledger
- [ ] It never modifies or deletes an event that carries no tag
- [ ] A second consecutive run produces an identical calendar and reports no changes
- [ ] It repairs a loan whose stored event identifier is missing by finding the event through its tag
- [ ] Repair reports what it changed, grouped by action
