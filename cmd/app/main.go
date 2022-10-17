package main

import (
	"log"
	"test_task/internal/options"
	"test_task/internal/server"
)

func main() {
	port, adsIpPort := options.BuildOptions()
	srv := server.NewServer(port, adsIpPort)
	if err := srv.Start(); err != nil {
		log.Fatalln(err)
	}
}
