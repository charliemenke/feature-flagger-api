package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/charliemenke/feature-flagger-api/api"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	ExitHandler()
	godotenv.Load(".env")
	var port = os.Getenv("SERVER_PORT")
	dbID, err := strconv.Atoi(os.Getenv("REDIS_DB_ID"))
	if err != nil {
		log.Errorf("Error converting redis ID to int")
	}
	server := api.FeatureFlaggerAPI{}
	server.Initialize(os.Getenv("REDIS_DB_HOST"), os.Getenv("REDIS_DB_PORT"), dbID, os.Getenv("REDIS_DB_PASSWORD"))
	server.Start(port)
}

func ExitHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("The server is stopping")
		os.Exit(0)
	}()
}
