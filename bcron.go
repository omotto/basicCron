package basiccron

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

type job struct {
	uuid			string
	triggerDate		time.Time
	period			time.Duration
	function 		interface{}
	fparams			[]interface{}
	freturns		[]reflect.Value
}

// Cron Struct
type Cron struct {
	lastIndex		int64				// Last inserted job index
	running			bool				// Scheduler status running (true) or stopped
	stop			chan bool			// Stop channel
	period			time.Duration		// Period of time between scheduler analysis
	jobs 			[]job				// Slice of jobs to be executed
	addJob			chan job			// Add Job Channel
	delJob			chan string			// Del Job Channel
}

// New Creates a new Cron instance for given:
// 		(period time.Duration) : period of time between time check
// returns:
//		(cron *Cron) : new Cron object
func New(period time.Duration) *Cron {
	return &Cron{
		running: false,
		period : period,
		stop: make(chan bool),
		addJob: make(chan job),
		delJob: make(chan string),
		lastIndex: 0,
	}
}

// AddFunc adds a func to Cron to be executed given:
//		(triggerDate time.Time) : Job's execution start time
//		(period time.Duration) : Period between job execution
//		(function interface{}) : function to execute
//		(fparams ...interface{}) : input argument list of the attached function
// returns:
//		(uuid string) : UUID to identify this new created Job
//		(err error) : error value (nil if no error)
func (c *Cron) AddFunc(triggerDate time.Time, period time.Duration, function interface{}, fparams ...interface{}) (uuid string, err error) {
	// Check input parameters
	// -- Check first trigger date is in the future
	if triggerDate.Before(time.Now()) {	return uuid, fmt.Errorf("triggerDate must be in the future") }
	// -- Check if param function contains a function
	if function == nil || reflect.ValueOf(function).Kind() != reflect.Func { return uuid, fmt.Errorf("invalid function parameter") }
	// -- Check number of params
	if len(fparams) != reflect.TypeOf(function).NumIn() { return uuid, fmt.Errorf("number of function params and number of provided params doesn't match") }
	// -- Check input params type belongs to function params type
	for i := 0; i < reflect.TypeOf(function).NumIn(); i++ {
		functionParam := reflect.TypeOf(function).In(i)
		inputParam := reflect.TypeOf(fparams[i])
		if functionParam != inputParam {
			if functionParam.Kind() != reflect.Interface { return uuid, fmt.Errorf(fmt.Sprintf("param[%d] must be be `%s` not `%s`", i, functionParam, inputParam)) }
			if !inputParam.Implements(functionParam) { return uuid, fmt.Errorf(fmt.Sprintf("param[%d] of type `%s` doesn't implement interface `%s`", i, functionParam, inputParam)) }
		}
	}
	// Add new job
	uuid = strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(c.lastIndex, 16)
	c.lastIndex++
	newJob := job { uuid: uuid, triggerDate: triggerDate, period: period, function: function, fparams: fparams }
	if c.running == true {
		c.addJob <- newJob
	} else {
		c.jobs = append(c.jobs, newJob)
	}
	return uuid, err
}


// DelFunc removes a func from Cron for given UUID to identify the job to be removed
func (c *Cron) DelFunc(uuid string) {
	if c.running == true {
		c.delJob <- uuid
	} else {
		c.delFunc(uuid)
	}
}

// Stop stops Cron scheduler
func (c *Cron) Stop() {
	if c.running == true {
		c.stop <- true
		c.running = false
	}
}

// Start starts Cron scheduler
func (c *Cron) Start() {
	if c.running == false {
		c.running = true
		go func() {
			ticker := time.NewTicker(c.period)
			for {
				select {
				case <-ticker.C:
					for i, j := range c.jobs {
						now := time.Now()
						if j.triggerDate.Before(now) || j.triggerDate.Equal(now) {
							c.jobs[i].triggerDate = j.triggerDate.Add(j.period)
							go c.execJob(j)
						}
					}
				case newJob := <-c.addJob:
					ticker.Stop()
					c.jobs = append(c.jobs, newJob)
					ticker = time.NewTicker(c.period)
				case uuid := <-c.delJob:
					ticker.Stop()
					c.delFunc(uuid)
					ticker = time.NewTicker(c.period)
				case <-c.stop:
					return
				}
			}
		}()
	}
}

// Private Methods

func (c *Cron) execJob(j job) {
	defer func() {
		if r := recover(); r != nil { log.Println("cron exec func ", j.uuid, " error ", r) }
	}()
	args := make([]reflect.Value, len(j.fparams))
	for i, param := range j.fparams {
		args[i] = reflect.ValueOf(param)
	}
	j.freturns = reflect.ValueOf(j.function).Call(args)
}

func (c *Cron) delFunc(uuid string) {
	var index = -1
	for i := 0; i < len(c.jobs); i++ {
		if c.jobs[i].uuid == uuid {
			index = i
			break
		}
	}
	if index != -1 {
		c.jobs[index] = c.jobs[len(c.jobs)-1]
		c.jobs = c.jobs[:len(c.jobs)-1]
	}
}
