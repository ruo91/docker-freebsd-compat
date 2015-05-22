package main

import (
	"os/exec"
	"strings"
	"time"

	"github.com/go-check/check"
)

func (s *DockerSuite) TestCliStatsNoStream(c *check.C) {
	out, _, err := runCommandWithOutput(exec.Command(dockerBinary, "run", "-d", "busybox", "top"))
	if err != nil {
		c.Fatalf("Error on container creation: %v, output: %s", err, out)
	}
	id := strings.TrimSpace(out)
	if err := waitRun(id); err != nil {
		c.Fatalf("error waiting for container to start: %v", err)
	}

	statsCmd := exec.Command(dockerBinary, "stats", "--no-stream", id)
	chErr := make(chan error)
	go func() {
		chErr <- statsCmd.Run()
	}()

	select {
	case err := <-chErr:
		if err != nil {
			c.Fatalf("Error running stats: %v", err)
		}
	case <-time.After(2 * time.Second):
		statsCmd.Process.Kill()
		c.Fatalf("stats did not return immediately when not streaming")
	}
}
