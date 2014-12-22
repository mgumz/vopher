package main

type JobUi interface {
	Start()
	Stop()

	Refresh()

	AddJob(job_id string)
	JobDone(job_id string)
	Wait() // wait for all jobs to be .Done()
}
