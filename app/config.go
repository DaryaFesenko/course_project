package app

import "time"

type Config struct {
	TimeOut time.Duration `yaml:"timeout"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetTimeOut() time.Duration {
	return c.TimeOut
}
