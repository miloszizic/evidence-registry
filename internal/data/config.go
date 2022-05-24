package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

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

//func (c Config) IsProd() bool {
//	return c.Env == "prod"
//}

func LoadProductionConfig(prod bool) (Config, error) {
	if !prod {
		return TestAppConfig(), nil
	}
	f, err := ioutil.ReadFile(".config.json")
	if err != nil {
		return Config{}, err
	}
	var c Config
	// unmarshal the json into the config object
	err = c.UnmarshalJSON(f)
	if err != nil {
		return Config{}, err
	}
	fmt.Println("Successfully loaded .config.json")
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
