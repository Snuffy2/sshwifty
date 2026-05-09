// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const moshDefaultServerCommand = "mosh-server"

var errMoshConnectLineNotFound = errors.New("mosh connect line not found")

type moshConnectInfo struct {
	Port int
	Key  string
}

type moshServerCommand struct {
	Binary string
	Args   []string
}

func parseMoshConnectLine(output string) (moshConnectInfo, error) {
	var lastConnectErr error

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}

		if len(fields) < 2 || fields[0] != "MOSH" || fields[1] != "CONNECT" {
			continue
		}

		if len(fields) < 4 {
			lastConnectErr = errors.New("mosh connect line missing port or key")
			continue
		}

		port, err := strconv.Atoi(fields[2])
		if err != nil {
			lastConnectErr = fmt.Errorf("invalid mosh port %q: %w", fields[2], err)
			continue
		}

		if port < 1 || port > 65535 {
			lastConnectErr = fmt.Errorf("invalid mosh port %d", port)
			continue
		}

		key := strings.TrimSpace(fields[3])
		if key == "" {
			lastConnectErr = errors.New("mosh key cannot be blank")
			continue
		}

		return moshConnectInfo{Port: port, Key: key}, nil
	}

	if lastConnectErr != nil {
		return moshConnectInfo{}, lastConnectErr
	}

	return moshConnectInfo{}, errMoshConnectLineNotFound
}

func buildMoshServerCommand(serverPath string) moshServerCommand {
	serverPath = strings.TrimSpace(serverPath)
	if serverPath == "" {
		serverPath = moshDefaultServerCommand
	}

	return moshServerCommand{
		Binary: serverPath,
		Args:   []string{"new", "-s", "-c", "256", "-l", "LANG=en_US.UTF-8"},
	}
}

func renderMoshServerCommand(meta map[string]string) string {
	serverPath := ""
	if meta != nil {
		serverPath = meta["Mosh Server"]
	}

	serverCommand := buildMoshServerCommand(serverPath)
	tokens := make([]string, 0, len(serverCommand.Args)+1)
	tokens = append(tokens, shellQuoteMoshCommandToken(serverCommand.Binary))

	for _, arg := range serverCommand.Args {
		tokens = append(tokens, shellQuoteMoshCommandToken(arg))
	}

	return strings.Join(tokens, " ")
}

func shellQuoteMoshCommandToken(token string) string {
	if token == "" {
		return "''"
	}

	return "'" + strings.ReplaceAll(token, "'", `'"'"'`) + "'"
}
