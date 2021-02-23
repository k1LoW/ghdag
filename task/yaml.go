package task

import "github.com/goccy/go-yaml"

func (t *Task) UnmarshalYAML(data []byte) error {
	raw := &struct {
		Id   string
		If   string `yaml:"if,omitempty"`
		Do   *Action
		Ok   *Action `yaml:"ok,omitempty"`
		Ng   *Action `yaml:"ng,omitempty"`
		Env  []Env      `yaml:"env,omitempty"`
		Desc string     `yaml:"desc,omitempty"`
	}{}
	if err := yaml.Unmarshal(data, raw); err != nil {
		return err
	}
	t.Id = raw.Id
	t.If = raw.If
	t.Do = raw.Do
	if t.Do != nil {
		t.Do.Type = ActionTypeDo
	}
	t.Ok = raw.Ok
	if t.Ok != nil {
		t.Ok.Type = ActionTypeOk
	}
	t.Ng = raw.Ng
	if t.Ng != nil {
		t.Ng.Type = ActionTypeNg
	}
	t.Env = raw.Env
	t.Desc = raw.Desc

	return nil
}
