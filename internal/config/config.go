package config

import (
	"io"
	"log"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/lmittmann/tint"
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
}

func (cfg *Config) Equal(otherCFG Config) bool {

	return cfg.TargetDir == otherCFG.TargetDir && cfg.SourceDir == otherCFG.SourceDir &&
		cfg.MaxFill == otherCFG.MaxFill && cfg.SyncDelay == otherCFG.SyncDelay && slices.Equal(cfg.ApprovedExtensions, otherCFG.ApprovedExtensions)
}

var debugLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
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

	var outWriter io.Writer
	if cfg.LogFile != "" {
		outFile, err := os.OpenFile(cfg.LogFile, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err.Error(), outFile)
			outWriter = os.Stdout
		} else {
			outWriter = io.MultiWriter(os.Stdout, outFile)
		}
	} else {
		outWriter = os.Stdout
	}

	logger := slog.New(tint.NewHandler(outWriter, &tint.Options{
		Level: debugLevels[cfg.LogLevel],
	}))

	slog.SetDefault(logger)
	return &cfg
}
