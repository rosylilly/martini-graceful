package main

import (
	".."
	"github.com/go-martini/martini"
	"log"
	"syscall"
	"time"
)

func main() {
	m := martini.Classic()

	gs := graceful.New(10*time.Second, syscall.SIGTERM, syscall.SIGINT)
	m.Use(gs.Handler)

	m.Get("/", func() string {
		time.Sleep(5 * time.Second)
		return "hello world\n"
	})

	err := gs.RunGracefully(m)
	if err != nil {
		log.Println(err)
	}
}
