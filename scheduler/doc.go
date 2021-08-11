// Package scheduler
// Flyte scheduler implementation that allows to schedule fixed rate and cron schedules on sandbox deployment
// Scheduler has two components
// 1] Schedule management
//    This component is part of the pkg/async/schedule/flytescheduler package
//    Role of this component is to create / activate / deactivate schedules
//    The above actions are exposed through launchplan activation/deactivation api's and donot have separate controls.
//    Whenever a launchplan with a schedule is activated, a new schedule entry is created in the datastore
//    On deactivation the created scheduled and launchplan is deactivated through a flag
//    Atmost one launchplan is active at any moment across its various versions and same semantics apply for the
//    schedules aswell.
// 2] Scheduler
//    This component is a singleton and has its source in the current folder and is responsible for reading the schedules
//    from the DB and running them at the cadence defined by there schedule
//    The lowest granularity supported is secs for scheduling through cron and minutes using the fixed rate scheduler
// 	  The scheduler should be running in one replica , two at the most during redeployment. Multiple replicas will just
// 	  duplicate the work since each execution for a scheduleTime will have unique identifier derived from schedule name
//	  and time of the schedule. The idempotency aspect of the admin for same identifier prevents duplication on the admin
//	  side.
//    The scheduler runs continously in a loop reading the updated schedule entries in the data store and adding or removing
//    the schedules. Removing a schedule doesn't gurantee about inflight routines launched by the scheduler.
//    Sub components:
//		a) Snapshoter
// 			This component is responsible for writing the snapshot state of all the schedules at a regular cadence to a
//			persistent store. The current implementation uses DB to store the GOB format of the snapshot which is versioned.
//			The snapshot is map[string]time.Time which stores a map of schedules names to there last execution times
// 			During bootup the snapshot is bootstraped from the data store and loaded in memory
//			The Scheduler use this snapshot to schedule any missed schedules.
//
//			We cannot use global snapshot time since each time snapshot doesn't contain information on how many schedules
//			were executed till that point in time. And hence the need to maintain map[string]time.Time of schedules to there
//			lastExectimes
// 		b) Catchuper :
//			This component runs at bootup and catches up all the schedules to there current time.Now()
//			The scheduler is not run until all the schedules have been caught up.
//			The current design is also not to snapshot until all the schedules are caught up.
//			This might be drawback in case catch up runs for a long time and hasn't been snapshotted.(reassess)
//		c) GOGFWrapper :
//			This component is responsible for locking in the time for the scheduled job to be invoked and adding those
//			to the GOGF scheduler. Right now this uses https://github.com/gogf/gf framework for fixed rate and cron
// 			schedules
// 			The current implementation uses the lastExecTime from the snapshot to compute the next scheduled time to be
//			sent for execution. This same time is then used as lastExecTime when the jobfunction is invoked again
//		d) Job function :
//			The job function accepts the scheduleTime and the schedule which is used for creating an execution request
//			to the admin. Each job fuction is tied to schedule which gets executed in separate go routine by the gogf
// 			framework in according the schedule cadence.

package scheduler
