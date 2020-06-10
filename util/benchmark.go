package util

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	numKeys   = 1
	mil       = 100000 // Run the benchmark numKeys * mil times
	runs      = 10     // Number of time to run the benchmark
	wait      = 3000   // Time to wait before each step and before read / write
	firstWait = 10000  // Time to wait before starting benchmark
	step      = 100    // If read fails, wait this much before trying to avoid overloading the system
	maxTries  = 10     // Only retry an operation this many times
)

func Bench(test bool, logFile string, read func(string) bool, write func(string, string) bool) {
	defer WaitForCtrlC()
	if !test {
		return
	}

	f, err := os.Create(logFile)
	if err != nil {
		log.Fatal("unable to create csv log")
	}
	defer f.Close()

	time.Sleep(firstWait)
	log.Printf("Starting benchmark...\n")
	for i := 0; i < runs; i++ {
		log.Printf("BENCHMARK %v OF %v\n", i+1, runs)
		time.Sleep(wait)

		start := time.Now()
		k := 0
		success := 0
		for k < numKeys*mil {
			v := rand.Int()
			tries := 0
			for ok := false; !ok; ok = write(string(k), string(v)) {
				time.Sleep(step)
				tries++
				if tries > maxTries {
					success--
					break
				}
			}
			success++
			k += 1
		}
		_, _ = f.WriteString(fmt.Sprintf("write,%v,%v,%v,%v\n", i+1, success, numKeys*mil, time.Since(start).Microseconds()))

		time.Sleep(wait)
		start = time.Now()
		k = 0
		success = 0
		for k < numKeys*mil {
			tries := 0
			for ok := false; !ok; ok = read(string(k)) {
				time.Sleep(step)
				tries++
				if tries > maxTries {
					success--
					break
				}
			}
			success++
			k += 1
		}
		_, _ = f.WriteString(fmt.Sprintf("read,%v,%v,%v,%v\n", i+1, success, numKeys*mil, time.Since(start).Microseconds()))
	}
	log.Printf("BENCHMARK COMPLETE\n")
}
