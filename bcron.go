package basicCron

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

type Job struct {
	uuid			string
	triggerDate		time.Time
	period			time.Duration
	function 		interface{}
	fparams			[]interface{}
}

type Cron struct {
	lastIndex		int64
	//mu 				*sync.Mutex
	running			bool
	stop			chan bool
	period			time.Duration
	jobs 			[]Job
	addJob			chan Job
	delJob			chan string
}

// New returns a new Cron object

func New(period time.Duration) *Cron {
	return &Cron{
	//	mu: &sync.Mutex{},
		running: false,
		period : period,
		stop: make(chan bool),
		addJob: make(chan Job),
		delJob: make(chan string),
		lastIndex: 0,
	}
}

// AddFunc adds a func to the Cron to be executed

func (c *Cron) AddFunc(triggerDate time.Time, period time.Duration, function interface{}, fparams ...interface{}) (uuid string, err error) {
	// Check input parameters
	// -- Check first trigger date is in the future
	if triggerDate.Before(time.Now()) {	return uuid, errors.New("triggerDate must be in the future") }
	// -- Check if param function contains a function
	if function == nil || reflect.ValueOf(function).Kind() != reflect.Func { return uuid, errors.New("invalid function parameter") }
	// -- Check number of params
	if len(fparams) != reflect.TypeOf(function).NumIn() { return uuid, errors.New("number of function params and number of provided params doesn't match") }
	// -- Check input params type belongs to function params type
	for i := 0; i < reflect.TypeOf(function).NumIn(); i++ {
		functionParam := reflect.TypeOf(function).In(i)
		inputParam := reflect.TypeOf(fparams[i])
		if functionParam != inputParam {
			if functionParam.Kind() != reflect.Interface { return uuid, errors.New(fmt.Sprintf("param[%d] must be be `%s` not `%s`", i, functionParam, inputParam)) }
			if !inputParam.Implements(functionParam) { return uuid, errors.New(fmt.Sprintf("param[%d] of type `%s` doesn't implement interface `%s`", i, functionParam, inputParam)) }
		}
	}
	// Add new job
	uuid = strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(c.lastIndex, 16)
	c.lastIndex++
	job := Job { uuid: uuid, triggerDate: triggerDate, period: period, function: function, fparams: fparams }
	if c.running == true {
		c.addJob <- job
	} else {
		c.jobs = append(c.jobs, job)
	}
	return uuid, err
}


func (c *Cron) DelFunc(uuid string) {
	if c.running == true {
		c.delJob <- uuid
	} else {
		c.delFunc(uuid)
	}
}

func (c *Cron) delFunc(uuid string) {
	var index int = -1
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

// Stops Cron scheduler

func (c *Cron) Stop() {
	if c.running == true {
		c.stop <- true
		c.running = false
	}
}

// Starts Cron scheduler

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
				case job := <-c.addJob:
					ticker.Stop()
					c.jobs = append(c.jobs, job)
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
