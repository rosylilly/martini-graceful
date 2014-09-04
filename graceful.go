// Package graceful provides graceful shutdown to the martini application.
package graceful

import (
	"fmt"
	"github.com/go-martini/martini"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Runnable interface {
	Run()
}

type Shutdown struct {
	timeoutDuration time.Duration
	wg              sync.WaitGroup
	signals         []os.Signal
}

func New(timeoutDuration time.Duration, signals ...os.Signal) *Shutdown {
	return &Shutdown{
		timeoutDuration: timeoutDuration,
		signals:         signals,
	}
}

func (s *Shutdown) Handler(c martini.Context) {
	s.wg.Add(1)
	c.Next()
	s.wg.Done()
}

func (s *Shutdown) WaitForRequests() <-chan error {
	c := make(chan error, 1)

	wgChan := make(chan struct{})
	go func() {
		s.wg.Wait()
		wgChan <- struct{}{}
	}()

	go func() {
		select {
		case <-time.After(s.timeoutDuration):
			c <- fmt.Errorf("Graceful shutdown timed out")
			return
		case <-wgChan:
			c <- nil
			return
		}
	}()

	return c
}

func (s *Shutdown) WaitForSignals() chan os.Signal {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, s.signals...)

	return sigChan
}

func (s *Shutdown) Wait() error {
	sigChan := s.WaitForSignals()
	<-sigChan

	return <-s.WaitForRequests()
}

func (s *Shutdown) RunGracefully(m Runnable) error {
	go func() {
		m.Run()
	}()

	return s.Wait()
}
