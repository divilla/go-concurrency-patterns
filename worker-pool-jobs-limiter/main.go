package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

const (
	// maximum random worker duration in seconds
	maxDurSecs = 3

	// total number of jobs
	jobsCount = 256

	// maximum number of concurrent workers
	workersCount = 8
)

type (
	Job struct {
		Id        int
		Name      string
		TimeInSec int
	}
)

func main() {
	jobs := makeJobs()

	// sync.WaitGroup is used to keep main goroutine running until all worker goroutines finish
	workersWG := &sync.WaitGroup{}

	// buffered channel is used to limit total number of concurrent workers running
	workersCh := make(chan struct{}, workersCount)

	// Wait for interrupt signal to gracefully shutdown the server without a timeout implemented.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt)

	// total workers time required to finish jobs
	workersTotalTime := 0

	// timer for measuring actual time required to finish work
	start := time.Now()

MainLoop:
	for key, job := range jobs {
		select {
		case <-quitCh:
			break MainLoop
		default:
		}

		// register single job Add(1) in sync.WaitGroup
		// this line will block main goroutine at line 77, preventing main program from exiting
		// for as long as nr# of Add() > Done()
		workersWG.Add(1)

		// fill the buffer with single struct{}
		// this line is blocking further execution when buffer is full:
		// 'workersCh' is filled with 'workersCount' number of struct{}
		workersCh <- struct{}{}
		workersTotalTime += job.TimeInSec

		// this pattern creating new goroutine for each job, but goroutines are cheap,
		// so creating & dumping single goroutine for each job has marginal impact on overall performance,
		// while making the code way more readable
		go func(k int, j Job) {
			defer func() {
				// release single struct{} & unblock line 54
				<-workersCh

				// mark job in sync.WaitGroup Done()
				// when all jobs are Done() - total nr# Add(x) == total nr# Done() main goroutine will unblock line 77
				workersWG.Done()
			}()

			fmt.Println(fmt.Sprintf("Started #key: %d", k), j.Name, fmt.Sprintf("Total running goroutines #nr: %d", runtime.NumGoroutine()))
			time.Sleep(time.Duration(j.TimeInSec) * time.Second)
			fmt.Println(fmt.Sprintf("Finished #key: %d", k), j.Name, fmt.Sprintf("Total running goroutines #nr: %d", runtime.NumGoroutine()))
		}(key+1, job)
	}

	// this line is blocking until all the workers executed 'workersWG.Done()'
	workersWG.Wait()

	fmt.Printf("Workers time #sec: %d, Elapsed time #sec: %v", workersTotalTime, time.Since(start).Seconds())
}

func makeJobs() []Job {
	var jobs []Job
	for i := 0; i < jobsCount; i++ {
		tis := rand.Intn(maxDurSecs) + 1
		jobs = append(jobs, Job{
			Id:        i + 1,
			Name:      fmt.Sprintf("Job #id: %d, #sec: %d", i+1, tis),
			TimeInSec: tis,
		})
	}

	return jobs
}
