package config

import (
	"fmt"
	"strings"

	"github.com/k1LoW/ghdag/task"
)

type Config struct {
	Tasks task.Tasks `yaml:"tasks"`
}

func (c *Config) CheckSyntax() error {
	valid, errors := c.Tasks.CheckSyntax()
	if !valid {
		return fmt.Errorf("invalid config syntax\n%s\n", strings.Join(errors, "\n"))
	}
	return nil
}
