package data_test

import (
	"evidence/internal/data"
	"flag"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	dat, err := ioutil.ReadFile("testdata/.config.json")
	if err != nil {
		t.Errorf("failed to read test data: %v", err)
	}
	want := data.Config{
		Port:                3000,
		Env:                 "dev",
		SymmetricKey:        "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration: time.Hour,
		Database: data.PostgresConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "postgres",
		},
		Minio: data.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
	}
	var got data.Config

	err = got.UnmarshalJSON(dat)
	if err != nil {
		t.Errorf("failed to unmarshal test data: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestLoadProductionConfig(t *testing.T) {
	want := data.Config{
		Port:                3000,
		Env:                 "dev",
		SymmetricKey:        "nigkjtvbrhugwpgaqbemmvnqbtywfrcq",
		AccessTokenDuration: time.Hour,
		Database: data.PostgresConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "postgres",
		},
		Minio: data.MinioConfig{
			Endpoint:  "localhost:9000",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
		},
	}
	confFile := flag.String("config path", "", "Provide a path to a .config.json file. This file should be provided in production.")
	got, err := data.LoadProductionConfig(*confFile)
	if err != nil {
		t.Errorf("failed to load production config: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestLoadingFlags(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf *data.ConfigFlag // config flag
	}{
		{
			name: "test data",
			args: []string{"-config", "testdata/.config.json"},
			conf: &data.ConfigFlag{
				Path: "testdata/.config.json",
			},
		},
		{
			name: "no config",
			args: []string{},
			conf: &data.ConfigFlag{
				Path: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strings.Join(tt.args, " ")
			got, _, err := data.ParseFlags(tt.name, tt.args)
			if err != nil {
				t.Errorf("ParseFlags() error = %v", err)
			}
			if !cmp.Equal(got, tt.conf) {
				t.Errorf(cmp.Diff(tt.conf, got))
			}
		})
	}

}
func TestLoadConfigIfFileDoesNotExist(t *testing.T) {
	config, err := data.LoadProductionConfig("testdata/.configuration.json")
	if err == nil && config != (data.Config{}) {
		t.Errorf("expected error, got nil")
	}
}
func TestLoadProductionConfigWithInvalidJSONFormatFailed(t *testing.T) {
	config, err := data.LoadProductionConfig("testdata/.invalid_json_format.json")
	if err == nil && config != (data.Config{}) {
		t.Errorf("expected error, got nil")
	}
}