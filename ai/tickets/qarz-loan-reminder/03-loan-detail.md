# 03: Loan detail

**What to build:** A lender can reconstruct the whole history of one loan. `qarz show 7` prints the borrower, principal, balance, status, the date of lending and the due date, the note, any extra details recorded against it, and every repayment with its date and note in order.

**Blocked by:** 02.

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] `qarz show <id>` prints the loan's full detail including every repayment
- [ ] Repayments are listed oldest first with their dates and notes
- [ ] Extra key-value details recorded at add time are displayed
- [ ] Status is shown as derived: open, paid, or overdue
- [ ] An unknown loan number is rejected with a non-zero exit
- [ ] Output is covered by golden files, including a loan with no repayments and a loan that is fully paid
