package importcmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds"
	"github.com/pitshou243/rancher-access-migrator/pkg/client"
	"github.com/pitshou243/rancher-access-migrator/pkg/migration"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/urfave/cli/v2"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

var file, target string
var dryRun, allowUnresolved bool

func NewCommand() *cli.Command {
	return &cli.Command{Name: "import", Usage: "Plan or apply a portable Rancher access snapshot", Flags: append(cmds.CommonFlags, &cli.StringFlag{Name: "file", Destination: &file, Required: true}, &cli.StringFlag{Name: "target-cluster", Aliases: []string{"target"}, Destination: &target, Required: true}, &cli.BoolFlag{Name: "dry-run", Destination: &dryRun}, &cli.BoolFlag{Name: "allow-unresolved", Destination: &allowUnresolved}), Action: run}
}
func run(_ *cli.Context) error {
	ctx := context.Background()
	s, err := migration.ReadFile(file)
	if err != nil {
		return err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", cmds.Kubeconfig)
	if err != nil {
		return err
	}
	cl, err := client.New(ctx, cfg)
	if err != nil {
		return err
	}
	var clusters v3.ClusterList
	if err = cl.Clusters.List(ctx, "", &clusters, metav1.ListOptions{}); err != nil {
		return err
	}
	var tc *v3.Cluster
	for i := range clusters.Items {
		if clusters.Items[i].Name == target || clusters.Items[i].Spec.DisplayName == target {
			tc = clusters.Items[i].DeepCopy()
			break
		}
	}
	if tc == nil {
		return errors.New("target cluster not found")
	}
	state := migration.State{Projects: map[string]migration.Project{}, Principals: map[string]bool{}, RoleTemplates: map[string]bool{}}
	var projects v3.ProjectList
	if err = cl.Projects.List(ctx, tc.Name, &projects, metav1.ListOptions{}); err != nil {
		return err
	}
	for _, p := range projects.Items {
		state.Projects[p.Spec.DisplayName] = migration.Project{Name: p.Spec.DisplayName}
	}
	var users v3.UserList
	if err = cl.Users.List(ctx, "", &users, metav1.ListOptions{}); err != nil {
		return err
	}
	principalUser := map[string]string{}
	for _, u := range users.Items {
		state.Principals[u.Name] = true
		principalUser[u.Name] = u.Name
		for _, p := range u.PrincipalIDs {
			state.Principals[p] = true
			principalUser[p] = u.Name
		}
	}
	var roles v3.RoleTemplateList
	if err = cl.RoleTemplates.List(ctx, "", &roles, metav1.ListOptions{}); err != nil {
		return err
	}
	for _, r := range roles.Items {
		state.RoleTemplates[r.Name] = true
	}
	plan := migration.BuildPlan(s, state)
	printPlan(plan)
	if !plan.Safe() && !allowUnresolved {
		return errors.New("validation failed; resolve dependencies or explicitly use --allow-unresolved")
	}
	if dryRun {
		fmt.Println("Dry-run: no changes made")
		return nil
	}
	for _, r := range s.RoleTemplates {
		if state.RoleTemplates[r.Name] {
			continue
		}
		obj := &v3.RoleTemplate{ObjectMeta: metav1.ObjectMeta{Name: r.Name}, Context: r.Context, External: r.External}
		for _, rule := range r.Rules {
			obj.Rules = append(obj.Rules, rbacv1.PolicyRule{APIGroups: rule.APIGroups, Resources: rule.Resources, Verbs: rule.Verbs})
		}
		if err = cl.RoleTemplates.Create(ctx, "", obj, nil, metav1.CreateOptions{}); err != nil {
			return err
		}
	}
	projectIDs := map[string]string{}
	for _, p := range projects.Items {
		projectIDs[p.Spec.DisplayName] = p.Name
	}
	for _, p := range s.Projects {
		pid := projectIDs[p.Name]
		if pid == "" {
			pid = "p-" + suffix()
			obj := &v3.Project{ObjectMeta: metav1.ObjectMeta{Name: pid, Namespace: tc.Name}, Spec: v3.ProjectSpec{DisplayName: p.Name, Description: p.Description, ClusterName: tc.Name}}
			if err = cl.Projects.Create(ctx, tc.Name, obj, nil, metav1.CreateOptions{}); err != nil {
				return err
			}
			projectIDs[p.Name] = pid
		}
		for _, b := range p.Bindings {
			obj := bindingProject(tc.Name, pid, b, principalUser)
			if err = cl.ProjectRoleTemplateBindings.Create(ctx, pid, obj, nil, metav1.CreateOptions{}); err != nil && !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	for _, b := range s.ClusterBindings {
		obj := bindingCluster(tc.Name, b, principalUser)
		if err = cl.ClusterRoleTemplateBindings.Create(ctx, tc.Name, obj, nil, metav1.CreateOptions{}); err != nil && !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	down := *cfg
	down.Host = cfg.Host + "/k8s/clusters/" + tc.Name
	dc, err := client.New(ctx, &down)
	if err != nil {
		return err
	}
	for _, p := range s.Projects {
		for _, name := range p.Namespaces {
			ns, err := dc.Namespace.Get(name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if ns.Labels == nil {
				ns.Labels = map[string]string{}
			}
			if ns.Annotations == nil {
				ns.Annotations = map[string]string{}
			}
			ns.Labels["field.cattle.io/projectId"] = projectIDs[p.Name]
			ns.Annotations["field.cattle.io/projectId"] = tc.Name + ":" + projectIDs[p.Name]
			if _, err = dc.Namespace.Update(ns); err != nil {
				return err
			}
		}
	}
	fmt.Println("Import completed")
	return nil
}
func printPlan(p migration.Plan) {
	for _, a := range p.Actions {
		fmt.Printf("%-12s %-8s %s", a.Operation, a.Kind, a.Name)
		if a.Detail != "" {
			fmt.Printf(" (%s)", a.Detail)
		}
		fmt.Println()
	}
	for _, v := range p.UnresolvedPrincipals {
		fmt.Println("UNRESOLVED principal", v)
	}
	for _, v := range p.MissingRoleTemplates {
		fmt.Println("MISSING RoleTemplate", v)
	}
}
func suffix() string {
	const a = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = a[int(b[i])%len(a)]
	}
	return string(b)
}
func setIdentity(user, group *string, b migration.Binding, m map[string]string) {
	if b.PrincipalType == "group" {
		*group = b.Principal
	} else if b.PrincipalType == "user" {
		*user = m[b.Principal]
	}
}
func stable(prefix string, b migration.Binding) string {
	h := sha256.Sum256([]byte(b.PrincipalType + "|" + b.Principal + "|" + b.RoleTemplate + "|" + b.ServiceAccount))
	return prefix + hex.EncodeToString(h[:])[:10]
}
func bindingProject(cluster, pid string, b migration.Binding, m map[string]string) *v3.ProjectRoleTemplateBinding {
	o := &v3.ProjectRoleTemplateBinding{ObjectMeta: metav1.ObjectMeta{Name: stable("prtb-", b), Namespace: pid}, ProjectName: cluster + ":" + pid, RoleTemplateName: b.RoleTemplate, ServiceAccount: b.ServiceAccount}
	setIdentity(&o.UserName, &o.GroupPrincipalName, b, m)
	return o
}
func bindingCluster(cluster string, b migration.Binding, m map[string]string) *v3.ClusterRoleTemplateBinding {
	o := &v3.ClusterRoleTemplateBinding{ObjectMeta: metav1.ObjectMeta{Name: stable("crtb-", b), Namespace: cluster}, ClusterName: cluster, RoleTemplateName: b.RoleTemplate}
	setIdentity(&o.UserName, &o.GroupPrincipalName, b, m)
	return o
}
