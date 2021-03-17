package name

import "fmt"

type LinkedName struct {
	Github string
	Slack  string
}

type LinkedNames []*LinkedName

func (l LinkedNames) CheckSyntax() (bool, []string) {
	valid := true
	errors := []string{}
	if len(l) == 0 {
		return valid, errors
	}

	g := map[string]int{}
	s := map[string]int{}
	for i, n := range l {
		if j, ok := g[n.Github]; ok {
			valid = false
			errors = append(errors, fmt.Sprintf("'%s' is found in both linkedNames[%d].github and linkedNames[%d].github", n.Github, i, j))
		} else {
			g[n.Github] = i
		}

		if j, ok := g[n.Slack]; ok {
			valid = false
			errors = append(errors, fmt.Sprintf("'%s' is found in both linkedNames[%d].slack and linkedNames[%d].slack", n.Slack, i, j))
		} else {
			g[n.Slack] = i
		}
	}

	for n, i := range g {
		if j, ok := s[n]; ok {
			valid = false
			errors = append(errors, fmt.Sprintf("'%s' is found in both linkedNames[%d].github and linkedNames[%d].slack", n, i, j))
		}
	}

	return valid, errors
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
