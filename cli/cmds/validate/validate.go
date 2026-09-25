package validate

import (
	"fmt"
	"github.com/pitshou243/rancher-access-migrator/pkg/migration"
	"github.com/urfave/cli/v2"
)

var file string

func NewCommand() *cli.Command {
	return &cli.Command{Name: "validate", Usage: "Validate a portable migration file", Flags: []cli.Flag{&cli.StringFlag{Name: "file", Destination: &file, Required: true}}, Action: run}
}
func run(_ *cli.Context) error {
	s, err := migration.ReadFile(file)
	if err != nil {
		return err
	}
	ns, pb := 0, 0
	for _, p := range s.Projects {
		ns += len(p.Namespaces)
		pb += len(p.Bindings)
	}
	fmt.Printf("Projects: %19d\nNamespaces mapped: %10d\nProject RBAC bindings: %5d\nCluster RBAC bindings: %5d\nCustom RoleTemplates: %6d\n\nSchema validation: OK\n", len(s.Projects), ns, pb, len(s.ClusterBindings), len(s.RoleTemplates))
	return nil
}
