// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Snuffy2/sshwifty/application/log"
)

// fileTypeName is the loader name reported when configuration is loaded from a
// JSON file.
const (
	fileTypeName = "File"
)

// loadFile opens filePath, JSON-decodes it into a commonInput, and returns the
// resulting Configuration. It returns the fileTypeName string along with the
// configuration or the first error encountered.
func loadFile(filePath string) (string, Configuration, error) {
	f, fErr := os.Open(filePath)
	if fErr != nil {
		return fileTypeName, Configuration{}, fErr
	}
	defer f.Close()
	cfg := commonInput{}
	jDecoder := json.NewDecoder(f)
	jDecodeErr := jDecoder.Decode(&cfg)
	if jDecodeErr != nil {
		return fileTypeName, Configuration{}, jDecodeErr
	}
	finalCfg, err := cfg.concretize()
	return fileTypeName, finalCfg, err
}

// CustomFile creates a configuration file loader that loads configuration from
// the specified file path
func CustomFile(customPath string) Loader {
	return func(log log.Logger) (string, Configuration, error) {
		log.Info("Loading configuration from: %s", customPath)
		return loadFile(customPath)
	}
}

// DefaultFile creates a configuration file loader that loads configuration from
// one of the default file path
func DefaultFile() Loader {
	return func(log log.Logger) (string, Configuration, error) {
		log.Info("Loading configuration from one of the default " +
			"configuration files ...")
		fallbackFileSearchList := make([]string, 0, 3)

		// ~/.config/sshwifty.conf.json
		if u, userErr := user.Current(); userErr == nil {
			fallbackFileSearchList = append(
				fallbackFileSearchList,
				filepath.Join(u.HomeDir, ".config", "sshwifty.conf.json"))
		}

		// /etc/sshwifty.conf.json
		fallbackFileSearchList = append(
			fallbackFileSearchList,
			filepath.Join("/", "etc", "sshwifty.conf.json"),
		)

		// sshwifty.conf.json located at the same directory as Sshwifty bin
		if ex, exErr := os.Executable(); exErr == nil {
			fallbackFileSearchList = append(
				fallbackFileSearchList,
				filepath.Join(filepath.Dir(ex), "sshwifty.conf.json"))
		}

		// Search given locations to select the config file
		for f := range fallbackFileSearchList {
			if fInfo, fErr := os.Stat(fallbackFileSearchList[f]); fErr != nil {
				continue
			} else if fInfo.IsDir() {
				continue
			} else {
				log.Info("Configuration file \"%s\" has been selected",
					fallbackFileSearchList[f])
				return loadFile(fallbackFileSearchList[f])
			}
		}
		return fileTypeName, Configuration{}, fmt.Errorf(
			"Configuration file was not specified. Also tried fallback files "+
				"\"%s\", but none of them was available",
			strings.Join(fallbackFileSearchList, "\", \""))
	}
}
