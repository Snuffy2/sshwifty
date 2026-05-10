// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package controller

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/Snuffy2/sshwifty/application/command"
	"github.com/Snuffy2/sshwifty/application/configuration"
	"github.com/Snuffy2/sshwifty/application/log"
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
	if err := p.requireAuth(r); err != nil {
		return err
	}
	if p.commonCfg.SourceFile == "" {
		return NewError(
			http.StatusConflict,
			"Preset updates require a file-backed configuration",
		)
	}

	var request presetConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}

	presets := make([]configuration.Preset, len(request.Presets))
	for i, preset := range request.Presets {
		presets[i] = configuration.Preset{
			ID:       preset.ID,
			Title:    preset.Title,
			Type:     preset.Type,
			Host:     preset.Host,
			TabColor: preset.TabColor,
			Meta:     preset.Meta,
		}
	}

	normalized, _, err := configuration.EnsurePresetIDs(presets)
	if err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	normalized, err = p.commands.Reconfigure(normalized)
	if err != nil {
		return NewError(http.StatusBadRequest, err.Error())
	}
	if err := configuration.ReplaceFilePresets(
		p.commonCfg.SourceFile,
		normalized,
	); err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	p.commonCfg.PresetRepository.Replace(normalized)
	return p.writePresets(w, normalized)
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

// writePresets serializes presets using the same preset shape sent to the UI.
func (p presetConfig) writePresets(
	w *ResponseWriter,
	presets []configuration.Preset,
) error {
	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-store")
	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	return json.NewEncoder(w).Encode(presetConfigResponse{
		Presets: newSocketAccessConfiguration(presets, "").Presets,
	})
}
