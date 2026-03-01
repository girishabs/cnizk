package main

type ProveJob struct {
	Data              []int
	Constraints       []Constraint
	CreationTimestamp int32
	Message           []byte
	Signature         []byte
	PubKeyX           []byte
	PubKeyY           []byte
	ResultChan        chan ProveResult
}

var jobQueue chan ProveJob

func StartWorkers(n int) {

	jobQueue = make(chan ProveJob, 100)

	for i := 0; i < n; i++ {
		go worker()
	}
}

func worker() {

	for job := range jobQueue {

		res := RunProver(job)

		job.ResultChan <- res
	}
}

func SubmitProveJob(job ProveJob) ProveResult {

	job.ResultChan = make(chan ProveResult)

	jobQueue <- job

	return <-job.ResultChan
}
