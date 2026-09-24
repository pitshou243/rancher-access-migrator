package main

import (
	"fmt"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds"
	cmdexport "github.com/pitshou243/rancher-access-migrator/cli/cmds/export"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds/importcmd"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds/interactive"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds/migrate"
	"github.com/pitshou243/rancher-access-migrator/cli/cmds/status"
	cmdvalidate "github.com/pitshou243/rancher-access-migrator/cli/cmds/validate"
	"github.com/pitshou243/rancher-access-migrator/pkg/version"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cmds.NewApp()
	app.Commands = []*cli.Command{
		status.NewCommand(),
		migrate.NewCommand(),
		interactive.NewCommand(),
		cmdexport.NewCommand(),
		cmdvalidate.NewCommand(),
		importcmd.NewCommand(),
	}
	app.Version = fmt.Sprintf("%s (%s)", version.Version, version.GitCommit)

	logrus.SetOutput(io.Discard)
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("exiting tool: %v", err)
		os.Exit(1)
	}
}
