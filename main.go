package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Zwlin98/moon/cluster"
	"github.com/Zwlin98/moon/service"
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

	signchan := make(chan os.Signal, 1)
	signal.Notify(signchan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	sign := <-signchan
	log.Println("receive sign program stop:", sign)
	time.Sleep(time.Second * 1)
}
