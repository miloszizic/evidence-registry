package service_test

import (
	"testing"

	"github.com/miloszizic/der/service"
)

func TestGenerateCaseNameForMinioSuccessfully(t *testing.T) {
	t.Parallel()

	courtShortName := "ASCG"
	caseTypeName := "KM"
	caseNumber := int32(2)
	caseYear := int32(2023)

	expectedCaseName := "ascg-km-2-23"

	actualCaseName, err := service.GenerateCaseNameForMinio(courtShortName, caseTypeName, caseNumber, caseYear)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if actualCaseName != expectedCaseName {
		t.Errorf("Expected: %v, Got: %v", expectedCaseName, actualCaseName)
	}
}

func TestGenerateCaseNameForDBSuccessfully(t *testing.T) {
	t.Parallel()

	courtShortName := "ASCG"
	caseTypeName := "KM"
	caseNumber := int32(2)
	caseYear := int32(2023)

	expectedCaseName := "ASCG KM 2/23"

	actualCaseName, err := service.GenerateCaseNameForDB(courtShortName, caseTypeName, caseNumber, caseYear)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if actualCaseName != expectedCaseName {
		t.Errorf("Expected: %v, Got: %v", expectedCaseName, actualCaseName)
	}
}

func TestGenerateCaseNameForMinioFailedFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc            string
		courtShortName  string
		caseTypeName    string
		caseNumber      int32
		caseYear        int32
		expectedMessage string
	}{
		{
			desc:            "empty court short name",
			courtShortName:  "",
			caseTypeName:    "KM",
			caseNumber:      2,
			caseYear:        2023,
			expectedMessage: "court short name must not be empty",
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.desc, func(t *testing.T) {
			t.Parallel()
			_, err := service.GenerateCaseNameForMinio(pt.courtShortName, pt.caseTypeName, pt.caseNumber, pt.caseYear)
			if err == nil || err.Error() != pt.expectedMessage {
				t.Errorf("Expected error message: %v, got: %v", pt.expectedMessage, err)
			}
		})
	}
}

func TestGenerateCaseNameForDBFailedFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc            string
		courtShortName  string
		caseTypeName    string
		caseNumber      int32
		caseYear        int32
		expectedMessage string
	}{
		{
			desc:            "empty court short name",
			courtShortName:  "",
			caseTypeName:    "KM",
			caseNumber:      2,
			caseYear:        2023,
			expectedMessage: "court short name must not be empty",
		},
		{
			desc:            "empty case type name",
			courtShortName:  "ASCG",
			caseTypeName:    "",
			caseNumber:      2,
			caseYear:        2023,
			expectedMessage: "case type name must not be empty",
		},
		{
			desc:            "negative case number",
			courtShortName:  "ASCG",
			caseTypeName:    "KM",
			caseNumber:      -1,
			caseYear:        2023,
			expectedMessage: "case number must not be negative",
		},
		{
			desc:            "invalid case year",
			courtShortName:  "ASCG",
			caseTypeName:    "KM",
			caseNumber:      2,
			caseYear:        999,
			expectedMessage: "case year must be a valid year (not less than 1000)",
		},
	}

	for _, tt := range tests {
		pt := tt
		t.Run(pt.desc, func(t *testing.T) {
			t.Parallel()
			_, err := service.GenerateCaseNameForDB(pt.courtShortName, pt.caseTypeName, pt.caseNumber, pt.caseYear)
			if err == nil || err.Error() != pt.expectedMessage {
				t.Errorf("Expected error message: %v, got: %v", pt.expectedMessage, err)
			}
		})
	}
}
