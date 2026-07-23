package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                           string   `env:"PORT" envDefault:"8080"`
	ScyllaHosts                    []string `env:"SCYLLA_HOSTS,required"`
	ScyllaKeyspace                 string   `env:"SCYLLA_KEYSPACE" envDefault:"shortener"`
	ScyllaConsistency              string   `env:"SCYLLA_CONSISTENCY" envDefault:"ONE"`
	ScyllaDisableInitialHostLookup bool     `env:"SCYLLA_DISABLE_INITIAL_HOST_LOOKUP" envDefault:"false"`
	RedisAddr                      string   `env:"REDIS_ADDR,required"`
	BaseURL                        string   `env:"BASE_URL,required"`
	PingDatabaseFrequencyMillis    int      `env:"PING_DATABASE_FREQUENCY_IN_MILLIS" envDefault:"180000"`
	CacheDefaultTTLMillis          int      `env:"CACHE_DEFAULT_TTL_IN_MILLIS" envDefault:"3600000"`
}

func (c *Config) CacheDefaultTTL() time.Duration {
	return time.Duration(c.CacheDefaultTTLMillis) * time.Millisecond
}

func (c *Config) PingDatabaseFrequency() time.Duration {
	return time.Duration(c.PingDatabaseFrequencyMillis) * time.Millisecond
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
