// +build !exclude_graphdriver_overlay, +build !freebsd

package daemon

import (
	_ "github.com/docker/docker/daemon/graphdriver/overlay"
)
