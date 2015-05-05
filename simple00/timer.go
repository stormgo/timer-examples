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

// timeManager is responsible for starting and shutting down the program.
type timeManager struct {
	Shutdown        int32
	ShutdownChannel chan string
}

var wm timeManager // Reference to the singleton.

// Startup brings the manager to a running state.
func Startup() error {
	var err error
	defer CatchPanic(&err, "main", "timeManager.Startup")

	WriteStdout("main", "timeManager.Startup", "Started")

	// Create the work manager to get the program going
	wm = timeManager{
		Shutdown:        0,
		ShutdownChannel: make(chan string),
	}

	// Start the work timer routine.
	// When timeManager returns the program terminates.
	go wm.GoRoutineworkTimer()

	WriteStdout("main", "timeManager.Startup", "Completed")
	return err
}

// Shutdown brings down the manager gracefully.
func Shutdown() error {
	var err error
	defer CatchPanic(&err, "main", "timeManager.Shutdown")

	WriteStdout("main", "timeManager.Shutdown", "Started")

	// Shutdown the program
	WriteStdout("main", "timeManager.Shutdown", "Info : Shutting Down")
	atomic.CompareAndSwapInt32(&wm.Shutdown, 0, 1)

	WriteStdout("main", "timeManager.Shutdown", "Info : Shutting Down Work Timer")
	wm.ShutdownChannel <- "Down"
	<-wm.ShutdownChannel

	close(wm.ShutdownChannel)

	WriteStdout("main", "timeManager.Shutdown", "Completed")
	return err
}

// GoRoutineworkTimer perform the work on the defined interval.
func (timeManager *timeManager) GoRoutineworkTimer() {
	WriteStdout("wt", "grt", "Started")

	wait := timerPeriod

	for {
		WriteStdoutf("wt", "grt", "Info : Wait To Run Top: Seconds[%.0f]", wait.Seconds())

		select {
		case <-timeManager.ShutdownChannel:
			WriteStdoutf("wt", "grt", "Shutting Down")
			timeManager.ShutdownChannel <- "Down"
			return

		case <-time.After(wait):
			WriteStdoutf("wt", "grt", "Woke Up")
			break
		}

		// Perform the work
		timeManager.PerformTheWork()
		wait = 3 * time.Second
	}
}

// PerformTheWork simulate some silly display work with silly sleep times.
func (timeManager *timeManager) PerformTheWork() {
	defer CatchPanic(nil, "timeManager", "timeManager.PerformTheWork")
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
