// Copyright (C) 2019-2026 Ni Rui <ranqus@gmail.com>
// Copyright (C) 2026 Snuffy2
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	// PresetSecretKeyEnv is the environment variable containing the preset
	// encryption key as base64-encoded 32-byte data.
	PresetSecretKeyEnv = "SSHWIFTY_PRESET_SECRET_KEY"
	// PresetMetaPassword is the legacy plaintext preset password metadata key.
	PresetMetaPassword = "Password"
	// PresetMetaEncryptedPassword is the encrypted preset password metadata key.
	PresetMetaEncryptedPassword = "Encrypted Password"
)

const encryptedPasswordPrefix = "v1:aes-256-gcm:"

// ApplyPresetSecrets migrates and decrypts preset password metadata.
func ApplyPresetSecrets(presets []Preset) ([]Preset, bool, error) {
	key, hasKey, err := loadPresetSecretKey()
	if err != nil {
		return nil, false, err
	}

	processed := make([]Preset, len(presets))
	changed := false
	for i, preset := range presets {
		processed[i] = copyPreset(preset)
		plaintext, hasPlaintext := processed[i].Meta[PresetMetaPassword]
		encrypted, hasEncrypted := processed[i].Meta[PresetMetaEncryptedPassword]
		if !hasPlaintext && !hasEncrypted {
			continue
		}
		if processed[i].SecretMeta == nil {
			processed[i].SecretMeta = map[string]string{}
		}
		if hasPlaintext {
			processed[i].SecretMeta[PresetMetaPassword] = plaintext
			if !hasKey {
				continue
			}
			encryptedPassword, encryptErr := encryptPresetSecret(key, plaintext)
			if encryptErr != nil {
				return nil, false, encryptErr
			}
			processed[i].Meta[PresetMetaEncryptedPassword] = encryptedPassword
			delete(processed[i].Meta, PresetMetaPassword)
			changed = true
			continue
		}
		if !hasKey {
			return nil, false, fmt.Errorf(
				"%s is required to decrypt preset %q",
				PresetSecretKeyEnv,
				processed[i].Title,
			)
		}
		plaintextPassword, decryptErr := decryptPresetSecret(key, encrypted)
		if decryptErr != nil {
			return nil, false, decryptErr
		}
		processed[i].SecretMeta[PresetMetaPassword] = plaintextPassword
	}

	return processed, changed, nil
}

func copyPreset(preset Preset) Preset {
	copied := preset
	if preset.Meta != nil {
		copied.Meta = make(map[string]string, len(preset.Meta))
		for key, value := range preset.Meta {
			copied.Meta[key] = value
		}
	}
	if preset.SecretMeta != nil {
		copied.SecretMeta = make(map[string]string, len(preset.SecretMeta))
		for key, value := range preset.SecretMeta {
			copied.SecretMeta[key] = value
		}
	}
	return copied
}

func loadPresetSecretKey() ([]byte, bool, error) {
	value := strings.TrimSpace(GetEnv(PresetSecretKeyEnv))
	if value == "" {
		return nil, false, nil
	}
	key, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, false, fmt.Errorf("decode %s: %w", PresetSecretKeyEnv, err)
	}
	if len(key) != 32 {
		return nil, false, fmt.Errorf(
			"%s must decode to 32 bytes, got %d",
			PresetSecretKeyEnv,
			len(key),
		)
	}
	return key, true, nil
}

func encryptPresetSecret(key []byte, plaintext string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	return encryptedPasswordPrefix +
		base64.StdEncoding.EncodeToString(nonce) +
		":" +
		base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptPresetSecret(key []byte, encrypted string) (string, error) {
	if !strings.HasPrefix(encrypted, encryptedPasswordPrefix) {
		return "", errors.New("encrypted preset password has unsupported format")
	}
	parts := strings.Split(strings.TrimPrefix(encrypted, encryptedPasswordPrefix), ":")
	if len(parts) != 2 {
		return "", errors.New("encrypted preset password has invalid format")
	}
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("decode preset password nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode preset password ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt preset password: %w", err)
	}
	return string(plaintext), nil
}
