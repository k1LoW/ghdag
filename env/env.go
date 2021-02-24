package env

import (
	"bytes"
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
	em := envMap()
	for k, v := range e {
		parsed, err := parseWithEnviron(v, em)
		if err != nil {
			return err
		}
		os.Setenv(k, parsed)
	}
	return nil
}

func Revert(envCache []string) {
	for _, e := range os.Environ() {
		splitted := strings.Split(e, "=")
		os.Unsetenv(splitted[0])
	}
	for _, e := range envCache {
		splitted := strings.Split(e, "=")
		os.Setenv(splitted[0], splitted[1])
	}
}

func parseWithEnviron(v string, envMap map[string]string) (string, error) {
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

func envMap() map[string]string {
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
