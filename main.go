package main

import (
	"log"
	"moon/cluster"
	"moon/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	httpService := service.NewHttpService()

	clusterd := cluster.GetClusterd()
	clusterd.Reload(cluster.DefaultConfig{
		"moon": "127.0.0.1:3345",
		"db":   "127.0.0.1:2528",
	})

	clusterd.Register("http", httpService)

	clusterd.Open("moon")

	signchan := make(chan os.Signal, 1)
	signal.Notify(signchan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	sign := <-signchan
	log.Println("receive sign program stop:", sign)
	time.Sleep(time.Second * 1)
}
