// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PersistPresetIDs writes generated preset IDs back to a JSON configuration file.
//
// It preserves the existing preset order and only updates ID fields. The caller
// must pass the same number of presets that were loaded from filePath.
func PersistPresetIDs(filePath string, presets []Preset) error {
	if filePath == "" {
		return nil
	}

	doc, err := readCommonInputFileDocument(filePath)
	if err != nil {
		return err
	}
	if len(doc.input.Presets) != len(presets) {
		return fmt.Errorf(
			"cannot persist preset IDs: file has %d presets, runtime has %d",
			len(doc.input.Presets),
			len(presets),
		)
	}
	for i := range doc.input.Presets {
		doc.input.Presets[i].ID = presets[i].ID
	}
	return writeCommonInputFileDocument(filePath, doc)
}

// ReplaceFilePresets atomically updates the Presets list in a JSON config file.
func ReplaceFilePresets(filePath string, presets []Preset) error {
	return replaceFilePresets(filePath, presets, nil)
}

// ReplaceFilePresetsWithRuntime atomically updates a JSON config file using
// runtimePresets to distinguish deleted presets from raw entries the runtime
// did not understand.
func ReplaceFilePresetsWithRuntime(
	filePath string,
	presets []Preset,
	runtimePresets []Preset,
) error {
	return replaceFilePresets(filePath, presets, runtimePresets)
}

func replaceFilePresets(
	filePath string,
	presets []Preset,
	runtimePresets []Preset,
) error {
	if filePath == "" {
		return fmt.Errorf("preset config updates require a file-backed configuration")
	}

	doc, err := readCommonInputFileDocument(filePath)
	if err != nil {
		return err
	}
	concrete := runtimePresets
	if concrete == nil {
		var concreteErr error
		concrete, concreteErr = doc.input.Presets.concretize()
		if concreteErr != nil {
			return concreteErr
		}
	}
	doc.input.Presets = mergePresetInputs(
		doc.input.Presets,
		concrete,
		presets,
		runtimePresets,
	)
	return writeCommonInputFileDocument(filePath, doc)
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
		inputs[i] = presetInputFromPreset(preset)
	}
	return inputs
}

func presetInputFromPreset(preset Preset) presetInput {
	return presetInput{
		ID:       preset.ID,
		Title:    preset.Title,
		Type:     preset.Type,
		Host:     preset.Host,
		TabColor: preset.TabColor,
		Meta:     metaInputFromPreset(preset.Meta),
	}
}

func mergePresetInputs(
	raw presetInputs,
	concrete []Preset,
	presets []Preset,
	runtimePresets []Preset,
) presetInputs {
	rawByID := presetInputIndexByID(raw)
	concreteByID := presetMapByID(concrete)
	runtimeByID := presetMapByID(runtimePresets)
	merged := make(presetInputs, 0, len(raw)+len(presets))
	touched := make(map[string]struct{}, len(presets))

	for _, preset := range presets {
		id := strings.TrimSpace(preset.ID)
		touched[id] = struct{}{}
		rawIndex, rawOK := rawByID[id]
		current, currentOK := concreteByID[id]
		if rawOK && currentOK {
			merged = append(merged, mergePresetInput(raw[rawIndex], current, preset))
			continue
		}
		merged = append(merged, presetInputFromPreset(preset))
	}

	for _, input := range raw {
		id := strings.TrimSpace(input.ID)
		if _, ok := touched[id]; ok {
			continue
		}
		if len(runtimeByID) > 0 {
			if _, ok := runtimeByID[id]; ok {
				continue
			}
		}
		merged = append(merged, input)
	}

	return merged
}

func mergePresetInput(raw presetInput, current Preset, preset Preset) presetInput {
	merged := raw
	merged.ID = preset.ID
	merged.Title = preserveRawString(raw.Title, current.Title, preset.Title)
	merged.Type = preserveRawString(raw.Type, current.Type, preset.Type)
	merged.Host = preserveRawString(raw.Host, current.Host, preset.Host)
	merged.TabColor = preserveRawString(raw.TabColor, current.TabColor, preset.TabColor)
	merged.Meta = mergePresetMeta(raw.Meta, current.Meta, preset.Meta)
	return merged
}

func preserveRawString(raw string, current string, next string) string {
	if next == current {
		return raw
	}
	return next
}

func mergePresetMeta(raw Meta, current map[string]string, next map[string]string) Meta {
	merged := Meta{}
	for key, value := range next {
		if currentValue, ok := current[key]; ok && value == currentValue {
			if rawValue, rawOK := raw[key]; rawOK {
				merged[key] = rawValue
				continue
			}
			merged[key] = String(value)
			continue
		}
		merged[key] = String(value)
	}
	if _, ok := next[PresetMetaEncryptedPassword]; ok {
		delete(merged, PresetMetaPassword)
	}
	return merged
}

func copyMeta(meta Meta) Meta {
	if meta == nil {
		return Meta{}
	}
	copied := make(Meta, len(meta))
	for key, value := range meta {
		copied[key] = value
	}
	return copied
}

func metaInputFromPreset(meta map[string]string) Meta {
	input := make(Meta, len(meta))
	for key, value := range meta {
		input[key] = String(value)
	}
	return input
}

func presetInputIndexByID(inputs presetInputs) map[string]int {
	byID := make(map[string]int, len(inputs))
	for i, input := range inputs {
		byID[strings.TrimSpace(input.ID)] = i
	}
	return byID
}

func presetMapByID(presets []Preset) map[string]Preset {
	byID := make(map[string]Preset, len(presets))
	for _, preset := range presets {
		byID[strings.TrimSpace(preset.ID)] = preset
	}
	return byID
}

type commonInputFileDocument struct {
	input commonInput
	raw   map[string]json.RawMessage
	mode  os.FileMode
}

func readCommonInputFileDocument(filePath string) (commonInputFileDocument, error) {
	info, statErr := os.Stat(filePath)
	if statErr != nil {
		return commonInputFileDocument{}, statErr
	}

	data, readErr := os.ReadFile(filePath)
	if readErr != nil {
		return commonInputFileDocument{}, readErr
	}

	cfg := commonInput{}
	if decodeErr := json.Unmarshal(data, &cfg); decodeErr != nil {
		return commonInputFileDocument{}, decodeErr
	}
	raw := map[string]json.RawMessage{}
	if decodeErr := json.Unmarshal(data, &raw); decodeErr != nil {
		return commonInputFileDocument{}, decodeErr
	}
	return commonInputFileDocument{
		input: cfg,
		raw:   raw,
		mode:  info.Mode(),
	}, nil
}

// readCommonInputFile decodes filePath and returns its file mode for rewrites.
func readCommonInputFile(filePath string) (commonInput, os.FileMode, error) {
	doc, err := readCommonInputFileDocument(filePath)
	if err != nil {
		return commonInput{}, 0, err
	}
	return doc.input, doc.mode, nil
}

func writeCommonInputFileDocument(
	filePath string,
	doc commonInputFileDocument,
) error {
	raw := doc.raw
	if raw == nil {
		raw = map[string]json.RawMessage{}
	}
	presets, marshalErr := json.Marshal(doc.input.Presets)
	if marshalErr != nil {
		return marshalErr
	}
	raw["Presets"] = presets
	return writeCommonInputFile(filePath, raw, doc.mode)
}

// writeCommonInputFile atomically rewrites filePath with cfg encoded as JSON.
func writeCommonInputFile(
	filePath string,
	cfg map[string]json.RawMessage,
	mode os.FileMode,
) error {
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
