package env

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	re  = regexp.MustCompile(`\${\s*([^{}]+)\s*}`)
	re2 = regexp.MustCompile(`{{([^\.])`)
	re3 = regexp.MustCompile(`__GHDAG__(.)`)
)

type Env map[string]string

func (e Env) Setenv() error {
	for k, v := range e {
		if err := os.Setenv(k, os.ExpandEnv(v)); err != nil {
			return err
		}
	}
	return nil
}

func Revert(envCache []string) error {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		if err := os.Unsetenv(splitted[0]); err != nil {
			return err
		}
	}
	for _, e := range envCache {
		splitted := strings.Split(e, "=")
		if err := os.Setenv(splitted[0], splitted[1]); err != nil {
			return err
		}
	}
	return nil
}

func Split(in string) ([]string, error) {
	if in == "" {
		return []string{}, nil
	}
	sq := strings.Count(in, "'")
	if sq > 0 && (sq%2 == 0) {
		in = strings.Replace(in, `'`, `"`, -1)
	}
	r := csv.NewReader(strings.NewReader(in))
	if !strings.Contains(in, ",") {
		r.Comma = ' '
	}
	c, err := r.Read()
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, s := range c {
		trimed := strings.Trim(s, " ")
		if trimed == "" {
			continue
		}
		res = append(res, trimed)
	}
	return res, nil
}

func Join(in []string) string {
	formated := []string{}
	for _, s := range in {
		if strings.Contains(s, " ") {
			s = fmt.Sprintf(`"%s"`, s)
		}
		formated = append(formated, s)
	}
	return strings.Join(formated, " ")
}

func EnvMap() map[string]string {
	m := map[string]string{}
	for _, kv := range os.Environ() {
		if !strings.Contains(kv, "=") {
			continue
		}
		parts := strings.SplitN(kv, "=", 2)
		k := parts[0]
		if len(parts) < 2 {
			m[k] = ""
			continue
		}
		m[k] = parts[1]
	}
	return m
}

func GetenvAsBool(k string) bool {
	if os.Getenv(k) == "" || strings.ToLower(os.Getenv(k)) == "false" || os.Getenv(k) == "0" {
		return false
	}
	return true
}
