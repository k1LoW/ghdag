package name

type LinkedName struct {
	Github string
	Slack  string
}

type LinkedNames []*LinkedName

func (l LinkedNames) CheckSyntax() (bool, []string) {
	if len(l) == 0 {
		return true, []string{}
	}
	return true, []string{} // TODO
}

func (l LinkedNames) ToGithubNames(in []string) []string {
	if len(l) == 0 {
		return in
	}
	m := map[string]string{}
	for _, n := range l {
		m[n.Slack] = n.Github
	}
	o := []string{}
	for _, n := range in {
		if v, ok := m[n]; ok {
			o = append(o, v)
		} else {
			o = append(o, n)
		}
	}
	return o
}

func (l LinkedNames) ToSlackNames(in []string) []string {
	if len(l) == 0 {
		return in
	}
	m := map[string]string{}
	for _, n := range l {
		m[n.Github] = n.Slack
	}
	o := []string{}
	for _, n := range in {
		if v, ok := m[n]; ok {
			o = append(o, v)
		} else {
			o = append(o, n)
		}
	}
	return o
}
