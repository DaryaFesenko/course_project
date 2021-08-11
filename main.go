package main

import (
	"context"
	"course_project/app"
	"course_project/parsing"
	"fmt"
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	app := app.NewApp()
	err := app.App()
	if err != nil {
		l := log.WithField("method", "app.App")
		l.Fatal(err)
	}

	timeOut := app.Config.GetTimeOut()
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	go exit(ctx)

	go app.WatchSignals(cancel)

	sel, err := parsing.Parse("select first_name from table where id = 5 and name = 'dasha' or count <= 5;")

	fmt.Println(sel, err)
}

func exit(ctx context.Context) {
	<-ctx.Done()
	os.Exit(int(syscall.SIGINT))
}
