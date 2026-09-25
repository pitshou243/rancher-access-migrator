package migration

import "fmt"

type State struct {
	Projects      map[string]Project
	Principals    map[string]bool
	RoleTemplates map[string]bool
}
type Action struct {
	Kind      string `yaml:"kind"`
	Name      string `yaml:"name"`
	Operation string `yaml:"operation"`
	Detail    string `yaml:"detail,omitempty"`
}
type Plan struct {
	Actions              []Action `yaml:"actions"`
	UnresolvedPrincipals []string `yaml:"unresolvedPrincipals,omitempty"`
	MissingRoleTemplates []string `yaml:"missingRoleTemplates,omitempty"`
	Conflicts            []string `yaml:"conflicts,omitempty"`
}

func (p Plan) Safe() bool {
	return len(p.UnresolvedPrincipals) == 0 && len(p.MissingRoleTemplates) == 0 && len(p.Conflicts) == 0
}

func BuildPlan(s Snapshot, dst State) Plan {
	p := Plan{}
	roles := map[string]bool{}
	for k, v := range dst.RoleTemplates {
		roles[k] = v
	}
	for _, r := range s.RoleTemplates {
		roles[r.Name] = true
		op := "create"
		if dst.RoleTemplates[r.Name] {
			op = "skip"
		}
		p.Actions = append(p.Actions, Action{"RoleTemplate", r.Name, op, ""})
	}
	check := func(b Binding, where string) {
		if b.Principal != "" && !dst.Principals[b.Principal] && b.PrincipalType != "serviceAccount" {
			p.UnresolvedPrincipals = appendUnique(p.UnresolvedPrincipals, b.Principal)
		}
		if !roles[b.RoleTemplate] {
			p.MissingRoleTemplates = appendUnique(p.MissingRoleTemplates, b.RoleTemplate)
		}
		p.Actions = append(p.Actions, Action{"Binding", where + ":" + bindingKey(b), "create", "logical identity match"})
	}
	for _, sp := range s.Projects {
		dp, ok := dst.Projects[sp.Name]
		op := "create"
		if ok {
			op = "skip"
		}
		p.Actions = append(p.Actions, Action{"Project", sp.Name, op, "destination ID assigned by Rancher"})
		ns := map[string]bool{}
		for _, n := range dp.Namespaces {
			ns[n] = true
		}
		for _, n := range sp.Namespaces {
			nop := "remap"
			if ns[n] {
				nop = "skip"
			}
			p.Actions = append(p.Actions, Action{"Namespace", n, nop, fmt.Sprintf("project %s", sp.Name)})
		}
		for _, b := range sp.Bindings {
			check(b, "project/"+sp.Name)
		}
	}
	for _, b := range s.ClusterBindings {
		check(b, "cluster")
	}
	return p
}
func appendUnique(a []string, s string) []string {
	for _, v := range a {
		if v == s {
			return a
		}
	}
	return append(a, s)
}
