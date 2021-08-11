package interfaces

// GoGFJobWrapper Wrapper interface to add or remove jobs from the scheduler
type GoGFJobWrapper interface {
	// ScheduleJob Wrapper method to schedule cron or fixed interval jobs
	ScheduleJob()
	// DeScheduleJob Wrapper method to deschedule cron or fixed interval jobs
	DeScheduleJob()
	// AddCronJob adds a cron job to the scheduler
	AddCronJob() error
	// AddFixedIntervalJob adds a fixed interval job to the scheduler
	AddFixedIntervalJob() error
	// RemoveCronJob removes a cron job from the scheduler
	RemoveCronJob()
	// RemoveFixedIntervalJob removes a fixed interval job from the scheduler
	RemoveFixedIntervalJob()
}
