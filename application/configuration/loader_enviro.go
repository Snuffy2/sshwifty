// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Snuffy2/sshwifty/application/log"
)

// environTypeName is the loader name reported when configuration is loaded from
// environment variables.
const (
	environTypeName = "Environment Variable"
)

// parseJsonStringArray parses s as a JSON array of strings, returning the
// resulting slice or a JSON decode error.
func parseJsonStringArray(s string) (r []string, err error) {
	err = json.Unmarshal([]byte(s), &r)
	return
}

// parseEnvUint reads the environment variable named env, trims whitespace, and
// parses it as an unsigned integer of bitSize bits. It returns a descriptive
// error if the variable is missing, empty, or not a valid integer.
func parseEnvUint(env string, bitSize int) (uint64, error) {
	u, err := strconv.ParseUint(strings.TrimSpace(GetEnv(env)), 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf(
			"invalid integer for environment variable %q: %s",
			env,
			err,
		)
	}
	return u, nil
}

// parseEnvUintDefault calls parseEnvUint and returns def if parsing fails,
// making it safe to use for optional environment variables.
func parseEnvUintDefault(env string, def uint64, bitSize int) uint64 {
	if u, err := parseEnvUint(env, bitSize); err == nil {
		return u
	} else {
		return def
	}
}

// Environ creates an environment variable based configuration loader
func Environ() Loader {
	return func(log log.Logger) (string, Configuration, error) {
		log.Info("Loading configuration from environment variables ...")

		// Hooks
		var hooks map[HookType][]HookCommand
		if h := GetEnv("SSHWIFTY_HOOK_BEFORE_CONNECTING"); len(h) > 0 {
			hooks = make(map[HookType][]HookCommand)
			hookBeforeConnecting, err := parseJsonStringArray(h)
			if err != nil {
				return "", Configuration{}, fmt.Errorf(
					"Unable to parse %q: %s",
					"SSHWIFTY_HOOK_BEFORE_CONNECTING",
					err,
				)
			}
			hooks[HOOK_BEFORE_CONNECTING] = []HookCommand{hookBeforeConnecting}
		}

		// Server
		cfgSer := serverInput{
			ListenInterface: GetEnv("SSHWIFTY_LISTENINTERFACE"),
			ListenPort: uint16(
				parseEnvUintDefault("SSHWIFTY_LISTENPORT", 0, 16),
			),
			InitialTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_INITIALTIMEOUT", 0, 32),
			),
			ReadTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_READTIMEOUT", 0, 32),
			),
			WriteTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_WRITETIMEOUT", 0, 32),
			),
			HeartbeatTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_HEARTBEATTIMEOUT", 0, 32),
			),
			ReadDelay: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_READDELAY", 0, 32),
			),
			WriteDelay: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_WRITEDELAY", 0, 32),
			),
			TLSCertificateFile:    GetEnv("SSHWIFTY_TLSCERTIFICATEFILE"),
			TLSCertificateKeyFile: GetEnv("SSHWIFTY_TLSCERTIFICATEKEYFILE"),
			ServerMessage:         GetEnv("SSHWIFTY_SERVERMESSAGE"),
		}

		// Preset
		var presets presetInputs
		presetStr := strings.TrimSpace(GetEnv("SSHWIFTY_PRESETS"))
		if len(presetStr) > 0 {
			presets = make(presetInputs, 0, 16)
			if e := json.Unmarshal([]byte(presetStr), &presets); e != nil {
				return environTypeName, Configuration{}, fmt.Errorf(
					"invalid \"SSHWIFTY_PRESETS\": %s", e)
			}
		}

		cfg, err := commonInput{
			HostName:       GetEnv("SSHWIFTY_HOSTNAME"),
			SharedKey:      GetEnv("SSHWIFTY_SHAREDKEY"),
			PresetAdminKey: GetEnv("SSHWIFTY_PRESET_ADMIN_KEY"),
			DialTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_DIALTIMEOUT", 0, 32),
			),
			Socks5:         GetEnv("SSHWIFTY_SOCKS5"),
			Socks5User:     GetEnv("SSHWIFTY_SOCKS5_USER"),
			Socks5Password: GetEnv("SSHWIFTY_SOCKS5_PASSWORD"),
			Hooks:          hooks,
			HookTimeout: castUintToInt(
				parseEnvUintDefault("SSHWIFTY_HOOKTIMEOUT", 0, 32),
			),
			Servers: serverInputs{cfgSer},
			Presets: presets,
			OnlyAllowPresetRemotes: len(
				GetEnv("SSHWIFTY_ONLYALLOWPRESETREMOTES"),
			) > 0,
		}.concretize()
		return environTypeName, cfg, err
	}
}
