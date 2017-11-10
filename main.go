package main

import (
	"fmt"
	"sync"
	"time"
)

type SafeMap struct {
	M map[string]bool
	L sync.Mutex
}

var seen = &SafeMap{M: make(map[string]bool)}
var processed = make(chan string, 1)

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

func main() {
	urlchan := make(chan string)
	closed := make(chan struct{}, 10)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(30 * time.Second)
		timeout <- true
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
	for i := 0; i < 10; i++ {
		select {
		case url := <-processed:
			// a read from ch has occurred
			fmt.Println("Processed ", url)
			fmt.Println("Url has been seen: ", seen.Get(url))
		case <-timeout:
			// the read from ch has timed out
			fmt.Println("Timeout")
			for i := 0; i < 10; i++ {
				<-closed
			}
		}
	}
}
