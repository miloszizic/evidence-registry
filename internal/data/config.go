package data

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

type Config struct {
	Port                int            `json:"port"`
	Env                 string         `json:"env"`
	Pepper              string         `json:"pepper"`
	SymmetricKey        string         `json:"symmetric"`
	AccessTokenDuration time.Duration  `json:"duration"`
	Database            PostgresConfig `json:"database"`
	Minio               MinioConfig    `json:"ostorage"`
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

//func (c Config) IsProd() bool {
//	return c.Env == "prod"
//}

func LoadProductionConfig(prod bool) (Config, error) {
	if !prod {
		return DefaultAppConfig(), nil
	}
	f, err := os.Open(".config.json")
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
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
func DefaultAppConfig() Config {
	return Config{
		Port:                3000,
		Env:                 "dev",
		SymmetricKey:        "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration: time.Minute * 15,
		Database:            TestPostgresConfig(),
		Minio:               TestMinioConfig(),
	}
}
