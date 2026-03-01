package main

import "fmt"

type ProveJob struct {
	Data              []int
	Constraints       []Constraint
	CreationTimestamp int32
	Message           []byte
	Randomness        []byte
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

		func() {
			defer func() {
				if r := recover(); r != nil {
					job.ResultChan <- ProveResult{
						Err: fmt.Errorf("worker panic: %v", r),
					}
				}
			}()

			_, circuitDir := CreateTmpDir()

			res := RunProver(circuitDir, job)

			job.ResultChan <- res
		}()
	}
}

func SubmitProveJob(job ProveJob) ProveResult {

	job.ResultChan = make(chan ProveResult, 1)

	jobQueue <- job

	return <-job.ResultChan
}
