package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP      HTTPConfig      `yaml:"http"`
	Reindexer ReindexerConfig `yaml:"reindexer"`
	Cache     CacheConfig     `yaml:"cache"`
	Redis     RedisConfig     `yaml:"redis"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Prefix   string `yaml:"prefix"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type ReindexerConfig struct {
	DSN       string `yaml:"dsn"`
	Namespace string `yaml:"namespace"`
}

type CacheConfig struct {
	TTL time.Duration `yaml:"ttl"`
}

type InformationConfig struct {
	HTTPAddr     string
	ReindexerDSN string
}

func Load() (Config, InformationConfig, error) {

	_ = godotenv.Load()

	cfg := Config{
		HTTP: HTTPConfig{
			Addr: ":8080",
		},
		Reindexer: ReindexerConfig{
			DSN:       "cproto://127.0.0.1:6534/db",
			Namespace: "documents",
		},
		Cache: CacheConfig{
			TTL: 15 * time.Minute,
		},
		Redis: RedisConfig{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
			Prefix:   "doc:",
		},
	}

	infConfig := InformationConfig{HTTPAddr: "default HTTP_ADDR", ReindexerDSN: "default HTTP_ADDR"}

	yamlPath := getenv("CONFIGYAML_PATH", "config.yaml")
	if fileExists(yamlPath) {
		if err := readYAML(yamlPath, &cfg); err != nil {
			return Config{}, infConfig, err
		}
	}
	infConfig.HTTPAddr = "Yaml HTTP_ADDR"
	infConfig.ReindexerDSN = "Yaml REINDEXER_DSN"

	if v, ok := os.LookupEnv("HTTP_ADDR"); ok && v != "" {
		cfg.HTTP.Addr = v
		infConfig.HTTPAddr = "env HTTP_ADDR"
	}
	if v, ok := os.LookupEnv("REINDEXER_DSN"); ok && v != "" {
		cfg.Reindexer.DSN = v
		infConfig.ReindexerDSN = "env REINDEXER_DSN"
	}
	if v, ok := os.LookupEnv("REINDEXER_NAMESPACE"); ok && v != "" {
		cfg.Reindexer.Namespace = v
	}
	if v, ok := os.LookupEnv("CACHE_TTL"); ok && v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, infConfig, errors.New("Не правильно задано время ttl кеша")
		}
		cfg.Cache.TTL = d
	}

	if v, ok := os.LookupEnv("REDIS_ADDR"); ok && v != "" {
		cfg.Redis.Addr = v
	}
	if v, ok := os.LookupEnv("REDIS_PASSWORD"); ok {
		cfg.Redis.Password = v
	}
	if v, ok := os.LookupEnv("REDIS_DB"); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, infConfig, errors.New("REDIS_DB должен быть числом (пример: 0)")
		}
		if n < 0 {
			return Config{}, infConfig, errors.New("REDIS_DB не может быть < 0")
		}
		cfg.Redis.DB = n
	}
	if v, ok := os.LookupEnv("REDIS_PREFIX"); ok && v != "" {
		cfg.Redis.Prefix = v
	}

	if err := validate(cfg); err != nil {
		return Config{}, infConfig, err
	}

	return cfg, infConfig, nil
}

func readYAML(path string, cfg *Config) error {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, cfg)
}

func validate(cfg Config) error {
	if cfg.HTTP.Addr == "" {
		return errors.New("HTTP addr пустой")
	}
	if cfg.Reindexer.DSN == "" {
		return errors.New("REINDEXER DSN пустой")
	}
	if cfg.Reindexer.Namespace == "" {
		return errors.New("REINDEXER пустой")
	}
	if cfg.Cache.TTL <= 0 {
		return errors.New("CACHE_TTL не больше 0")
	}
	return nil
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
