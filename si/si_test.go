package si

import (
	"testing"
)

func TestCanon(t *testing.T) {
	type testCase struct {
		Input Number
		Result Number
	}

	testCases := []testCase{
		testCase{
			Input: Number{ 33000, None },
			Result: Number{ 33, Kilo },
		},
		testCase{
			Input: Number{ 1000, Kilo },
			Result: Number{ 1, Mega },
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase)

		res := testCase.Input.Canon()
		if res != testCase.Result {
			t.Errorf("Canon(%s) should be %s, got %s", testCase.Input, testCase.Result, res)
		}
	}
}

func TestValue(t *testing.T) {
	type testCase struct {
		Input Number
		Result float64
	}

	testCases := []testCase{
		testCase{
			Input: Number{ 3.3, Nano },
			Result: 0.0000000033,
		},
		testCase{
			Input: Number{ 1, Kilo },
			Result: 1000,
		},
	}

	for _, testCase := range testCases {
		t.Log(testCase)

		res := testCase.Input.Value()
		if res != testCase.Result {
			t.Errorf("Value(%s) should be %s, got %s", testCase.Input, testCase.Result, res)
		}
	}
}
