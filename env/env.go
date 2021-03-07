package env

import (
	"bytes"
	"encoding/csv"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var (
	re  = regexp.MustCompile(`\${\s*([^{}]+)\s*}`)
	re2 = regexp.MustCompile(`{{([^\.])`)
	re3 = regexp.MustCompile(`__GHDAG__(.)`)
)

type Env map[string]string

func (e Env) Setenv() error {
	em := EnvMap()
	for k, v := range e {
		parsed, err := ParseWithEnviron(v, em)
		if err != nil {
			return err
		}
		if err := os.Setenv(k, parsed); err != nil {
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

func ParseWithEnviron(v string, envMap map[string]string) (string, error) {
	if !re.MatchString(v) {
		return v, nil
	}
	replaced := re.ReplaceAllString(v, "{{.$1}}")
	replaced2 := re2.ReplaceAllString(replaced, "__GHDAG__$1")
	tmpl, err := template.New("config").Parse(replaced2)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, envMap)
	if err != nil {
		return "", err
	}
	return re3.ReplaceAllString(buf.String(), "{{$1"), nil
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

func ToSlice(in string) ([]string, error) {
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
