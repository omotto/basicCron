package basicCron

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
)

type Job struct {
	triggerDate		time.Time
	period			time.Duration
	function 		interface{}
	fparams			[]interface{}
}

type Cron struct {
//	mu 				*sync.Mutex
	running			bool
	stop			chan bool
	period			time.Duration
	jobs 			[]Job
}

// New returns a new Cron object

func New(period time.Duration) *Cron {
	return &Cron{running: false, period : period, stop: make(chan bool) }
}

// AddFunc adds a func to the Cron to be executed

func (c *Cron) AddFunc(triggerDate time.Time, period time.Duration, function interface{}, fparams ...interface{}) (err error) {
	// Check input parameters
	// -- Check first trigger date is in the future
	if triggerDate.Before(time.Now()) {	return errors.New("triggerDate must be in the future") }
	// -- Check if param function contains a function
	if function == nil || reflect.ValueOf(function).Kind() != reflect.Func { return errors.New("invalid function parameter") }
	// -- Check number of params
	if len(fparams) != reflect.TypeOf(function).NumIn() { return errors.New("number of function params and number of provided params doesn't match") }
	// -- Check input params type belongs to function params type
	for i := 0; i < reflect.TypeOf(function).NumIn(); i++ {
		functionParam := reflect.TypeOf(function).In(i)
		inputParam := reflect.TypeOf(fparams[i])
		if functionParam != inputParam {
			if functionParam.Kind() != reflect.Interface { return errors.New(fmt.Sprintf("param[%d] must be be `%s` not `%s`", i, functionParam, inputParam)) }
			if !inputParam.Implements(functionParam) { return errors.New(fmt.Sprintf("param[%d] of type `%s` doesn't implement interface `%s`", i, functionParam, inputParam)) }
		}
	}
	// Add new job
//	c.mu.Lock()
		c.jobs = append(c.jobs, Job { triggerDate: triggerDate, period: period, function: function, fparams: fparams })
//	c.mu.Unlock()
	return err
}

// Stops Cron scheduler

func (c *Cron) Stop() {
	if c.running == true {
		c.stop <- true
		c.running = false
	}
}

// Starts Cron scheduler

func (c *Cron) Start() {
	c.running = true
	go func() {
		for {
			select {
			case <-time.After(c.period):
//				c.mu.Lock()
				for i, j := range c.jobs {
					now := time.Now()
					if j.triggerDate.Before(now) || j.triggerDate.Equal(now) {
						c.jobs[i].triggerDate =	j.triggerDate.Add(j.period)
						go c.execJob(j)
					}
				}
//				c.mu.Unlock()
			case <-c.stop:
				return
			}
		}
	}()
}

// Private Methods

func (c *Cron) execJob(j Job) {
	defer func() {
		if r := recover(); r != nil { log.Println("crontab error", r) }
	}()
	args := make([]reflect.Value, len(j.fparams))
	for i, param := range j.fparams {
		args[i] = reflect.ValueOf(param)
	}
	reflect.ValueOf(j.function).Call(args)
}
