# 05: Projection onto the calendar

**What to build:** Every loan now appears in the lender's calendar as exactly one event, kept in step with the ledger. Recording a loan creates the event; recording a repayment retitles it with the new balance; the final repayment marks it paid, swaps the balance for the principal, and clears its reminders so it stops alerting but stays as a record. Three loans to the same person remain three distinguishable events.

A calendar failure never costs the lender a record: the loan is committed first, and a failed projection leaves the loan in a pending state, prints a warning naming the command that will retry it, and exits zero.

The projection depends on a Calendar interface rather than the Google client directly. A fake in-memory implementation satisfies it, and that is what keeps every test off the network.

**Blocked by:** 02, 04.

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] One event per loan, created on add and updated in place thereafter — never a second event for the same loan
- [ ] An open loan's event title carries the borrower and the current balance
- [ ] The event sits on the due date in a morning slot, with popup reminders one day before and shortly before
- [ ] The event description carries the note, the extra details, the principal, the date of lending, and the loan number
- [ ] Every event is stamped with the loan's unique tag in a private property
- [ ] A balance change retitles the event; a due date change moves it
- [ ] The final repayment marks the event paid, shows the principal, and clears reminders
- [ ] A calendar failure leaves the record committed, the loan pending, a warning printed, and the exit status zero
- [ ] Tests run against the fake Calendar and assert on the events it recorded; no test performs network access
