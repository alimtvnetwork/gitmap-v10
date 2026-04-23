// Package config handles loading and merging configuration.
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// LoadFromFile reads a JSON config file and returns a Config.
// Returns default config if the file does not exist.
func LoadFromFile(path string) (model.Config, error) {
	cfg := model.DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {

		return cfg, handleMissingFile(err)
	}

	return parseConfig(data, cfg)
}

// handleMissingFile returns nil for missing files, error otherwise.
func handleMissingFile(err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return err
}

// parseConfig unmarshals JSON data into a Config struct.
func parseConfig(data []byte, cfg model.Config) (model.Config, error) {
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

// MergeWithFlags applies CLI flag overrides to the loaded config.
// Flags take precedence when they are non-empty.
func MergeWithFlags(cfg model.Config, mode, output, outputDir string) model.Config {
	cfg = applyMode(cfg, mode)
	cfg = applyOutput(cfg, output)
	cfg = applyOutputDir(cfg, outputDir)

	return cfg
}

// applyMode overrides the default mode if the flag is set.
func applyMode(cfg model.Config, mode string) model.Config {
	if len(mode) > 0 {
		cfg.DefaultMode = mode
	}

	return cfg
}

// applyOutput overrides the default output if the flag is set.
func applyOutput(cfg model.Config, output string) model.Config {
	if len(output) > 0 {
		cfg.DefaultOutput = output
	}

	return cfg
}

// applyOutputDir overrides the output directory if the flag is set.
func applyOutputDir(cfg model.Config, outputDir string) model.Config {
	if len(outputDir) > 0 {
		cfg.OutputDir = outputDir
	}

	return cfg
}
