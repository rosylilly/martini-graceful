package graceful

import (
	"bytes"
	"github.com/go-martini/martini"
	"log"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func martiniApp(gs *Shutdown) *martini.ClassicMartini {
	buff := bytes.NewBufferString("")

	m := martini.Classic()
	m.Map(log.New(buff, "[martini] ", 0))
	m.Use(gs.Handler)

	m.Get("/", func() string {
		time.Sleep(5 * time.Second)
		return "Hello, world\n"
	})

	return m
}

func Test_GracefulShutdown(t *testing.T) {
	gs := New(10*time.Second, syscall.SIGINT)

	os.Setenv("PORT", "3000")
	m := martiniApp(gs)

	graceChan := make(chan error)
	go func() {
		graceChan <- gs.RunGracefully(m)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		res, err := http.Get("http://localhost:3000/")
		if err != nil {
			t.Error(err)
		}
		res.Body.Close()
		wg.Done()
	}()

	go func() {
		err := <-graceChan
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	wg.Wait()
}

func Test_GracefulShutDownWithTimeout(t *testing.T) {
	gs := New(1*time.Millisecond, syscall.SIGINT)

	os.Setenv("PORT", "3001")
	m := martiniApp(gs)

	graceChan := make(chan error)
	go func() {
		graceChan <- gs.RunGracefully(m)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		res, err := http.Get("http://localhost:3001/")
		if err != nil {
			t.Error(err)
		}
		res.Body.Close()
		wg.Done()
	}()

	go func() {
		err := <-graceChan
		if err == nil {
			t.Error("Not work time out")
		} else {
			t.Log(err)
		}
		wg.Done()
	}()

	wg.Wait()
}
