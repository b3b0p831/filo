// config/load.go
package config

import (
	"log"

	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	v := viper.New()

	// Set defaults.
	v.SetDefault("max_fill", 0.92) // The actions of this program cannot result in a change where new target size > max_fill
	v.SetDefault("log_level", "info")
	v.SetDefault("sync_delay", "10s")

	// Config file name and type
	v.SetConfigName("config") // without extension
	v.SetConfigType("toml")

	// Standard cross-platform config locations
	// if dir, err := os.UserConfigDir(); err == nil {
	// 	v.AddConfigPath(filepath.Join(dir, "mediasync"))
	// }
	v.AddConfigPath(".") // fallback: current dir

	// Read config
	if err := v.ReadInConfig(); err != nil {
		log.Default().Println("failed to read config:", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Default().Println("failed to read config:", err)

	}

	return &cfg, nil
}
