package util

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

type TestParams struct {
	NumKeys   int
	Mil       int
	Runs      int
	Wait      time.Duration
	FirstWait time.Duration
	Step      time.Duration
	MaxTries  int
	Enabled   bool
	LogFile   string
}

func Bench(testParams TestParams, read func(string) bool, write func(string, string) bool) {
	defer WaitForCtrlC()
	if !testParams.Enabled {
		return
	}

	f, err := os.Create(testParams.LogFile)
	if err != nil {
		log.Fatal("unable to create csv log")
	}
	defer f.Close()

	time.Sleep(testParams.FirstWait)
	log.Printf("Starting benchmark...\n")
	for i := 0; i < testParams.Runs; i++ {
		log.Printf("BENCHMARK %v OF %v\n", i+1, testParams.Runs)
		time.Sleep(testParams.Wait)

		start := time.Now()
		failure := 0
		for k := 0; k < testParams.NumKeys*testParams.Mil; k++ {
			v := rand.Int()
			tries := 0
			for ok := false; !ok; ok = write(fmt.Sprintf("%d", k), fmt.Sprintf("%d", v)) {
				time.Sleep(testParams.Step)
				tries++
				if tries > testParams.MaxTries {
					break
				}
			}
			failure += tries
		}
		_, _ = f.WriteString(fmt.Sprintf("write,%v,%v,%v,%v\n", i+1, failure, testParams.NumKeys*testParams.Mil, time.Since(start).Microseconds()))

		time.Sleep(testParams.Wait)
		start = time.Now()
		failure = 0
		for k := 0; k < testParams.NumKeys*testParams.Mil; k++ {
			tries := 0
			for ok := false; !ok; ok = read(fmt.Sprintf("%d", k)) {
				time.Sleep(testParams.Step)
				tries++
				if tries > testParams.MaxTries {
					break
				}
			}
			failure += tries
		}
		_, _ = f.WriteString(fmt.Sprintf("read,%v,%v,%v,%v\n", i+1, failure, testParams.NumKeys*testParams.Mil, time.Since(start).Microseconds()))
	}
	log.Printf("BENCHMARK COMPLETE\n")
}
