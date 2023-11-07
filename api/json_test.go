package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		KnownField string `json:"knownField"`
	}

	type testStructWithInt struct {
		Field int `json:"field"`
	}

	tests := []struct {
		name                  string
		input                 string
		disallowUnknownFields bool
		wantErr               bool
		dst                   interface{}
	}{
		{
			name:  "Valid JSON",
			input: `{"key":"value"}`,
			dst:   &map[string]interface{}{},
		},
		{
			name:    "Exceeding max bytes",
			input:   strings.Repeat("a", 1_048_577), // 1 more than max
			dst:     &map[string]interface{}{},
			wantErr: true,
		},
		{
			name:                  "Unknown fields",
			input:                 `{"knownField":"value","unknownField":"unknown"}`,
			disallowUnknownFields: true,
			wantErr:               true,
			dst:                   &testStruct{},
		},
		{
			name:    "Badly-formed JSON",
			input:   `{"key":"value"`,
			dst:     &map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "Incorrect JSON type for field",
			input:   `{"field":"string"}`, // expecting an int, but got a string
			dst:     &testStructWithInt{},
			wantErr: true,
		},
		{
			name:    "Multiple JSON values",
			input:   `{"key":"value"}{"key":"value"}`,
			dst:     &map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "Empty JSON",
			input:   ``,
			dst:     &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(pt.input))
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}

			err = decodeJSON(w, r, pt.dst, pt.disallowUnknownFields)

			if (err != nil) != pt.wantErr {
				t.Errorf("decodeJSON() error = %v, wantErr %v", err, pt.wantErr)
			}
		})
	}
}
