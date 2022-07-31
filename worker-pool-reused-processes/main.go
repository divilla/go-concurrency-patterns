package main

import (
	"fmt"
	"github.com/divilla/go-concurrency-patterns/tools"
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	// maximum random worker duration in seconds
	maxDurSecs = 3

	// added latency in between producing jobs
	jobsLatencyMS = 100

	// maximum number of concurrent workers
	workersCount = 8
)

func main() {
	makeJobFunc := tools.JobMaker(maxDurSecs)

	// turn on basic memory consumption monitoring, comment the line to turn it off
	go tools.MemoryMonitor()

	// channel used to exit worker goroutines
	stopCh := make(chan struct{})
	// channel used to pass Job(s) to workers
	inputCh := make(chan tools.Job)

	// Wait for interrupt signal to gracefully shutdown the server without a timeout implemented.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt)

	// sync.WaitGroup is used to keep main goroutine running until all worker goroutines finish
	workersWG := &sync.WaitGroup{}

	// total workers time required to finish jobs
	totalProcessingTime := 0

	// timer for measuring actual time required to finish work
	start := time.Now()

	// the total number of workers (concurrent goroutines used to do jobs) is fixed (workersCount)
	// jobs are passed to workers via 'inputCh' channel
	// workers are active during entire cycle of application run, all until application is terminated
	// and 'stopCh' is closed
	for i := 0; i < workersCount; i++ {
		go func(workerId int) {
			for {
				select {
				// this case is blocking until 'stopCh' channel is closed,
				// closed chanel is always unblocked, returning next buffered value or nil,
				// closed channel pattern is used to exit all worker goroutines,
				// preventing possible memory leak
				case <-stopCh:
					return
				// this case is blocking until job is written into inputCh,
				// first scheduled (one & only one) worker will read the job from channel and process it
				case job := <-inputCh:
					fmt.Println(fmt.Sprintf("Started worker nr#: %d, ", workerId), job.Name)

					// in this example runWorker is a simple pure function, not another goroutine
					runWorker(workersWG, job)
				}
			}
		}(i)
	}

MainLoop:
	for {
		select {

		// this case will be unblocked on 'Interrupt' os.Signal (ctrl+c), see line 34
		case <-quitCh:
			// closing channel, unblocks channel, returning buffered values one by one, then nil(s)
			// in this scenario, it's used to unblock line 60 & close all worker goroutines
			close(stopCh)
			// waiting is implemented to support 'graceful shutdown'
			// to prevent breaking goroutines code in the middle of execution
			workersWG.Wait()
			// this will break main loop
			break MainLoop

		// default case ensures code flow continues - line 77 is not blocking, removing it would wait for 'Interrupt'
		default:
		}

		// this line is here only to simulate stream incoming in particular velocity
		time.Sleep(time.Duration(jobsLatencyMS) * time.Millisecond)
		job := makeJobFunc()
		totalProcessingTime += job.TimeInSec

		// add a delta to *sync.WaitGroup to ensure main goroutine is running
		// until all workers (goroutines) finish processing
		workersWG.Add(1)

		// job starts its execution by writing into channel
		// with Go, it's the preferred way of communicating between goroutines
		inputCh <- job
	}

	fmt.Printf("Processor (workers) time #sec: %d, Elapsed actual time #sec: %v", totalProcessingTime, time.Since(start).Seconds())
}

func runWorker(workersWG *sync.WaitGroup, j tools.Job) {
	// defer fill ensure statements are executed at the end, no matter how 'runWorker' function exits
	// calling *sync.WaitGroup.Done(), decreases WaitGroups delta marking worker's goroutine job as Done
	defer workersWG.Done()

	if j.TimeInSec == 0 {
		fmt.Printf("invalid job id#: %d, error: time set to 0\r\n", j.Id)
		return
	}

	time.Sleep(time.Duration(j.TimeInSec) * time.Second)
	fmt.Println("Finished: ", j.Name)
}
