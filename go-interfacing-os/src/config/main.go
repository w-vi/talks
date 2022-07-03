package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Context is a struct to wrap the config file and guard it with a mutex
type context struct {
	config string
	sync.Mutex
}

// The global context variable
var ctx *context

func (c *context) Config() string {
	c.Lock()
	cnf := c.config
	c.Unlock()
	return cnf
}

// A function that reads the file "config.yml" and returns a pointer to an instance of the config struct.
func readConfig() string {
	content, err := ioutil.ReadFile("config.conf")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return string(content)
}

// That's the interesting part
func swapConfig(sigusr chan os.Signal) {
	for {
		// This line will block until the signal is received
		<-sigusr

		// Whenever the signal is received. Read the configuration file and swap the old config out.
		ctx.Lock()
		ctx.config = readConfig()
		ctx.Unlock()
	}
}

func main() {
	fmt.Printf("Process PID : %v\n", os.Getpid())

	ctx = &context{config: readConfig()}

	sigusr := make(chan os.Signal, 1)
	// Send to channel c whenver you receive a SIGUSR1 signal.
	signal.Notify(sigusr, syscall.SIGUSR1)
	go swapConfig(sigusr)

	for {
		fmt.Println(ctx.Config())
		time.Sleep(time.Second)
	}
}
