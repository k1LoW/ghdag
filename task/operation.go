package task

type OperationType int

const (
	OperationTypeDo OperationType = iota + 1
	OperationTypeOk
	OperationTypeNg
)

var operationTypes = [...]string{"", "do", "ok", "ng"}

func (t OperationType) String() string {
	return operationTypes[t]
}

type Operation struct {
	Type      OperationType `yaml:"-"`
	Run       string        `yaml:"run,omitempty"`
	Labels    []string      `yaml:"labels,omitempty"`
	Assignees []string      `yaml:"assignees,omitempty"`
	Comment   string        `yaml:"comment,omitempty"`
	State    string        `yaml:"state,omitempty"`
	Notify    string        `yaml:"notify,omitempty"`
	Next      []string      `yaml:"next,omitempty"`
}
