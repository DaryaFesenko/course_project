package main

import (
	"course_project/app"
	"flag"

	log "github.com/sirupsen/logrus"
)

func main() {
	app, err := app.New(getConfigPath())
	if err != nil {
		log.WithField("method", "app.App").Fatal(err)
	}

	app.Run()
}

func getConfigPath() string {
	var configPath string

	flag.StringVar(&configPath, "c", "config.yaml", "Used for set path to config file.")
	flag.Parse()

	return configPath
}
