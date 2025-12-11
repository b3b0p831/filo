package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
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

var re regexp.Regexp = *regexp.MustCompile(`^\d+[smh]$`)

func (cfg *Config) Equal(otherCFG Config) bool {

	return cfg.TargetDir == otherCFG.TargetDir && cfg.SourceDir == otherCFG.SourceDir &&
		cfg.MaxFill == otherCFG.MaxFill && cfg.SyncDelay == otherCFG.SyncDelay && slices.Equal(cfg.ApprovedExtensions, otherCFG.ApprovedExtensions)
}

func GetTimeInterval(interval string) (time.Duration, error) {
	if !re.Match([]byte(interval)) {
		return 0, fmt.Errorf("interval string does not match format (i.e 1s, 3m, 5h): %v", interval)
	}

	timeValStr, lastChar := interval[:len(interval)-1], interval[len(interval)-1]
	timeVal, err := strconv.ParseInt(timeValStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("util/util.go: unable to ParseInt(timevalStr)")
	}

	switch lastChar {
	case 's':
		return time.Duration(timeVal) * time.Second, nil
	case 'm':
		return time.Duration(timeVal) * time.Minute, nil
	case 'h':
		return time.Duration(timeVal) * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid time unit: %v", lastChar)
	}

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
