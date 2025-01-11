package collection

import (
	"fmt"
	"strings"
	"time"

	"github.com/hasura/ndc-sdk-go/schema"
	"github.com/hasura/ndc-sdk-go/utils"
)

func normalizeObjectName(objectName string) string {
	// replace Unix-compatible backslashes in the file path when run on Windows OS
	return strings.ReplaceAll(objectName, "\\", "/")
}

func getComparisonValue(input schema.ComparisonValue, variables map[string]any) (any, error) {
	if len(input) == 0 {
		return nil, nil
	}

	switch v := input.Interface().(type) {
	case *schema.ComparisonValueScalar:
		return v.Value, nil
	case *schema.ComparisonValueVariable:
		if value, ok := variables[v.Name]; ok {
			return value, nil
		}

		return nil, fmt.Errorf("variable %s does not exist", v.Name)
	default:
		return nil, fmt.Errorf("invalid comparison value: %v", input)
	}
}

func getComparisonValueString(input schema.ComparisonValue, variables map[string]any) (*string, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableString(rawValue)
}

func getComparisonValueBoolean(input schema.ComparisonValue, variables map[string]any) (*bool, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableBoolean(rawValue)
}

func getComparisonValueDateTime(input schema.ComparisonValue, variables map[string]any) (*time.Time, error) {
	rawValue, err := getComparisonValue(input, variables)
	if err != nil {
		return nil, err
	}

	return utils.DecodeNullableDateTime(rawValue)
}

func compareNullableString(a, b *string) int {
	if a == nil && b == nil {
		return 0
	}

	if a != nil && b == nil {
		return 1
	}

	if a == nil && b != nil {
		return -1
	}

	return strings.Compare(*a, *b)
}

func compareNullableBoolean(a, b *bool) int {
	if a == nil && b == nil {
		return 0
	}

	if a != nil && b == nil {
		return 1
	}

	if a == nil && b != nil {
		return -1
	}

	if *a == *b {
		return 0
	}

	if *a && !*b {
		return 1
	}

	return -1
}

func compareNullableTime(a, b *time.Time) int {
	if a == nil && b == nil {
		return 0
	}

	if a != nil && b == nil {
		return 1
	}

	if a == nil && b != nil {
		return -1
	}

	return int(a.Sub(*b))
}
