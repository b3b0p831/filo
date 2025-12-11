package config

import (
	"log"
	"os"
	"slices"
	"time"

	"github.com/spf13/viper"
)

// in-memory representation of the config.toml file.

type Config struct {
	SourceDir          string        `mapstructure:"source_dir"`
	TargetDir          string        `mapstructure:"target_dir"`
	MaxFill            float64       `mapstructure:"max_fill"`
	LogLevel           string        `mapstructure:"log_level"`
	SyncDelay          time.Duration `mapstructure:"sync_delay"`
	ApprovedExtensions []string      `mapstructure:"approved_extensions"`
	LogFile            string        `mapstructure:"log_file"`
	Flogger            *log.Logger
}

func (cfg *Config) Equal(otherCFG Config) bool {

	return cfg.TargetDir == otherCFG.TargetDir && cfg.SourceDir == otherCFG.SourceDir &&
		cfg.MaxFill == otherCFG.MaxFill && cfg.SyncDelay == otherCFG.SyncDelay && slices.Equal(cfg.ApprovedExtensions, otherCFG.ApprovedExtensions)
}

func Load() *Config {
	v := viper.New()

	// Set defaults.
	v.SetDefault("max_fill", 0.92) // The actions of this program cannot result in a change where new target size > max_fill
	v.SetDefault("log_level", "info")
	v.SetDefault("sync_delay", "30s")

	// Config file name and type
	v.SetConfigName("config") // without extension
	v.SetConfigType("toml")

	v.AddConfigPath("/etc/filo/filo.conf") // fallback: current dir
	v.AddConfigPath(".")                   // fallback: current dir

	// Read config
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("failed to read config:", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatal("failed to read config:", err)
	}

	// if cfg.LogFile == "" {
	// }
	cfg.Flogger = log.New(os.Stdin, "", log.Ltime|log.Lshortfile)

	return &cfg
}
