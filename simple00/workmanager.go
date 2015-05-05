// Copyright 2013 Ardan Studios. All rights reserved.
// Use of workManager source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package workmanager implements the WorkManager singleton. This manager
// controls the starting, shutdown and processing of work.
package main

import (
	"fmt"
	"sync/atomic"
	"runtime"
	"time"
)

const (
	timerPeriod time.Duration = 5 * time.Second // Interval to wake up on.
)

// workManager is responsible for starting and shutting down the program.
type workManager struct {
	Shutdown        int32
	ShutdownChannel chan string
}

var wm workManager // Reference to the singleton.

// Startup brings the manager to a running state.
func Startup() error {
	var err error
	defer CatchPanic(&err, "main", "workmanager.Startup")

	WriteStdout("main", "workmanager.Startup", "Started")

	// Create the work manager to get the program going
	wm = workManager{
		Shutdown:        0,
		ShutdownChannel: make(chan string),
	}

	// Start the work timer routine.
	// When workManager returns the program terminates.
	go wm.GoRoutineworkTimer()

	WriteStdout("main", "workmanager.Startup", "Completed")
	return err
}

// Shutdown brings down the manager gracefully.
func Shutdown() error {
	var err error
	defer CatchPanic(&err, "main", "workmanager.Shutdown")

	WriteStdout("main", "workmanager.Shutdown", "Started")

	// Shutdown the program
	WriteStdout("main", "workmanager.Shutdown", "Info : Shutting Down")
	atomic.CompareAndSwapInt32(&wm.Shutdown, 0, 1)

	WriteStdout("main", "workmanager.Shutdown", "Info : Shutting Down Work Timer")
	wm.ShutdownChannel <- "Down"
	<-wm.ShutdownChannel

	close(wm.ShutdownChannel)

	WriteStdout("main", "workmanager.Shutdown", "Completed")
	return err
}

// GoRoutineworkTimer perform the work on the defined interval.
func (workManager *workManager) GoRoutineworkTimer() {
	WriteStdout("wt", "grt", "Started")

	wait := timerPeriod

	for {
		WriteStdoutf("wt", "grt", "Info : Wait To Run : Seconds[%.0f]", wait.Seconds())

		select {
		case <-workManager.ShutdownChannel:
			WriteStdoutf("wt", "grt", "Shutting Down")
			workManager.ShutdownChannel <- "Down"
			return

		case <-time.After(wait):
			WriteStdoutf("wt", "grt", "Woke Up")
			break
		}

		// Perform the work
		workManager.PerformTheWork()
		wait = 3 * time.Second
	}
}

// PerformTheWork simulate some silly display work with silly sleep times.
func (workManager *workManager) PerformTheWork() {
	defer CatchPanic(nil, "workManager", "WorkManager.PerformTheWork")
	WriteStdout("wt", "wrt", "Started")

	// Perform work for 4 seconds
	for count := 0; count < 8; count++ {
		if atomic.CompareAndSwapInt32(&wm.Shutdown, 1, 1) == true {
			WriteStdout("wt", "grt", "Info : Request To Shutdown")
			return
		}

		WriteStdoutf("wt", "grt", "Counter = : %d", count)
		time.Sleep(time.Millisecond * 500)
	}

	WriteStdout("wt", "grt", "Completed")
}


// WriteStdout is used to write message directly stdout.
func WriteStdout(goRoutine string, functionName string, message string) {
	fmt.Printf("%s : %s : %s : %s\n", time.Now().Format("2006-01-02T15:04:05.000"), goRoutine, functionName, message)
}

// WriteStdoutf is used to write a formatted message directly stdout.
func WriteStdoutf(goRoutine string, functionName string, format string, a ...interface{}) {
	WriteStdout(goRoutine, functionName, fmt.Sprintf(format, a...))
}

// CatchPanic is used to catch and display panics.
func CatchPanic(err *error, goRoutine string, function string) {
	if r := recover(); r != nil {
		// Capture the stack trace
		buf := make([]byte, 10000)
		runtime.Stack(buf, false)

		WriteStdoutf(goRoutine, function, "PANIC Defered [%v] : Stack Trace : %v", r, string(buf))

		if err != nil {
			*err = fmt.Errorf("%v", r)
		}
	}
}
