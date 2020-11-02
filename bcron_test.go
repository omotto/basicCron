package basiccron

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCronError(t *testing.T) {

	cron := New(time.Second)

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func () {	fmt.Println("Hello, world") }, 10); err == nil {
		t.Error("This AddFunc should return Error, wrong number of args")
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, nil); err == nil {
		t.Error("This AddFunc should return Error, fn is nil")
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, 0); err == nil {
		t.Error("This AddFunc should return Error, fn is not func kind")
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (s string, n int) {	fmt.Printf("We have params here, string `%s` and nymber %d\n", s, n) }, "s", 10, 12); err == nil {
		t.Error("This AddFunc should return Error, wrong number of args")
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (s string, n int) {	fmt.Printf("We have params here, string `%s` and nymber %d\n", s, n) }, "s", "s2"); err == nil {
		t.Error("This AddFunc should return Error, args are not the correct type")
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (s string, n int) {	fmt.Printf("We have params here, string `%s` and nymber %d\n", s, n) }, "s", "s2"); err == nil {
		t.Error("This AddFunc should return Error, syntax error")
	}

	// custom types and interfaces as function params
	type user struct {
		ID   int
		Name string
	}
	var u user
	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (u user) { fmt.Println("Custom type as param") }, u); err != nil {
		t.Error(err)
	}

	type Foo interface {
		Bar() string
	}
	if _, err := cron.AddFunc(time.Now().Add(time.Second*2), time.Hour, func (i Foo) { i.Bar() }, u); err == nil {
		t.Error("This should return error, type that don't implements interface assigned as param")
	}
}

func TestCronBasic(t *testing.T) {
	testN := 0
	testS := ""

	cron := New(time.Second)

	if _, err := cron.AddFunc(time.Now().Add(time.Second*1), time.Second*3, func() { testN++ }); err != nil {
		t.Fatal(err)
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*1), time.Second*3, func(s string) { testS = s }, "param"); err != nil {
		t.Fatal(err)
	}

	cron.Start()

	time.Sleep(time.Second * 10)

	if testN != 3 {
		t.Error("func not executed correctly")
	}

	if testS != "param" {
		t.Error("func not executed or arg not passed")
	}

	cron.Stop()
}

func TestCronSchedule(t *testing.T) {
	testN := 0
	testS := ""

	cron := New(time.Second*2)

	var wg sync.WaitGroup
	wg.Add(2)

	if _, err := cron.AddFunc(time.Now().Add(time.Second*4), time.Second*10, func() { testN++; wg.Done() }); err != nil {
		t.Fatal(err)
	}

	if _, err := cron.AddFunc(time.Now().Add(time.Second*3), time.Second*10, func(s string) { testS = s; wg.Done() }, "param"); err != nil {
		t.Fatal(err)
	}

	cron.Start()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	if testN != 1 {
		t.Error("func 1 not executed as scheduled")
	}

	if testS != "param" {
		t.Error("func 2 not executed as scheduled")
	}
	cron.Stop()
}
