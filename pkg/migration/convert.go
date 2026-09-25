package migration

import (
	"github.com/pitshou243/rancher-access-migrator/pkg/cluster"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
)

func FromCluster(c *cluster.Cluster, rancher string, roleTemplates []v3.RoleTemplate) Snapshot {
	s := Snapshot{Source: Source{Rancher: rancher, Cluster: c.Obj.Spec.DisplayName, ClusterID: c.Obj.Name}}
	used := map[string]bool{}
	for _, p := range c.ToMigrate.Projects {
		out := Project{Name: p.Name, Description: p.Obj.Spec.Description}
		for _, n := range p.Namespaces {
			out.Namespaces = append(out.Namespaces, n.Name)
		}
		for _, b := range p.PRTBs {
			principal, kind := Principal(b.Obj.UserName, b.Obj.UserPrincipalName, b.Obj.GroupName, b.Obj.GroupPrincipalName, b.Obj.ServiceAccount)
			out.Bindings = append(out.Bindings, Binding{Principal: principal, PrincipalType: kind, RoleTemplate: b.Obj.RoleTemplateName, ServiceAccount: b.Obj.ServiceAccount})
			used[b.Obj.RoleTemplateName] = true
		}
		s.Projects = append(s.Projects, out)
	}
	for _, b := range c.ToMigrate.CRTBs {
		principal, kind := Principal(b.Obj.UserName, b.Obj.UserPrincipalName, b.Obj.GroupName, b.Obj.GroupPrincipalName, "")
		s.ClusterBindings = append(s.ClusterBindings, Binding{Principal: principal, PrincipalType: kind, RoleTemplate: b.Obj.RoleTemplateName})
		used[b.Obj.RoleTemplateName] = true
	}
	for _, r := range roleTemplates {
		if !used[r.Name] || r.External {
			continue
		}
		out := RoleTemplate{Name: r.Name, Context: r.Context, External: r.External}
		for _, rule := range r.Rules {
			out.Rules = append(out.Rules, Rule{APIGroups: rule.APIGroups, Resources: rule.Resources, Verbs: rule.Verbs})
		}
		s.RoleTemplates = append(s.RoleTemplates, out)
	}
	s.Normalize()
	return s
}
