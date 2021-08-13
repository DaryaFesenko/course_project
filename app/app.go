package app

import (
	"bufio"
	"context"
	"course_project/pkg/parsing"
	"course_project/pkg/sending"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	messageClient = "Введите запрос в формате:\n\n" +
		"	'SELECT * (или поля через запятую) FROM имя_csv_файла WHERE column_name OP 'example' [AND/OR column_name OP 5]';\n\n" +
		"В конце обязательно поставьте ';'\n\n "

	modeAppend = 0644
)

type App struct {
	Config     Config
	configPath string
}

func New(configPath string) (*App, error) {
	a := &App{configPath: configPath}

	err := a.getFilePath()
	if err != nil {
		return a, err
	}

	a.Config = *NewConfig()
	err = a.Config.ParseConfig(a.configPath)
	if err != nil {
		return a, err
	}

	err = a.getLastVersionCommit()
	if err != nil {
		return a, err
	}

	return a, nil
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

	message := fmt.Sprintf("time: \"%s\", request: \"%s\"\n", time.Now().Format("02.01.2006 15:04:05"), request)
	if err := a.writeToFile(filePath, message); err != nil {
		a.LogError(err)
	}
}

func (a *App) LogError(err error) {
	filePath := a.Config.GetFilePathErrorLog()

	message := fmt.Sprintf("time: \"%s\", error: \"%s\" \n", time.Now().Format("02.01.2006 15:04:05"), err)
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

		_, err = file.WriteString(message)
		if err != nil {
			return err
		}
	} else {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}

		data = append(data, []byte(message)...)

		err = ioutil.WriteFile(filePath, data, modeAppend)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Run() {
	timeOut := a.Config.GetTimeOut()
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	go a.exit(ctx)
	go a.WatchSignals(cancel)

	p := parsing.NewParser()
	request, err := a.getRequestFromClient()
	if err != nil {
		a.LogError(err)
		return
	}
	a.LogAccess(request)

	sel, err := p.Parse(request)
	if err != nil {
		a.LogError(err)
		return
	}

	s, err := sending.New(a.Config.GetCsvFilePath())
	if err != nil {
		a.LogError(err)
		return
	}

	res, err := s.SendRequest(sel)
	if err != nil {
		a.LogError(err)
		return
	}

	err = a.removeOldResultFileCsv()
	if err != nil {
		a.LogError(err)
		return
	}

	err = a.writeResultToCsv(res)
	if err != nil {
		a.LogError(err)
		return
	}

	fmt.Println("\ncount: ", len(res), " result in: ", a.Config.FilePathResultCsv)
}

func (a *App) getRequestFromClient() (string, error) {
	fmt.Println(messageClient)

	in := bufio.NewReader(os.Stdin)

	request, err := in.ReadString('\n')
	if err != nil {
		return request, err
	}

	i := strings.LastIndex(request, ";")
	request = request[:i]
	return request, nil
}

func (a *App) exit(ctx context.Context) {
	<-ctx.Done()
	a.LogError(errors.New("context end"))
	log.Fatal("exit")
}

func (a *App) removeOldResultFileCsv() error {
	if _, err := os.Stat(a.Config.FilePathResultCsv); os.IsExist(err) {
		if err != nil {
			return err
		}
		err = os.Remove(a.Config.FilePathResultCsv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) writeResultToCsv(result [][]string) error {
	outfile, err := os.Create(a.Config.FilePathResultCsv)
	if err != nil {
		return fmt.Errorf("Unable to open output: %s", err)
	}
	defer outfile.Close()

	w := csv.NewWriter(outfile)
	er := w.WriteAll(result)
	if er != nil {
		return er
	}

	if err := w.Error(); err != nil {
		return fmt.Errorf("error writing csv: %s", err)
	}

	return nil
}
