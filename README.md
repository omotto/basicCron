[![GoDoc](http://godoc.org/github.com/omotto/basicCron?status.png)](http://godoc.org/github.com/omotto/basicCron)
[![Build Status](https://travis-ci.com/omotto/basicCron.svg?branch=main)](https://travis-ci.com/omotto/basicCron)
[![Coverage Status](https://coveralls.io/repos/github/omotto/basicCron/badge.svg)](https://coveralls.io/github/omotto/basicCron)

# basicCron

Package basicCron implements a cron scheduler and job runner.

### Installation

To download the specific tagged release, run:

```
go get github.com/omotto/basicCron
```

Import it in your program as:

```
import "github.com/omotto/basicCron"
```

### Usage

```
cron := basicCron.New(time.Minute)

// Add new function to execute in two seconds, every hour
if id, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (name string) { fmt.Println("Hello " + name) }, "Bob"); err != nil {
    t.Error(err)
}

cron.Start() // Start cron scheduler

...

cron.DelFunc(id) // Remove job function

...

cron.Stop() // Stop cron scheduler
```