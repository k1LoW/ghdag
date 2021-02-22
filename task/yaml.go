package task

import "github.com/goccy/go-yaml"

func (t *Task) UnmarshalYAML(data []byte) error {
	if err := yaml.Unmarshal(data, t); err != nil {
		return err
	}
	if t.Do != nil {
		t.Do.Type = OperationTypeDo
	}
	if t.Ok != nil {
		t.Ok.Type = OperationTypeOk
	}
	if t.Ng != nil {
		t.Ng.Type = OperationTypeNg
	}
	return nil
}
