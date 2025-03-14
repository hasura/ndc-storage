package encoding

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestDecodeArbitraryCSV(t *testing.T) {
	testCases := []struct {
		Name        string
		Input       string
		ContentType string
		Expected    [][]string
	}{
		{
			Name: "tsv",
			Input: `
id	title	date
1	movie 1	2000
2	movie 2	2001
3	movie 1 "part 2"	2003
4	movie 4	2004
`,
			ContentType: contentTypeTextTabSeparatedValues,
			Expected: [][]string{
				{"id", "title", "date"},
				{"1", "movie 1", "2000"},
				{"2", "movie 2", "2001"},
				{"3", `movie 1 "part 2"`, "2003"},
				{"4", "movie 4", "2004"},
			},
		},
		{
			Name:        "tsv2",
			Input:       "first_name\tlast_name\tusername\n\"Rob\"\t\"Pike\"\trob\nKen\tThompson\tken\n\"Robert\"\t\"Griesemer\"\t\"gri\"",
			ContentType: contentTypeTextTabSeparatedValues,
			Expected: [][]string{
				{"first_name", "last_name", "username"},
				{"Rob", "Pike", "rob"},
				{"Ken", "Thompson", "ken"},
				{"Robert", `Griesemer`, "gri"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := decodeArbitraryCSV(strings.NewReader(tc.Input), tc.ContentType)
			assert.NilError(t, err)
			assert.DeepEqual(t, tc.Expected, result)
		})
	}
}
