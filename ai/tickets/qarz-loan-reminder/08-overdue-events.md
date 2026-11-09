# 08: Overdue events keep surfacing

**What to build:** A loan whose due date has passed stops going quiet. Its event is moved forward to the current day and retitled to state how many days late it is, so it lands on today's agenda instead of receding into last month, with a single popup at the slot time — the day-before reminder cannot fire on an event that has been moved to today. The lender can also list overdue loans directly.

The event's date is derived, not stored: the desired date for any loan is the later of its due date and the current day. This keeps repair a pure comparison between a computed desired state and the observed event, and avoids a column whose only job would be remembering a value that can be computed.

**Blocked by:** 06.

**Status:** ready-for-agent

- [ ] An open loan past its due date projects to the current day rather than its due date
- [ ] Its title states the number of days it is late
- [ ] Its reminders are reduced to a single popup at the slot time
- [ ] The loan's recorded due date is never altered by this move
- [ ] A loan that is not yet due projects to its due date, unchanged
- [ ] A paid loan past its due date is not moved
- [ ] `qarz list --overdue` filters to open loans past their due date
- [ ] A test syncs the same overdue loan on two different simulated days and asserts the event moved both times with the day count matching
