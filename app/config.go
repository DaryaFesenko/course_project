package app

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	TimeOut           time.Duration `yaml:"timeout"`
	FilePathAccessLog string        `yaml:"filePathAccessLog"`
	FilePathErrorLog  string        `yaml:"filePathErrorLog"`
	FilePathCsv       string        `yaml:"filePathCsv"`
	FilePathResultCsv string        `yaml:"filePathResultCsv"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetTimeOut() time.Duration {
	return c.TimeOut
}

func (c *Config) GetCsvFilePath() string {
	return c.FilePathCsv
}

func (c *Config) ParseConfig(configPath string) error {
	var data []byte

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetFilePathAccessLog() string {
	return c.FilePathAccessLog
}

func (c *Config) GetFilePathErrorLog() string {
	return c.FilePathErrorLog
}
