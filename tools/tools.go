package tools

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

type (
	Job struct {
		Id        int
		Name      string
		TimeInSec int
	}

	JobFn func() Job
)

func JobMaker(maxDurSecs int) JobFn {
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

func MakeJobs(count, maxDurSecs int) []Job {
	var jobs []Job
	for i := 0; i < count; i++ {
		tis := rand.Intn(maxDurSecs) + 1
		jobs = append(jobs, Job{
			Id:        i + 1,
			Name:      fmt.Sprintf("Job #id: %d, #sec: %d", i+1, tis),
			TimeInSec: tis,
		})
	}

	return jobs
}

func MemoryMonitor() {
	var stats runtime.MemStats
	stats.DebugGC = true

	for {
		time.Sleep(3 * time.Second)
		runtime.GC()
		runtime.ReadMemStats(&stats)
		fmt.Println("---------------------------------------------")
		fmt.Printf("Goroutines nr#: %d, HeapObjects: %d, StackInuse: %d KB, HeapAlloc: %d KB\r\n", runtime.NumGoroutine(), stats.HeapObjects, toKB(stats.StackInuse), toKB(stats.HeapAlloc))
		fmt.Println("---------------------------------------------")
	}
}

func toKB(input uint64) uint64 {
	return input / 1024
}
