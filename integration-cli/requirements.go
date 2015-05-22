package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/go-check/check"
)

type TestCondition func() bool

type TestRequirement struct {
	Condition   TestCondition
	SkipMessage string
}

// List test requirements
var (
	daemonExecDriver string

	SameHostDaemon = TestRequirement{
		func() bool { return isLocalDaemon },
		"Test requires docker daemon to runs on the same machine as CLI",
	}
	UnixCli = TestRequirement{
		func() bool { return isUnixCli },
		"Test requires posix utilities or functionality to run.",
	}
	ExecSupport = TestRequirement{
		func() bool { return supportsExec },
		"Test requires 'docker exec' capabilities on the tested daemon.",
	}
	Network = TestRequirement{
		func() bool {
			resp, err := http.Get("http://hub.docker.com")
			if resp != nil {
				resp.Body.Close()
			}
			return err == nil
		},
		"Test requires network availability, environment variable set to none to run in a non-network enabled mode.",
	}
	RegistryHosting = TestRequirement{
		func() bool {
			// for now registry binary is built only if we're running inside
			// container through `make test`. Figure that out by testing if
			// registry binary is in PATH.
			_, err := exec.LookPath(v2binary)
			return err == nil
		},
		fmt.Sprintf("Test requires an environment that can host %s in the same host", v2binary),
	}
	NativeExecDriver = TestRequirement{
		func() bool {
			if daemonExecDriver == "" {
				// get daemon info
				status, body, err := sockRequest("GET", "/info", nil)
				if err != nil || status != http.StatusOK {
					log.Fatalf("sockRequest failed for /info: %v", err)
				}

				type infoJSON struct {
					ExecutionDriver string
				}
				var info infoJSON
				if err = json.Unmarshal(body, &info); err != nil {
					log.Fatalf("unable to unmarshal body: %v", err)
				}

				daemonExecDriver = info.ExecutionDriver
			}

			return strings.HasPrefix(daemonExecDriver, "native")
		},
		"Test requires the native (libcontainer) exec driver.",
	}

	NotOverlay = TestRequirement{
		func() bool {
			cmd := exec.Command("grep", "^overlay / overlay", "/proc/mounts")
			if err := cmd.Run(); err != nil {
				return true
			}
			return false
		},
		"Test requires underlying root filesystem not be backed by overlay.",
	}
)

// testRequires checks if the environment satisfies the requirements
// for the test to run or skips the tests.
func testRequires(c *check.C, requirements ...TestRequirement) {
	for _, r := range requirements {
		if !r.Condition() {
			c.Skip(r.SkipMessage)
		}
	}
}
