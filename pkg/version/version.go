package version

import "strings"

var (
	Program      = "rancher-access-migrator"
	ProgramUpper = strings.ToUpper(Program)
	Version      = "dev"
	GitCommit    = "HEAD"
)
