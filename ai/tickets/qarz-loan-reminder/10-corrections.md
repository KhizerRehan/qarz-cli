# 10: Corrections

**What to build:** A lender can fix a loan that was entered wrong without destroying it. `qarz edit 7` changes the amount, the due date, the note, or the borrower's name, and the calendar event follows. Correcting a mistyped amount no longer means deleting the loan and re-entering it along with every repayment against it.

The amount has a floor: a new principal below what has already been repaid would produce a negative balance and a status with no meaning, so it is refused. Exactly equal is allowed, and lands the loan on paid.

**Blocked by:** 05.

**Status:** ready-for-agent

- [ ] `qarz edit <id>` changes any of amount, due date, note, and borrower, individually or together
- [ ] The calendar event is reprojected after the change
- [ ] A new amount below the sum of existing repayments is rejected with a non-zero exit, naming both numbers
- [ ] A new amount exactly equal to the sum of repayments is accepted and leaves the loan paid
- [ ] A new due date earlier than the date of lending is rejected
- [ ] Changing the due date on an overdue loan to a future date returns the event to that date and restores its normal reminders
- [ ] Editing with no fields given is rejected rather than performing a silent no-op
