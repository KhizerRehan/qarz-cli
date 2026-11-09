# 06: Replay pending projections

**What to build:** A lender who was offline can catch the calendar up in one command. `qarz sync` finds every loan whose projection previously failed, pushes it, and clears its pending state. A backlog of several loans clears in one run.

**Blocked by:** 05.

**Status:** done — code landed, acceptance criteria unverified (build and tests not run)

- [ ] `qarz sync` projects every pending loan and reports how many were pushed
- [ ] A loan that was pending a create gets its event created and its event identifier stored
- [ ] A loan that was pending an update gets its existing event rewritten
- [ ] A loan that was pending a deletion gets its event removed
- [ ] A still-failing projection stays pending and the command reports it without a non-zero exit
- [ ] Running sync with nothing pending succeeds and says so
- [ ] A test records a failed projection, then asserts that a subsequent sync clears the pending state
