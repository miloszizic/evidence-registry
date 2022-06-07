package data_test

import (
	"flag"
	"github.com/google/go-cmp/cmp"
	"github.com/miloszizic/der/internal/data"
	"io/ioutil"
	"testing"
	"time"
)

func TestUnmarshalJSONWasSuccessfully(t *testing.T) {
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
func TestLoadProductionConfigWasSuccessful(t *testing.T) {
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
	got, err := data.LoadProductionConfig("")
	if err != nil {
		t.Errorf("failed to load production config: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf(cmp.Diff(want, got))
	}
}
func TestParsingFlags(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf *data.ConfigFlag // config flag
	}{
		{
			name: "with test JSON data successful",
			args: []string{"-config", "testdata/.config.json"},
			conf: &data.ConfigFlag{
				Path: "testdata/.config.json",
			},
		},
		{
			name: "that are not specified returns default config successful",
			args: []string{},
			conf: &data.ConfigFlag{
				Path: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := data.ParseFlags("", tt.args)
			if err != nil {
				t.Errorf("ParseFlags error: %v", err)
			}
			if !cmp.Equal(got, tt.conf) {
				t.Errorf(cmp.Diff(tt.conf, got))
			}
		})
	}
}
func TestParsingFlagsWithHelp(t *testing.T) {
	argument := []string{"-h"}
	_, _, err := data.ParseFlags("", argument)
	if err != flag.ErrHelp {
		t.Errorf("expected help but got %v", err)
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
