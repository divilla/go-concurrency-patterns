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

	// added latency in between producing jobs
	jobsLatencyMS = 100

	// maximum number of concurrent workers
	workersCount = 8
)

type (
	Job struct {
		Id        int
		Name      string
		TimeInSec int
	}

	JobFn func() Job
)

func main() {
	makeJobFunc := jobMaker()

	stopCh := make(chan struct{})
	inputCh := make(chan Job)

	// Wait for interrupt signal to gracefully shutdown the server without a timeout implemented.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt)

	// sync.WaitGroup is used to keep main goroutine running until all worker goroutines finish
	workersWG := &sync.WaitGroup{}

	// total workers time required to finish jobs
	workersTotalTime := 0

	// timer for measuring actual time required to finish work
	start := time.Now()

	// the total number of workers (goroutines being assigned concurrent jobs) is fixed (workersCount)
	// jobs are passed via 'inputCh' channel
	// goroutines are active during entire cycle of application run, all until application is terminated
	// and 'stopCh' is closed
	for i := 0; i < workersCount; i++ {
		go func() {
			for {
				select {
				// closing stopCh enables this case, which sole purpose is to exit goroutine
				case <-stopCh:
					return
				// this line is blocking until job is written into inputCh,
				// with streams there is usually no fixed distribution of jobs,
				// so first available worker take the job and process it
				case job := <-inputCh:
					defer workersWG.Done()

					fmt.Println("Started", job.Name, fmt.Sprintf("Total running goroutines #nr: %d", runtime.NumGoroutine()))
					if job.TimeInSec == 0 {
						fmt.Printf("invalid job id#: %d, error: time set to 0\r\n", job.Id)
						continue
					}

					time.Sleep(time.Duration(job.TimeInSec) * time.Second)
					fmt.Println(fmt.Sprintf("Finished #id: %d", job.Id), fmt.Sprintf("Total running goroutines #nr: %d", runtime.NumGoroutine()))
				}
			}
		}()
	}

MainLoop:
	for {
		select {
		// code reaches this point on os.Signal - any signal,
		// for example ctrl+c will trigger this case
		case <-quitCh:
			// closing channel, unblocks channel,
			// returning buffered values (if any), then nil(s)
			// in this scenario, it's used to unblock line 58 & close that goroutine
			close(stopCh)
			// waiting is implemented to support 'graceful shutdown'
			// to prevent breaking goroutines code in the middle of execution
			workersWG.Wait()
			// this will break main loop
			break MainLoop
		default:
			// this line is here only to simulate stream incoming in particular speed
			time.Sleep(time.Duration(jobsLatencyMS) * time.Millisecond)
			job := makeJobFunc()
			workersWG.Add(1)
			workersTotalTime += job.TimeInSec
			// job starts its execution by writing into channel
			// with Go, it's the preferred way of communicating between goroutines
			inputCh <- job
		}
	}

	fmt.Printf("Workers time #sec: %d, Elapsed time #sec: %v", workersTotalTime, time.Since(start).Seconds())
}

func jobMaker() JobFn {
	var i int
	rndSrc := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(rndSrc)

	return func() Job {
		i++
		tis := rnd.Intn(maxDurSecs + 1)

		return Job{
			Id:        i,
			Name:      fmt.Sprintf("Job #id: %d, #sec: %d", i, tis),
			TimeInSec: tis,
		}
	}
}
