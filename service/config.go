package service

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	defaultPostgresHost        = "127.0.0.1"
	defaultPostgresPort        = 5432
	defaultPostgresUser        = "postgres"
	defaultPostgresPass        = "postgres"
	defaultPostgresName        = "DER"
	defaultPostgresTimeout     = 5
	defaultPostgresAutomigrate = true

	defaultMinioEndpoint  = "localhost:9000"
	defaultMinioAccessKey = "minioadmin"
	defaultMinioSecretKey = "minioadmin"

	defaultAppPort              = 3000
	defaultAppEnv               = "test"
	defaultAppSymmetricKey      = "nigkjtvbrhugwpgaqbemmvnqbtywfrcq"
	defaultAccessTokenDuration  = time.Hour
	defaultRefreshTokenDuration = time.Hour * 24 * 7
)

// Config holds the application configuration settings.
type Config struct {
	Port                 int            `json:"port"`
	Env                  string         `json:"env"`
	SymmetricKey         string         `json:"symmetric"`
	AccessTokenDuration  time.Duration  `json:"duration"`
	RefreshTokenDuration time.Duration  `json:"refresh"`
	Database             PostgresConfig `json:"db"`
	Minio                MinioConfig    `json:"minio"`
}

// PostgresConfig holds the configuration settings for the postgres database.
type PostgresConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Timeout     int    `json:"timeout"`
	Automigrate bool   `json:"automigrate"`
}

// MinioConfig holds the configuration settings for the minio database.
type MinioConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access"`
	SecretKey string `json:"secret"`
}

// ConnectionInfo returns the connection string for the postgres database.
func (p *PostgresConfig) ConnectionInfo() string {
	if p.Password == "" {
		return fmt.Sprintf("postgres://%s:%d/%s?sslmode=disable&connect_timeout=%d", p.Host, p.Port, p.Name, p.Timeout)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&connect_timeout=%d", p.User, p.Password, p.Host, p.Port, p.Name, p.Timeout)
}

// UnmarshalJSON unmarshal the json data into the config struct.
func (c *Config) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Port                 int            `json:"port"`
		Env                  string         `json:"env"`
		SymmetricKey         string         `json:"symmetric"`
		AccessTokenDuration  string         `json:"duration"`
		RefreshTokenDuration string         `json:"refresh"`
		Database             PostgresConfig `json:"db"`
		Minio                MinioConfig    `json:"minio"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	duration, err := time.ParseDuration(tmp.AccessTokenDuration)
	if err != nil {
		return err
	}

	*c = Config{
		Port:                tmp.Port,
		Env:                 tmp.Env,
		SymmetricKey:        tmp.SymmetricKey,
		AccessTokenDuration: duration,
		Database:            tmp.Database,
		Minio:               tmp.Minio,
	}

	return nil
}

// LoadProductionConfig loads production config
func LoadProductionConfig(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var c Config

	err = c.UnmarshalJSON(f)
	if err != nil {
		return Config{}, err
	}

	fmt.Println("Successfully loaded config file")

	return c, nil
}

// LoadDefaultConfig loads default config
func LoadDefaultConfig() Config {
	return TestAppConfig()
}

// TestPostgresConfig returns a postgres config for testing.
func TestPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:        defaultPostgresHost,
		Port:        defaultPostgresPort,
		User:        defaultPostgresUser,
		Password:    defaultPostgresPass,
		Name:        defaultPostgresName,
		Timeout:     defaultPostgresTimeout,
		Automigrate: defaultPostgresAutomigrate,
	}
}

// TestMinioConfig returns a minio config for testing.
func TestMinioConfig() MinioConfig {
	return MinioConfig{
		Endpoint:  defaultMinioEndpoint,
		AccessKey: defaultMinioAccessKey,
		SecretKey: defaultMinioSecretKey,
	}
}

// TestAppConfig returns a config for testing.
func TestAppConfig() Config {
	return Config{
		Port:                 defaultAppPort,
		Env:                  defaultAppEnv,
		SymmetricKey:         defaultAppSymmetricKey,
		AccessTokenDuration:  defaultAccessTokenDuration,
		RefreshTokenDuration: defaultRefreshTokenDuration,
		Database:             TestPostgresConfig(),
		Minio:                TestMinioConfig(),
	}
}

// ConfigFlag holds the path to the config file.
type ConfigFlag struct {
	Path string `json:"path"`
}

// ParseFlags parses the flags passed to the application.
func ParseFlags(programme string, args []string) (config *ConfigFlag, output string, err error) {
	flags := flag.NewFlagSet(programme, flag.ContinueOnError)

	var buf bytes.Buffer

	flags.SetOutput(&buf)

	var conf ConfigFlag

	flags.StringVar(&conf.Path, "config", "", "Path to config file")

	err = flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	return &conf, buf.String(), nil
}
