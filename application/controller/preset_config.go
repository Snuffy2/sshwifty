// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
)

const (
	preserveHiddenPresetPasswordsHeader = "X-Preserve-Hidden-Preset-Passwords"
	presetFingerprintIDHeader           = "X-Preset-Fingerprint-ID"
	presetAdminKeyHeader                = "X-Preset-Admin-Key"
	maxPresetConfigRequestBytes         = 256 * 1024
	maxPresetConfigPresets              = 512
	maxPresetConfigStringBytes          = 4096
	maxPresetFingerprintBytes           = 256
)

// presetConfig handles backend preset configuration reads and writes.
type presetConfig struct {
	baseController

	commonCfg configuration.Common
	commands  command.Commands
}

// presetConfigResponse is the JSON envelope returned by preset config APIs.
type presetConfigResponse struct {
	Presets []socketRemotePreset `json:"presets"`
}

// presetConfigRequest is the JSON envelope accepted by preset config APIs.
type presetConfigRequest struct {
	Presets []socketRemotePreset `json:"presets"`
}

// newPresetConfig builds the backend-only preset configuration controller.
func newPresetConfig(
	commonCfg configuration.Common,
	commands command.Commands,
) presetConfig {
	return presetConfig{
		commonCfg: commonCfg,
		commands:  commands,
	}
}

// Get returns the current live preset list.
func (p presetConfig) Get(
	w *ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	if err := p.requireAuth(r); err != nil {
		return err
	}
	return p.writePresets(w, p.commonCfg.CurrentPresets())
}

// Put replaces the full preset list, adding missing IDs and persisting the file.
func (p presetConfig) Put(
	w *ResponseWriter,
	r *http.Request,
	l log.Logger,
) error {
	if p.commonCfg.SharedKey == "" {
		return NewError(
			http.StatusForbidden,
			"Preset updates require SharedKey authentication",
		)
	}
	if err := p.requireAuth(r); err != nil {
		return err
	}
	if !p.commonCfg.PresetConfigWritable() {
		return NewError(
			http.StatusConflict,
			"Preset updates require a writable file-backed configuration",
		)
	}

	var request presetConfigRequest
	if err := json.NewDecoder(http.MaxBytesReader(
		w,
		r.Body,
		maxPresetConfigRequestBytes,
	)).Decode(&request); err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	if err := validatePresetConfigRequest(request); err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}

	presets := make([]configuration.Preset, len(request.Presets))
	for i, preset := range request.Presets {
		presets[i] = configuration.Preset{
			ID:       strings.TrimSpace(preset.ID),
			Title:    preset.Title,
			Type:     preset.Type,
			Host:     preset.Host,
			TabColor: preset.TabColor,
			Meta:     preset.Meta,
		}
	}

	p.lockPresetUpdates()
	defer p.unlockPresetUpdates()

	currentPresets := p.commonCfg.CurrentPresets()
	fingerprintOnly := r.Header.Get(preserveHiddenPresetPasswordsHeader) == "yes"
	if fingerprintOnly {
		presets = preserveHiddenPresetPasswords(presets, currentPresets)
		if err := validateFingerprintOnlyPresetUpdate(
			presets,
			currentPresets,
			strings.TrimSpace(r.Header.Get(presetFingerprintIDHeader)),
		); err != nil {
			return NewError(http.StatusBadRequest, err.Error())
		}
	} else if err := p.requirePresetAdminAuth(r); err != nil {
		return err
	}
	normalized, _, err := configuration.EnsurePresetIDs(presets)
	if err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	normalized, _, err = configuration.ApplyPresetSecrets(normalized)
	if err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	normalized, err = p.commands.Reconfigure(normalized)
	if err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	if err := configuration.ReplaceFilePresetsWithRuntime(
		p.commonCfg.SourceFile,
		normalized,
		currentPresets,
	); err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	if p.commonCfg.PresetRepository != nil {
		p.commonCfg.PresetRepository.Replace(normalized)
	}
	return p.writePresets(w, normalized)
}

func (p presetConfig) lockPresetUpdates() {
	if p.commonCfg.PresetRepository != nil {
		p.commonCfg.PresetRepository.LockUpdates()
	}
}

func (p presetConfig) unlockPresetUpdates() {
	if p.commonCfg.PresetRepository != nil {
		p.commonCfg.PresetRepository.UnlockUpdates()
	}
}

func validatePresetConfigRequest(request presetConfigRequest) error {
	if len(request.Presets) > maxPresetConfigPresets {
		return fmt.Errorf(
			"preset count %d exceeds maximum %d",
			len(request.Presets),
			maxPresetConfigPresets,
		)
	}
	for _, preset := range request.Presets {
		if stringTooLong(preset.ID, configuration.MaxPresetIDLength) {
			return fmt.Errorf("preset ID exceeds maximum length")
		}
		if stringTooLong(preset.Title, maxPresetConfigStringBytes) ||
			stringTooLong(preset.Type, maxPresetConfigStringBytes) ||
			stringTooLong(preset.Host, maxPresetConfigStringBytes) ||
			stringTooLong(preset.TabColor, maxPresetConfigStringBytes) {
			return fmt.Errorf("preset fields exceed maximum length")
		}
		for key, value := range preset.Meta {
			if stringTooLong(key, maxPresetConfigStringBytes) ||
				stringTooLong(value, maxPresetConfigStringBytes) {
				return fmt.Errorf("preset metadata exceeds maximum length")
			}
		}
	}
	return nil
}

func stringTooLong(value string, maxBytes int) bool {
	return len(value) > maxBytes
}

func validateFingerprintOnlyPresetUpdate(
	presets []configuration.Preset,
	current []configuration.Preset,
	targetPresetID string,
) error {
	if targetPresetID == "" {
		return fmt.Errorf("fingerprint save requires a preset ID")
	}
	if len(presets) != len(current) {
		return fmt.Errorf("fingerprint save cannot add or remove presets")
	}
	currentByID := make(map[string]configuration.Preset, len(current))
	for _, preset := range current {
		currentByID[preset.ID] = preset
	}
	changedFingerprint := false
	for _, preset := range presets {
		currentPreset, ok := currentByID[preset.ID]
		if !ok {
			return fmt.Errorf("fingerprint save cannot add preset %q", preset.ID)
		}
		if preset.Title != currentPreset.Title ||
			preset.Type != currentPreset.Type ||
			preset.Host != currentPreset.Host ||
			preset.TabColor != currentPreset.TabColor {
			return fmt.Errorf("fingerprint save cannot change preset %q", preset.ID)
		}
		if !samePresetMetaExceptFingerprint(preset.Meta, currentPreset.Meta) {
			return fmt.Errorf(
				"fingerprint save cannot change non-fingerprint metadata for preset %q",
				preset.ID,
			)
		}
		if preset.Meta["Fingerprint"] == currentPreset.Meta["Fingerprint"] {
			continue
		}
		if preset.ID != targetPresetID {
			return fmt.Errorf("fingerprint save cannot change preset %q", preset.ID)
		}
		if err := validatePresetFingerprint(
			preset.Meta["Fingerprint"],
		); err != nil {
			return err
		}
		if changedFingerprint {
			return fmt.Errorf("fingerprint save can only change one preset")
		}
		changedFingerprint = true
	}
	if !changedFingerprint {
		return fmt.Errorf("fingerprint save did not change preset %q", targetPresetID)
	}
	return nil
}

func validatePresetFingerprint(fingerprint string) error {
	fingerprint = strings.TrimSpace(fingerprint)
	if fingerprint == "" {
		return fmt.Errorf("fingerprint save cannot remove a fingerprint")
	}
	if len(fingerprint) > maxPresetFingerprintBytes {
		return fmt.Errorf("fingerprint exceeds maximum length")
	}
	if !strings.Contains(fingerprint, ":") {
		return fmt.Errorf("fingerprint has invalid format")
	}
	for _, r := range fingerprint {
		if r < 0x21 || r == 0x7f {
			return fmt.Errorf("fingerprint has invalid format")
		}
	}
	return nil
}

func samePresetMetaExceptFingerprint(
	next map[string]string,
	current map[string]string,
) bool {
	for key, value := range next {
		if key == "Fingerprint" {
			continue
		}
		if current[key] != value {
			return false
		}
	}
	for key, value := range current {
		if key == "Fingerprint" {
			continue
		}
		if next[key] != value {
			return false
		}
	}
	return true
}

func preserveHiddenPresetPasswords(
	presets []configuration.Preset,
	current []configuration.Preset,
) []configuration.Preset {
	currentByID := make(map[string]configuration.Preset, len(current))
	for _, preset := range current {
		currentByID[preset.ID] = preset
	}
	merged := make([]configuration.Preset, len(presets))
	for i, preset := range presets {
		merged[i] = preset
		currentPreset, ok := currentByID[preset.ID]
		if !ok {
			continue
		}
		if preset.Meta["Authentication"] != "Password" ||
			currentPreset.Meta["Authentication"] != "Password" {
			continue
		}
		if hasPresetPasswordMeta(preset.Meta) {
			continue
		}
		merged[i].Meta = copyPresetMeta(preset.Meta)
		copyHiddenPresetPassword(merged[i].Meta, currentPreset.Meta)
	}
	return merged
}

func hasPresetPasswordMeta(meta map[string]string) bool {
	if meta == nil {
		return false
	}
	if _, ok := meta[configuration.PresetMetaPassword]; ok {
		return true
	}
	if _, ok := meta[configuration.PresetMetaEncryptedPassword]; ok {
		return true
	}
	return false
}

func copyPresetMeta(meta map[string]string) map[string]string {
	copied := make(map[string]string, len(meta))
	for key, value := range meta {
		copied[key] = value
	}
	return copied
}

func copyHiddenPresetPassword(
	target map[string]string,
	source map[string]string,
) {
	if source == nil {
		return
	}
	if value, ok := source[configuration.PresetMetaPassword]; ok {
		target[configuration.PresetMetaPassword] = value
	}
	if value, ok := source[configuration.PresetMetaEncryptedPassword]; ok {
		target[configuration.PresetMetaEncryptedPassword] = value
	}
}

// requireAuth applies the same shared-key verification used by socket verify.
func (p presetConfig) requireAuth(r *http.Request) error {
	if p.commonCfg.SharedKey == "" {
		return nil
	}
	key := r.Header.Get("X-Key")
	if len(key) <= 0 || len(key) > 64 {
		return ErrSocketInvalidAuthKey
	}
	time.Sleep(500 * time.Millisecond)
	decodedKey, decodedKeyErr := base64.StdEncoding.DecodeString(key)
	if decodedKeyErr != nil {
		return NewError(http.StatusBadRequest, decodedKeyErr.Error())
	}
	verifier := newSocketVerification(
		socket{commonCfg: p.commonCfg},
		configuration.Server{},
		p.commonCfg,
	)
	if !hmac.Equal(verifier.authKey(r), decodedKey) {
		return ErrSocketAuthFailed
	}
	return nil
}

func (p presetConfig) requirePresetAdminAuth(r *http.Request) error {
	if p.commonCfg.PresetAdminKey == "" {
		return NewError(
			http.StatusForbidden,
			"Full preset updates require PresetAdminKey authentication",
		)
	}
	key := r.Header.Get(presetAdminKeyHeader)
	if len(key) <= 0 || len(key) > 64 {
		return ErrSocketInvalidAuthKey
	}
	time.Sleep(500 * time.Millisecond)
	decodedKey, decodedKeyErr := base64.StdEncoding.DecodeString(key)
	if decodedKeyErr != nil {
		return NewError(http.StatusBadRequest, decodedKeyErr.Error())
	}
	if !hmac.Equal(presetAdminAuthKey(p.commonCfg.PresetAdminKey), decodedKey) {
		return ErrSocketAuthFailed
	}
	return nil
}

func presetAdminAuthKey(key string) []byte {
	timeMixer := strconv.FormatInt(time.Now().Unix()/100, 10)
	return hashCombineSocketKeys(timeMixer, key)[:32]
}

// writePresets serializes presets using the same preset shape sent to the UI.
func (p presetConfig) writePresets(
	w *ResponseWriter,
	presets []configuration.Preset,
) error {
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-store")
	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	return json.NewEncoder(w).Encode(presetConfigResponse{
		Presets: newSocketAccessConfiguration(presets, "", false).Presets,
	})
}
