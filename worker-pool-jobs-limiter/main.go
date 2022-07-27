package main

import (
	"fmt"
	"github.com/divilla/go-concurrency-patterns/tools"
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
	jobsCount = 1024

	// maximum number of concurrent workers
	workersCount = 8
)

func main() {
	jobs := tools.MakeJobs(jobsCount, maxDurSecs)

	// turn on basic memory consumption monitoring, comment the line to turn it off
	go tools.MemoryMonitor()

	// sync.WaitGroup is used to keep main goroutine running until all worker goroutines finish
	workersWG := &sync.WaitGroup{}

	// buffered channel is used to limit total number of concurrent workers running
	workersCh := make(chan struct{}, workersCount)

	// Wait for interrupt signal to gracefully shut down the server without a timeout implemented.
	// Buffered channel is used to avoid missing signals as recommended for signal.Notify
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt)

	// total processor time needed to run all the workers & finish jobs
	totalProcessingTime := 0

	// timer for measuring actual (Earth) elapsed time required to finish work
	start := time.Now()

MainLoop:
	for key, job := range jobs {
		select {
		case <-quitCh:
			break MainLoop

		// try to fill the buffer with single struct{}
		// this case is blocking (not executing) when workersCh buffer is full
		// 'workersCh' is full when filled with 'workersCount' number of struct{}
		case workersCh <- struct{}{}:
			// add a delta to *sync.WaitGroup to ensure main goroutine is running
			// until all workers (goroutines) finish processing
			workersWG.Add(1)

			// add time required for job to total required processing time
			totalProcessingTime += job.TimeInSec

			// start 'runWorker' as concurrent goroutine
			go runWorker(workersCh, workersWG, key+1, job)
		}
	}

	// this line is blocking until *sync.WaitGroup's delta becomes 0,
	// in other words, all the worker goroutines finished and executed 'Done()'
	workersWG.Wait()

	fmt.Printf("Processor (workers) time #sec: %d, Elapsed actual time #sec: %v", totalProcessingTime, time.Since(start).Seconds())
}

func runWorker(workersCh <-chan struct{}, workersWG *sync.WaitGroup, k int, j tools.Job) {
	// defer fill ensure statements are executed at the end, no matter how 'runWorker' function exits
	defer func() {
		// calling *sync.WaitGroup.Done(), decreases WaitGroups delta marking worker's goroutine job as Done
		workersWG.Done()

		// empty single struct{} from 'workersCh' channel buffer
		// that will unblock app main loop & start new worker goroutine
		<-workersCh
	}()

	if j.TimeInSec == 0 {
		fmt.Printf("invalid job id#: %d, error: time set to 0\r\n", j.Id)
		return
	}

	time.Sleep(time.Duration(j.TimeInSec) * time.Second)
	fmt.Println(fmt.Sprintf("Finished #key: %d", k), j.Name, fmt.Sprintf("Total running goroutines #nr: %d", runtime.NumGoroutine()))
}
