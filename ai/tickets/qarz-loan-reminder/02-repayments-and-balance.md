# 02: Repayments and derived balance

**What to build:** A lender can record money coming back, and the ledger does the arithmetic. `qarz repay 7 --amount 20000` records a repayment; the listing then shows the remaining balance rather than the original principal. When repayments sum to the principal the loan becomes paid automatically, with nothing marked by hand. A repayment larger than what is still outstanding is refused.

Balance is derived on every read, never stored. Repayments are append-only: nothing deletes or edits one, because the repayment history is what proves a loan was paid.

**Blocked by:** 01.

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] `qarz repay <id> --amount` records a repayment against an existing loan
- [ ] The repayment date defaults to today and is settable
- [ ] An optional note can be attached to a repayment
- [ ] The listing shows balance, not principal, and reflects partial repayments
- [ ] A loan with no repayments still appears in the listing with its full principal
- [ ] A loan reaches paid status exactly when its repayments sum to its principal
- [ ] A repayment above the outstanding balance is rejected with a non-zero exit, naming both the amount given and the balance available
- [ ] An unknown loan number is rejected with a non-zero exit
- [ ] A non-positive repayment amount is rejected
- [ ] `qarz list --paid` filters to paid loans
- [ ] Table-driven tests on the domain arithmetic cover zero repayments, a partial repayment, an exact-principal repayment, and one minor unit above the balance
- [ ] A test asserts that inserting a repayment against a non-existent loan fails, which is what keeps the foreign-key pragma from regressing into a silent no-op
