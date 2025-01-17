package workers

import "time"

// startWorker starts a background goroutine that repeatedly executes the given function at specified intervals.
//
// Parameters:
//   - f: A function (of type `func()`) that will be executed periodically.
//   - t: The time interval (of type `time.Duration`) between each execution of the function.
//
// Example usage:
//   workers.startWorker(func() { fmt.Println("Hello, world!") }, time.Minute)
//
// The function `f` will be called once every `t` duration, and this will continue indefinitely until the program exits.
//
// Note:
//   - The `f` function is run in a separate goroutine, allowing the main program to continue running independently.
//   - The `time.Duration` argument `t` can be specified in units such as `time.Second`, `time.Minute`, or `time.Hour`.
func startWorker(f func(), t time.Duration) {
	go func() {
		for {
			f()
			time.Sleep(t)
		}
	}()
}
