package app

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type App struct {
	Config Config
}

func NewApp() *App {
	return &App{}
}

func (a *App) App() error {
	err := a.getFilePath()
	if err != nil {
		return err
	}

	err = a.getLastVersionCommit()
	if err != nil {
		return err
	}

	err = a.parseConfig()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) WatchSignals(cancel context.CancelFunc) {
	osSignalChan := make(chan os.Signal, 1)

	signal.Notify(osSignalChan, syscall.SIGINT)

	sig := <-osSignalChan

	l := log.WithField("method", "app.watchSignals")
	l.Infof("got signal: %q", sig.String())

	cancel()
}

func (a *App) parseConfig() error {
	var data []byte

	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}

	a.Config = *NewConfig()
	err = yaml.Unmarshal(data, &a.Config)
	if err != nil {
		return err
	}

	l := log.WithField("method", "app.parseConfig")
	l.Info("parse config end")
	return nil
}

func (a *App) getFilePath() error {
	l := log.WithField("method", "app.getFilePath")

	ex, err := os.Executable()
	if err != nil {
		return err
	}

	l.Infof("path: %s", ex)
	return nil
}

func (a *App) getLastVersionCommit() error {
	l := log.WithField("method", "app.getFilePath")

	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	stdout, err := cmd.Output()

	if err != nil {
		return err
	}

	result := string(stdout)[:len(string(stdout))-1]
	l.Infof("version: %s", result)
	return nil
}
