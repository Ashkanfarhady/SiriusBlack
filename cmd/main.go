package main

import (
	"os"
	"strconv"

	"github.com/Ashkanfarhady/SiriusBlack/pkg/handlers"
)

func main() {
	Port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
	BindAddress := os.Getenv("BIND_ADDRESS")
	RedisPort, err := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		panic(err)
	}
	RedisHost := os.Getenv("REDIS_HOST")

	server := handlers.Server{
		BindAddress: BindAddress,
		Port:        Port,
		RedisHost:   RedisHost,
		RedisPort:   RedisPort,
	}
	server.Serve()
}
