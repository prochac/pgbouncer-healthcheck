package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"

	"github.com/pkg/errors"
)

func handleHealth(ctx context.Context, _ http.ResponseWriter) error {
	// This is a blind yes/no response
	var err error
	if config.EnhancedCheck {
		_, err = getUsers(ctx, db)
		if err != nil {
			return errors.Wrap(err, "PGbouncer enhanced health check failed")
		}
	} else {
		err = probeLocalPort(config.PGBouncerPort)
		if err != nil {
			return errors.Wrap(err, "PGbouncer port probe check failed")
		}
	}
	if config.CheckDDAgent {
		return getAgentHealth(ctx)
	}
	return nil
}

func probeLocalPort(port int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err == nil {
		conn.Close()
	}
	return err
}

func getAgentHealth(ctx context.Context) error {
	psCmd := exec.CommandContext(ctx, "sudo", "-n", "datadog-agent", "health")
	err := psCmd.Run()
	if err != nil {
		return errors.Wrap(err, "Datadog agent health command failed")
	}
	return nil
}
