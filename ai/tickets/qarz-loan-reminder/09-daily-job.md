# 09: Daily background job

**What to build:** Overdue reminders arrive without the lender doing anything. Setup installs a daily background job that runs the sync, so an overdue loan keeps surfacing even if the lender has not opened a terminal for a month. Without this, the forward move only happens when the lender is already at the keyboard — which is the gap the reminders exist to close.

**Blocked by:** 08.

**Status:** ready-for-agent

- [ ] `qarz setup` installs a launchd job that runs the sync once a day
- [ ] The job is installed for the current user and survives a reboot
- [ ] Installing it twice does not create a duplicate job
- [ ] The job's output and errors go to a log file under the lender's qarz directory
- [ ] A command reports whether the job is currently installed
- [ ] The job can be removed
