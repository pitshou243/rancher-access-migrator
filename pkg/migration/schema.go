package migration

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

const APIVersion = "rancher-access-migrator.io/v1alpha1"

type Snapshot struct {
	APIVersion      string         `yaml:"apiVersion"`
	Kind            string         `yaml:"kind"`
	Source          Source         `yaml:"source"`
	RoleTemplates   []RoleTemplate `yaml:"roleTemplates,omitempty"`
	Projects        []Project      `yaml:"projects,omitempty"`
	ClusterBindings []Binding      `yaml:"clusterBindings,omitempty"`
}
type Source struct {
	Rancher   string `yaml:"rancher,omitempty"`
	Cluster   string `yaml:"cluster"`
	ClusterID string `yaml:"clusterID,omitempty"`
}
type RoleTemplate struct {
	Name     string `yaml:"name"`
	Context  string `yaml:"context"`
	External bool   `yaml:"external,omitempty"`
	Rules    []Rule `yaml:"rules,omitempty"`
}
type Rule struct {
	APIGroups []string `yaml:"apiGroups,omitempty"`
	Resources []string `yaml:"resources,omitempty"`
	Verbs     []string `yaml:"verbs,omitempty"`
}
type Project struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	Namespaces  []string  `yaml:"namespaces,omitempty"`
	Bindings    []Binding `yaml:"bindings,omitempty"`
}
type Binding struct {
	Principal      string `yaml:"principal"`
	PrincipalType  string `yaml:"principalType"`
	RoleTemplate   string `yaml:"roleTemplate"`
	ServiceAccount string `yaml:"serviceAccount,omitempty"`
}

func (s *Snapshot) Normalize() {
	s.APIVersion, s.Kind = APIVersion, "RancherAccessSnapshot"
	sort.Slice(s.Projects, func(i, j int) bool { return s.Projects[i].Name < s.Projects[j].Name })
	for i := range s.Projects {
		sort.Strings(s.Projects[i].Namespaces)
		sortBindings(s.Projects[i].Bindings)
	}
	sortBindings(s.ClusterBindings)
	sort.Slice(s.RoleTemplates, func(i, j int) bool { return s.RoleTemplates[i].Name < s.RoleTemplates[j].Name })
}
func sortBindings(b []Binding) {
	sort.Slice(b, func(i, j int) bool { return bindingKey(b[i]) < bindingKey(b[j]) })
}
func bindingKey(b Binding) string {
	return b.PrincipalType + "|" + b.Principal + "|" + b.RoleTemplate + "|" + b.ServiceAccount
}

func (s Snapshot) Validate() error {
	if s.APIVersion != APIVersion || s.Kind != "RancherAccessSnapshot" {
		return fmt.Errorf("unsupported snapshot %s %s", s.APIVersion, s.Kind)
	}
	if s.Source.Cluster == "" {
		return errors.New("source.cluster is required")
	}
	seen := map[string]bool{}
	for _, p := range s.Projects {
		if p.Name == "" {
			return errors.New("project name is required")
		}
		if seen[p.Name] {
			return fmt.Errorf("duplicate project %q", p.Name)
		}
		seen[p.Name] = true
	}
	return nil
}
func WriteFile(path string, s Snapshot) error {
	s.Normalize()
	b, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}
func ReadFile(path string) (Snapshot, error) {
	var s Snapshot
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	if err = yaml.Unmarshal(b, &s); err != nil {
		return s, err
	}
	return s, s.Validate()
}

func Principal(user, userPrincipal, group, groupPrincipal, serviceAccount string) (string, string) {
	switch {
	case groupPrincipal != "":
		return groupPrincipal, "group"
	case group != "":
		return group, "group"
	case userPrincipal != "":
		return userPrincipal, "user"
	case user != "":
		return user, "user"
	default:
		return serviceAccount, "serviceAccount"
	}
}
