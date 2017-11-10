package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type SafeMap struct {
	M map[string]bool
	L sync.Mutex
}

var seen = &SafeMap{M: make(map[string]bool)}
var processed = make(chan string, 1)
var closed = make(chan struct{}, 10)

func (m *SafeMap) Add(k string) {
	m.L.Lock()
	m.M[k] = true
	m.L.Unlock()
}

func (m *SafeMap) Get(k string) bool {
	m.L.Lock()
	defer m.L.Unlock()
	return m.M[k]
}

func fetch(url string) {
	seen.Add(url)
	processed <- url
}

func closeChans() {
	for i := 0; i < 10; i++ {
		<-closed
	}
}

func signalHandler() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// foreach signal received
	for signal := range c {
		fmt.Println("Signal received: ", signal.String())

		switch signal {
		case syscall.SIGINT, syscall.SIGTERM:
			closeChans()
			os.Exit(0)
		case syscall.SIGHUP:
			fmt.Println("SIGHUP received")
		}

	}
}

func main() {
	urlchan := make(chan string)
	timeout := make(chan struct{}, 1)

	go signalHandler()

	go func() {
		time.Sleep(30 * time.Second)
		timeout <- struct{}{}
	}()

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { closed <- struct{}{} }()
			for url := range urlchan {
				if seen.Get(url) {
					continue
				}
				fetch(url)
			}
		}()
	}

	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"
	urlchan <- "http://google.com"

	close(urlchan)

	// wait for goroutines to settle
	for {
		select {
		case url := <-processed:
			// a read from ch has occurred
			fmt.Println("Processed ", url)
			fmt.Println("Url has been seen: ", seen.Get(url))
		case <-timeout:
			// the read from ch has timed out
			fmt.Println("Timeout")
			closeChans()
			return
		}
	}
}
