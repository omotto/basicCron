/*

Package basiccron implements a cron scheduler and job runner.

Installation

To download the specific tagged release, run:

	go get github.com/omotto/basicCron

Import it in your program as:

	import "github.com/omotto/basicCron"

Usage

	cron := basicCron.New(time.Minute)

	// Add new function to execute in two seconds, every hour
	if id, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (name string) { fmt.Println("Hello " + name) }, "Bob"); err != nil {
		t.Error(err)
	} else {
		fmt.Println(id)
	}

	cron.Start() // Start cron scheduler

	...

	cron.Stop() // Stop cron scheduler

*/
package basiccron

