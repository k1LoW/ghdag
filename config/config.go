package config

import (
	"fmt"
	"strings"

	"github.com/k1LoW/ghdag/env"
	"github.com/k1LoW/ghdag/name"
	"github.com/k1LoW/ghdag/task"
)

type Config struct {
	Tasks       task.Tasks       `yaml:"tasks"`
	Env         env.Env          `yaml:"env"`
	LinkedNames name.LinkedNames `yaml:"linkedNames"`
}

func New() *Config {
	return &Config{}
}

func (c *Config) CheckSyntax() error {
	valid := true
	errors := []string{}
	if ok, te := c.Tasks.CheckSyntax(); !ok {
		valid = false
		errors = append(errors, te...)
	}
	if ok, le := c.LinkedNames.CheckSyntax(); !ok {
		valid = false
		errors = append(errors, le...)
	}
	if !valid {
		return fmt.Errorf("invalid config syntax\n%s\n", strings.Join(errors, "\n"))
	}
	return nil
}
