package export

import (
	"context"
	"errors"
	"fmt"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds"
	"github.com/pitshou243/rancher-access-migrator/pkg/client"
	"github.com/pitshou243/rancher-access-migrator/pkg/cluster"
	"github.com/pitshou243/rancher-access-migrator/pkg/migration"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/urfave/cli/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var clusterName, output string

func NewCommand() *cli.Command {
	return &cli.Command{Name: "export", Usage: "Export logical Rancher project and RBAC state", Flags: append(cmds.CommonFlags, &cli.StringFlag{Name: "cluster", Aliases: []string{"source-cluster"}, Destination: &clusterName, Required: true}, &cli.StringFlag{Name: "output", Aliases: []string{"o"}, Destination: &output, Required: true}), Action: run}
}
func run(_ *cli.Context) error {
	ctx := context.Background()
	cfg, err := clientcmd.BuildConfigFromFlags("", cmds.Kubeconfig)
	if err != nil {
		return err
	}
	cl, err := client.New(ctx, cfg)
	if err != nil {
		return err
	}
	var list v3.ClusterList
	if err = cl.Clusters.List(ctx, "", &list, metav1.ListOptions{}); err != nil {
		return err
	}
	var obj *v3.Cluster
	for i := range list.Items {
		if list.Items[i].Name == clusterName || list.Items[i].Spec.DisplayName == clusterName {
			obj = list.Items[i].DeepCopy()
			break
		}
	}
	if obj == nil {
		return errors.New("source cluster not found")
	}
	down := *cfg
	down.Host = cfg.Host + "/k8s/clusters/" + obj.Name
	dc, err := client.New(ctx, &down)
	if err != nil {
		return err
	}
	c := &cluster.Cluster{Obj: obj, Client: dc, ExternalRancher: true}
	if err = c.Populate(ctx, cl); err != nil {
		return err
	}
	var roles v3.RoleTemplateList
	if err = cl.RoleTemplates.List(ctx, "", &roles, metav1.ListOptions{}); err != nil {
		return err
	}
	s := migration.FromCluster(c, cfg.Host, roles.Items)
	if err = migration.WriteFile(output, s); err != nil {
		return err
	}
	fmt.Printf("Exported %d projects, %d namespaces, %d project bindings, %d cluster bindings, and %d custom role templates to %s\n", len(s.Projects), countNS(s), countPB(s), len(s.ClusterBindings), len(s.RoleTemplates), output)
	return nil
}
func countNS(s migration.Snapshot) int {
	n := 0
	for _, p := range s.Projects {
		n += len(p.Namespaces)
	}
	return n
}
func countPB(s migration.Snapshot) int {
	n := 0
	for _, p := range s.Projects {
		n += len(p.Bindings)
	}
	return n
}
