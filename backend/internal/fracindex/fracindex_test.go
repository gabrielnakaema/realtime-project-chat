package fracindex

import "testing"

func TestGenerateKeyBetween(t *testing.T) {
	cases := []struct {
		name  string
		left  string
		right string
	}{
		{name: "default", left: "", right: ""},
		{name: "before first", left: "", right: "500000000000"},
		{name: "after last", left: "500000000000", right: ""},
		{name: "between adjacent", left: "500000000000", right: "500000000001"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenerateKeyBetween(tc.left, tc.right)
			if err != nil {
				t.Fatalf("GenerateKeyBetween returned error: %v", err)
			}
			if tc.left != "" && !(got > tc.left) {
				t.Fatalf("expected %q to be greater than %q", got, tc.left)
			}
			if tc.right != "" && !(got < tc.right) {
				t.Fatalf("expected %q to be less than %q", got, tc.right)
			}
		})
	}
}

func TestGenerateKeyBetweenRepeatedInsertions(t *testing.T) {
	left := "500000000000"
	right := "500000000001"

	for range 32 {
		got, err := GenerateKeyBetween(left, right)
		if err != nil {
			t.Fatalf("GenerateKeyBetween returned error: %v", err)
		}
		if !(got > left && got < right) {
			t.Fatalf("expected %q to be between %q and %q", got, left, right)
		}
		right = got
	}
}
