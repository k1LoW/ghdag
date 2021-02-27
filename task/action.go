package task

type ActionType int

const (
	ActionTypeDo ActionType = iota + 1
	ActionTypeOk
	ActionTypeNg
)

var actionTypes = [...]string{"", "do", "ok", "ng"}

func (t ActionType) String() string {
	return actionTypes[t]
}

type Action struct {
	Type      ActionType `yaml:"-"`
	Run       string     `yaml:"run,omitempty"`
	Labels    []string   `yaml:"labels,omitempty"`
	Assignees []string   `yaml:"assignees,omitempty"`
	Reviewers []string   `yaml:"reviewers,omitempty"`
	Comment   string     `yaml:"comment,omitempty"`
	State     string     `yaml:"state,omitempty"`
	Notify    string     `yaml:"notify,omitempty"`
	Next      []string   `yaml:"next,omitempty"`
}
