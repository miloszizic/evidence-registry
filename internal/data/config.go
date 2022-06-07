package data

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"
)

type Config struct {
	Port                int            `json:"port"`
	Env                 string         `json:"env"`
	SymmetricKey        string         `json:"symmetric"`
	AccessTokenDuration time.Duration  `json:"duration"`
	Database            PostgresConfig `json:"database"`
	Minio               MinioConfig    `json:"minio"`
}

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type MinioConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access"`
	SecretKey string `json:"secret"`
}

func (p *PostgresConfig) ConnectionInfo() string {
	if p.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Name)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Name)
}

// UnmarshalJSON is a custom unmarshaller for the Config struct
// that allows us to unmarshal the json into the Config struct
// with Time.Duration values that are not supported by the default
func (c *Config) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Port                int            `json:"port"`
		Env                 string         `json:"env"`
		SymmetricKey        string         `json:"symmetric"`
		AccessTokenDuration string         `json:"duration"`
		Database            PostgresConfig `json:"database"`
		Minio               MinioConfig    `json:"minio"`
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

func LoadProductionConfig(path string) (Config, error) {
	if path == "" {
		return TestAppConfig(), nil
	}
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var c Config
	// unmarshal the json into the config object
	err = c.UnmarshalJSON(f)
	if err != nil {
		return Config{}, err
	}
	fmt.Println("Successfully loaded config file")
	return c, nil
}

func TestPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host: "localhost",
		Port: 5432,
		User: "postgres",
		Name: "postgres",
	}
}

func TestMinioConfig() MinioConfig {
	return MinioConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
	}
}
func TestAppConfig() Config {
	return Config{
		Port:                3000,
		Env:                 "dev",
		SymmetricKey:        "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration: time.Hour,
		Database:            TestPostgresConfig(),
		Minio:               TestMinioConfig(),
	}
}

type ConfigFlag struct {
	Path string `json:"path"`
}

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
