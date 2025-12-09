package config

import "slices"

// in-memory representation of the config.toml file.

type Config struct {
	SourceDir          string   `mapstructure:"source_dir"`
	TargetDir          string   `mapstructure:"target_dir"`
	MaxFill            float64  `mapstructure:"max_fill"`
	LogLevel           string   `mapstructure:"log_level"`
	SyncDelay          string   `mapstructure:"sync_delay"`
	ApprovedExtensions []string `mapstructure:"approved_extensions"`
}

func (cfg Config) Equal(otherCFG Config) bool {

	return cfg.TargetDir == otherCFG.TargetDir && cfg.SourceDir == otherCFG.SourceDir &&
		cfg.MaxFill == otherCFG.MaxFill && cfg.SyncDelay == otherCFG.SyncDelay && slices.Equal(cfg.ApprovedExtensions, otherCFG.ApprovedExtensions)
}
