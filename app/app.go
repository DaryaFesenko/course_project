package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type App struct {
	Config     Config
	configPath string
}

func NewApp(configPath string) *App {
	return &App{configPath: configPath}
}

func (a *App) App() error {
	err := a.getFilePath()
	if err != nil {
		return err
	}

	a.Config = *NewConfig()
	err = a.Config.ParseConfig(a.configPath)
	if err != nil {
		return err
	}

	err = a.getLastVersionCommit()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) WatchSignals(cancel context.CancelFunc) {
	osSignalChan := make(chan os.Signal, 1)

	signal.Notify(osSignalChan, syscall.SIGINT)

	<-osSignalChan

	a.LogAccess("user interrupted")
	cancel()
}

func (a *App) getFilePath() error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	log.WithField("method", "app.getFilePath").Infof("path: %s", ex)
	return nil
}

func (a *App) getLastVersionCommit() error {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	stdout, err := cmd.Output()

	if err != nil {
		return err
	}

	result := string(stdout)[:len(string(stdout))-1]
	log.WithField("method", "app.getFilePath").Infof("version: %s", result)
	return nil
}

func (a *App) LogAccess(request string) {
	filePath := a.Config.GetFilePathAccessLog()

	message := fmt.Sprintf("time: '%s', request: '%s'\n", time.Now().Format("02.01.2006 15:04:05"), request)
	if err := a.writeToFile(filePath, message); err != nil {
		a.LogError(err)
	}
}

func (a *App) LogError(err error) {
	filePath := a.Config.GetFilePathErrorLog()

	message := fmt.Sprintf("time: '%s', error: '%s'\n", time.Now().Format("02.01.2006 15:04:05"), err)
	if err := a.writeToFile(filePath, message); err != nil {
		log.Error(err)
	}
}

func (a *App) writeToFile(filePath, message string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		if _, err := file.WriteString(message); err != nil {
			return err
		}
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		data = append(data, []byte(message)...)

		err = os.WriteFile(filePath, data, fs.ModeAppend)
		if err != nil {
			return err
		}
	}

	return nil
}
