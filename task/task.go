package task

import (
	"fmt"
)

type Env struct {
	Name  string
	Value string
}

type Task struct {
	Id   string
	If   string `yaml:"if,omitempty"`
	Do   *Operation
	Ok   *Operation `yaml:"ok,omitempty"`
	Ng   *Operation `yaml:"ng,omitempty"`
	Env  []Env      `yaml:"env,omitempty"`
	Desc string     `yaml:"desc,omitempty"`
}

type Tasks []*Task

func (tasks Tasks) Find(id string) (*Task, error) {
	for _, t := range tasks {
		if t.Id == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("not found task: %s", id)
}

func (tasks Tasks) MaxLengthID() int {
	length := 0
	for _, t := range tasks {
		if length < len(t.Id) {
			length = len(t.Id)
		}
	}
	return length
}

func (t *Task) CheckSyntax() (bool, []string) {
	valid := true
	prefix := fmt.Sprintf("[%s] ", t.Id)
	errors := []string{}
	if t.Do != nil {
		v, e := t.CheckOperationSyntax(t.Do)
		if !v {
			valid = false
			errors = append(errors, e...)
		}
	} else {
		valid = false
		errors = append(errors, fmt.Sprintf("%snot found `do:` operation", prefix))
	}
	if t.Ok != nil {
		v, e := t.CheckOperationSyntax(t.Ok)
		if !v {
			valid = false
			errors = append(errors, e...)
		}
	}
	if t.Ng != nil {
		v, e := t.CheckOperationSyntax(t.Ng)
		if !v {
			valid = false
			errors = append(errors, e...)
		}
	}
	return valid, errors
}

func (t *Task) CheckOperationSyntax(o *Operation) (bool, []string) {
	valid := true
	prefix := fmt.Sprintf("[%s] ", t.Id)
	errors := []string{}
	c := 0
	if o.Run != "" {
		c++
	}
	if len(o.Labels) > 0 {
		c++
	}
	if len(o.Assignees) > 0 {
		c++
	}
	if o.Comment != "" {
		c++
	}
	if o.Action != "" {
		c++
	}
	if o.Notify != "" {
		c++
	}
	if len(o.Next) > 0 {
		c++
	}
	if c != 1 {
		valid = false
		errors = append(errors, fmt.Sprintf("%sinvalid `%s:` operation (want 1 definition, got %d)", prefix, o.Type, c))
	}
	return valid, errors
}

func (tasks Tasks) CheckSyntax() (bool, []string) {
	valid := true
	errors := []string{}
	ids := map[string]struct{}{}
	for _, t := range tasks {
		if v, e := t.CheckSyntax(); !v {
			valid = false
			errors = append(errors, e...)
		}
		if _, exist := ids[t.Id]; exist {
			valid = false
			errors = append(errors, fmt.Sprintf("duplicate task id: %s", t.Id))
		}
		ids[t.Id] = struct{}{}
	}
	return valid, errors
}
