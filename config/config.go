package config

// in-memory representation of the config.toml file.

type Config struct {
	SourceDir string  `mapstructure:"source_dir"`
	TargetDir string  `mapstructure:"target_dir"`
	MaxFill   float64 `mapstructure:"max_fill"`
	LogLevel  string  `mapstructure:"log_level"`
	SyncDelay string  `mapstructure:"sync_delay"`
}
