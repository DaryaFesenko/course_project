package main

import (
	"context"
	"course_project/app"
	"course_project/parsing"
	"errors"
	"flag"
	"fmt"
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	app := app.NewApp(getConfigPath())
	err := app.App()
	if err != nil {
		log.WithField("method", "app.App").Fatal(err)
	}

	timeOut := app.Config.GetTimeOut()
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	go exit(ctx, app)

	go app.WatchSignals(cancel)

	p := parsing.NewParser(app)
	request := "select first_name from table where id = 5 and name = 'dasha' or count <= 5;"
	sel, err := p.Parse(request)
	app.LogAccess(request)
	if err != nil {
		app.LogError(err)
	}
	fmt.Println(sel, err)
}

func exit(ctx context.Context, app *app.App) {
	<-ctx.Done()
	app.LogError(errors.New("context end"))
	os.Exit(int(syscall.SIGINT))
}

func getConfigPath() string {
	var configPath string

	flag.StringVar(&configPath, "c", "config.yaml", "Used for set path to config file.")
	flag.Parse()

	return configPath
}
