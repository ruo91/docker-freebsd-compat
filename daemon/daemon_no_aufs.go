// +build exclude_graphdriver_aufs, +build freebsd

package daemon

import (
	"github.com/docker/docker/daemon/graphdriver"
)

func migrateIfAufs(driver graphdriver.Driver, root string) error {
	return nil
}
