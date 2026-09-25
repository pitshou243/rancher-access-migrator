package migration

import (
	"path/filepath"
	"testing"
)

func sample() Snapshot {
	return Snapshot{Source: Source{Cluster: "old", ClusterID: "c-old"}, Projects: []Project{{Name: "Dev", Namespaces: []string{"z", "a"}, Bindings: []Binding{{Principal: "github_user://u", PrincipalType: "user", RoleTemplate: "project-member"}}}}, ClusterBindings: []Binding{{Principal: "github_team://ops", PrincipalType: "group", RoleTemplate: "cluster-owner"}}, RoleTemplates: []RoleTemplate{{Name: "project-member", Context: "project"}, {Name: "cluster-owner", Context: "cluster"}}}
}
func TestSerializationRoundTrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.yaml")
	s := sample()
	if err := WriteFile(p, s); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if got.Projects[0].Namespaces[0] != "a" {
		t.Fatalf("not normalized: %#v", got)
	}
}
func TestProjectAndClusterRemapping(t *testing.T) {
	p := BuildPlan(sample(), State{Projects: map[string]Project{}, Principals: map[string]bool{"github_user://u": true, "github_team://ops": true}, RoleTemplates: map[string]bool{}})
	if !p.Safe() {
		t.Fatalf("unexpected unsafe plan: %#v", p)
	}
	if p.Actions[2].Operation != "create" {
		t.Fatalf("expected project create: %#v", p.Actions)
	}
}
func TestMissingPrincipalAndRole(t *testing.T) {
	s := sample()
	s.RoleTemplates = nil
	p := BuildPlan(s, State{Projects: map[string]Project{}, Principals: map[string]bool{}, RoleTemplates: map[string]bool{}})
	if p.Safe() || len(p.UnresolvedPrincipals) != 2 || len(p.MissingRoleTemplates) != 2 {
		t.Fatalf("validation failed: %#v", p)
	}
}
func TestIdempotentPlan(t *testing.T) {
	s := sample()
	p := BuildPlan(s, State{Projects: map[string]Project{"Dev": s.Projects[0]}, Principals: map[string]bool{"github_user://u": true, "github_team://ops": true}, RoleTemplates: map[string]bool{"project-member": true, "cluster-owner": true}})
	for _, a := range p.Actions {
		if a.Kind == "Project" && a.Operation != "skip" {
			t.Fatal("project duplicated")
		}
		if a.Kind == "Namespace" && a.Operation != "skip" {
			t.Fatal("namespace duplicated")
		}
	}
}
func TestPrincipalPrecedence(t *testing.T) {
	p, k := Principal("u", "oidc://u", "g", "oidc_group://g", "sa")
	if p != "oidc_group://g" || k != "group" {
		t.Fatalf("got %s %s", p, k)
	}
}
