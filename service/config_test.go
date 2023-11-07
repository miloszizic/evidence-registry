package service_test

import (
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miloszizic/der/service"
)

func TestUnmarshalJSONWasSuccessfully(t *testing.T) {
	t.Parallel()

	dat, err := os.ReadFile("testdata/.config.json")
	if err != nil {
		t.Errorf("failed to read test data: %v", err)
	}

	want := service.Config{
		Port:                3000,
		Env:                 "dev",
		SymmetricKey:        "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration: time.Hour,
		Database: service.PostgresConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "postgres",
		},
		Minio: service.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
	}

	var got service.Config

	err = got.UnmarshalJSON(dat)
	if err != nil {
		t.Errorf("failed to unmarshal test data: %v", err)
	}

	if !cmp.Equal(got, want) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestLoadProductionConfigWasSuccessful(t *testing.T) {
	t.Parallel()

	want := service.Config{
		Port:                 3000,
		Env:                  "test",
		SymmetricKey:         "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: time.Hour * 24 * 7,
		Database: service.PostgresConfig{
			Host:        "127.0.0.1",
			Port:        5432,
			User:        "postgres",
			Name:        "DER",
			Password:    "postgres",
			Timeout:     5,
			Automigrate: true,
		},
		Minio: service.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
	}
	got := service.LoadDefaultConfig()

	if !cmp.Equal(got, want) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestParsingFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		conf *service.ConfigFlag // config flag
	}{
		{
			name: "with test JSON data successful",
			args: []string{"-config", "testdata/.config.json"},
			conf: &service.ConfigFlag{
				Path: "testdata/.config.json",
			},
		},
		{
			name: "that are not specified returns default config successful",
			args: []string{},
			conf: &service.ConfigFlag{
				Path: "",
			},
		},
	}

	for _, tt := range tests {
		ptt := tt
		t.Run(ptt.name, func(t *testing.T) {
			t.Parallel()

			got, _, err := service.ParseFlags("", ptt.args)
			if err != nil {
				t.Errorf("ParseFlags error: %v", err)
			}
			if !cmp.Equal(got, ptt.conf) {
				t.Errorf(cmp.Diff(ptt.conf, got))
			}
		})
	}
}

func TestParsingFlagsWithHelp(t *testing.T) {
	t.Parallel()

	argument := []string{"-h"}

	_, _, err := service.ParseFlags("", argument)
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("expected help but got %v", err)
	}
}

func TestLoadConfigIfFileDoesNotExist(t *testing.T) {
	t.Parallel()

	config, err := service.LoadProductionConfig("testdata/.configuration.json")
	if err == nil && config != (service.Config{}) {
		t.Errorf("expected error, got nil")
	}
}

func TestLoadProductionConfigWithInvalidJSONFormatFailed(t *testing.T) {
	t.Parallel()

	config, err := service.LoadProductionConfig("testdata/.invalid_json_format.json")
	if err == nil && config != (service.Config{}) {
		t.Errorf("expected error, got nil")
	}
}
