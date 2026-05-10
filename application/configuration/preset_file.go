// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PersistPresetIDs writes generated preset IDs back to a JSON configuration file.
//
// It preserves the existing preset order and only updates ID fields. The caller
// must pass the same number of presets that were loaded from filePath.
func PersistPresetIDs(filePath string, presets []Preset) error {
	if filePath == "" {
		return nil
	}

	raw, mode, err := readCommonInputFile(filePath)
	if err != nil {
		return err
	}
	if len(raw.Presets) != len(presets) {
		return fmt.Errorf(
			"cannot persist preset IDs: file has %d presets, runtime has %d",
			len(raw.Presets),
			len(presets),
		)
	}
	for i := range raw.Presets {
		raw.Presets[i].ID = presets[i].ID
	}
	return writeCommonInputFile(filePath, raw, mode)
}

// ReplaceFilePresets atomically replaces the Presets list in a JSON config file.
func ReplaceFilePresets(filePath string, presets []Preset) error {
	if filePath == "" {
		return fmt.Errorf("preset config updates require a file-backed configuration")
	}

	raw, mode, err := readCommonInputFile(filePath)
	if err != nil {
		return err
	}
	raw.Presets = presetInputsFromPresets(presets)
	return writeCommonInputFile(filePath, raw, mode)
}

// PresetConfigWritable reports whether filePath points to a writable config file.
func PresetConfigWritable(filePath string) bool {
	if filePath == "" {
		return false
	}
	f, err := os.OpenFile(filePath, os.O_RDWR, 0)
	if err != nil {
		return false
	}
	if closeErr := f.Close(); closeErr != nil {
		return false
	}
	tmp, createErr := os.CreateTemp(
		filepath.Dir(filePath),
		filepath.Base(filePath)+".writable.*.tmp",
	)
	if createErr != nil {
		return false
	}
	tmpName := tmp.Name()
	if closeErr := tmp.Close(); closeErr != nil {
		_ = os.Remove(tmpName)
		return false
	}
	return os.Remove(tmpName) == nil
}

// presetInputsFromPresets converts normalized presets back to file input shape.
func presetInputsFromPresets(presets []Preset) presetInputs {
	inputs := make(presetInputs, len(presets))
	for i, preset := range presets {
		meta := make(Meta, len(preset.Meta))
		for key, value := range preset.Meta {
			meta[key] = String(value)
		}
		inputs[i] = presetInput{
			ID:       preset.ID,
			Title:    preset.Title,
			Type:     preset.Type,
			Host:     preset.Host,
			TabColor: preset.TabColor,
			Meta:     meta,
		}
	}
	return inputs
}

// readCommonInputFile decodes filePath and returns its file mode for rewrites.
func readCommonInputFile(filePath string) (commonInput, os.FileMode, error) {
	info, statErr := os.Stat(filePath)
	if statErr != nil {
		return commonInput{}, 0, statErr
	}

	f, openErr := os.Open(filePath)
	if openErr != nil {
		return commonInput{}, 0, openErr
	}
	defer f.Close()

	cfg := commonInput{}
	if decodeErr := json.NewDecoder(f).Decode(&cfg); decodeErr != nil {
		return commonInput{}, 0, decodeErr
	}
	return cfg, info.Mode(), nil
}

// writeCommonInputFile atomically rewrites filePath with cfg encoded as JSON.
func writeCommonInputFile(filePath string, cfg commonInput, mode os.FileMode) error {
	tmp, createErr := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".*.tmp")
	if createErr != nil {
		return createErr
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if encodeErr := encoder.Encode(cfg); encodeErr != nil {
		tmp.Close()
		return encodeErr
	}
	if chmodErr := tmp.Chmod(mode); chmodErr != nil {
		tmp.Close()
		return chmodErr
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return closeErr
	}
	return os.Rename(tmpName, filePath)
}
