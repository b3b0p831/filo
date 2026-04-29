package config

import (
	"fmt"
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
	MaxOpenFile        int           `mapstructure:"max_openfile"`
}

func (cfg *Config) Equal(otherCFG Config) bool {

	return cfg.TargetDir == otherCFG.TargetDir && cfg.SourceDir == otherCFG.SourceDir &&
		cfg.MaxFill == otherCFG.MaxFill && cfg.SyncDelay == otherCFG.SyncDelay &&
		slices.Equal(cfg.ApprovedExtensions, otherCFG.ApprovedExtensions) && cfg.LogFile == otherCFG.LogFile &&
		cfg.MaxOpenFile == otherCFG.MaxOpenFile
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
	v.SetDefault("max_openfile", "100")

	// Config file name and type
	v.SetConfigName("filo") // without extension
	v.SetConfigType("toml")

	v.AddConfigPath("/etc/filo/") // fallback: current dir
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error(fmt.Sprintf("os.Getwd() failed in config.go: %s\n", err.Error()))
		os.Exit(2)
	}
	v.AddConfigPath(cwd) // fallback: current dir

	// Read config
	if err := v.ReadInConfig(); err != nil {
		slog.Error(fmt.Sprintf("viper.ReadInConfig() failed in config.go: %s\n", err.Error()))
		os.Exit(2)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error(fmt.Sprintf("viper.Unmarshal() failed in config.go: %s\n", err.Error()))
		os.Exit(2)
	}

	var outWriter io.Writer
	if cfg.LogFile != "" {
		outFile, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
		Level:      debugLevels[cfg.LogLevel],
		TimeFormat: time.DateTime,
	}))

	slog.SetDefault(logger)
	return &cfg
}
