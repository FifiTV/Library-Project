package workers

import (
	"my-firebase-project/controllers"
	"time"
)

// SetWorkers initializes background workers that will periodically execute tasks.
//
// The background worker is started by calling the startWorker function, passing a function
// to be executed periodically (in this case, controllers.StartWorker) and the interval (time.Hour).
//
// If you want to run some function in time interval just add code here.
//
// Example:
//
//	SetWorkers() will execute controllers.StartWorker every hour in the background.
func SetWorkers() {
	startWorker(func() {
		controllers.StartWorker()
	}, time.Hour)
}
