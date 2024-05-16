package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Zwlin98/moon/cluster"
	"github.com/Zwlin98/moon/lua"
	"github.com/Zwlin98/moon/service"
	"golang.org/x/sys/unix"
)

func main() {
	// initialize services
	httpService := service.NewHttpService()
	exampleService := service.NewExampleService()

	// initialize cluster
	clusterd := cluster.GetClusterd()
	clusterd.Reload(cluster.DefaultConfig{
		"moon": "127.0.0.1:3345",
		"db":   "127.0.0.1:2528",
	})

	// register services
	clusterd.Register("http", httpService)
	clusterd.Register("example", exampleService)

	// start cluster
	clusterd.Open("moon")

	// call Skynet
	callOnce := func(idx int) {
		if idx%2 == 1 {
			_, err := cluster.Call("db", "sdb1", "GET", []lua.Value{lua.String("ping")})
			if err != nil {
				log.Println("call db error:", err)
			}
		} else {
			_, err := cluster.Call("db", "sdb2", "GET", []lua.Value{lua.String("ping")})
			if err != nil {
				log.Println("call db error:", err)
			}
		}
	}

	cnt := 50000
	start := time.Now()
	var wg sync.WaitGroup
	for i := range cnt {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			callOnce(x)
		}(i)
	}
	wg.Wait()

	log.Printf("call db %d times cost: %s\n", cnt, time.Since(start))

	term := make(chan os.Signal, 1)

	signal.Notify(term, unix.SIGTERM)
	signal.Notify(term, os.Interrupt)

	<-term
}
