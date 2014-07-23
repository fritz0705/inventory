package reichelt

import (
	"testing"
)

func TestSearchID(t *testing.T) {
	type testCase struct {
		Input string
		Output int64
	}

	// One test case is enough, more would be hard to maintain
	testCases := []testCase{
		{
			Input: "1/4W 4,7K",
			Output: 1425,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase)

		res, err := SearchID(testCase.Input)
		if err != nil {
			t.Error(err)
		} else if res != testCase.Output {
			t.Errorf("SearchID(%#v) = %#v, wants %#v", testCase.Input, res, testCase.Output)
		}
	}
}
