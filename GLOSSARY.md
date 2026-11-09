# Qarz

A personal ledger for interest-free loans you have made to other people, projected one-way onto Google Calendar so that due dates arrive as reminders rather than as something you have to remember to check.

## Language

**Loan**:
Money you handed to one borrower, due back in full on a single date, with no interest, markup, or accrual of any kind. Only loans you made are recorded; money you owe other people is outside this context.
_Avoid_: Debt, qarz (as a record name), IOU, credit

**Borrower**:
The person who received a loan and owes it back to you.
_Avoid_: Counterparty, debtor, lendee, client

**Principal**:
The full amount originally handed over, fixed for the life of the loan. Because there is no interest, it is the most that can ever be owed.
_Avoid_: Amount, total, face value

**Repayment**:
A single payment a borrower made against one loan, on one date. Repayments accumulate and are never removed — the history of a loan is the proof it was paid.
_Avoid_: Payment, installment, settlement, credit

**Balance**:
What a borrower still owes on a loan: principal minus the sum of its repayments. Always derived, never stored.
_Avoid_: Outstanding, remaining, amount due

**Open**:
A loan whose balance is above zero.
_Avoid_: Active, unpaid, outstanding

**Paid**:
A loan whose balance has reached exactly zero. A loan cannot be paid by writing it off, forgiving it, or editing it — only by repayments that sum to the principal.
_Avoid_: Settled, closed, cleared, complete

**Overdue**:
An open loan whose due date has already passed.
_Avoid_: Late, defaulted, delinquent

**Projection**:
The one-way push of a loan's current state onto Google Calendar as a single event. The ledger is the only authority; the calendar is never read back as truth.
_Avoid_: Sync (as a noun for the data relationship), mirror, two-way sync
