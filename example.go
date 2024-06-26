package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Zwlin98/moon/cluster"
	"github.com/Zwlin98/moon/service"
	"golang.org/x/sys/unix"
)

func main() {
	// initialize services
	httpService := service.NewHttpService()
	pingService := service.NewPingService()

	// initialize cluster
	clusterd := cluster.GetClusterd()
	clusterd.Reload(cluster.DefaultConfig{
		"moon": "0.0.0.0:3345",
	})

	// register services
	clusterd.Register("http", httpService)
	clusterd.Register("ping", pingService)

	// start cluster
	clusterd.Open("moon")

	log.Printf("moon start")

	term := make(chan os.Signal, 1)

	signal.Notify(term, unix.SIGTERM)
	signal.Notify(term, os.Interrupt)

	<-term
}
