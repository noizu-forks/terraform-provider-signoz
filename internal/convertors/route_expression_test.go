package conv

import "testing"

func TestCanonicalRouteExpression(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "word AND becomes &&",
			in:   `severity == "critical" AND environment == "production"`,
			want: `severity == "critical" && environment == "production"`,
		},
		{
			name: "word OR becomes ||",
			in:   `severity == "critical" OR severity == "warning"`,
			want: `severity == "critical" || severity == "warning"`,
		},
		{
			name: "canonical operators untouched",
			in:   `severity == "critical" && environment == "production"`,
			want: `severity == "critical" && environment == "production"`,
		},
		{
			name: "quoted literals untouched",
			in:   `label == "AND" OR label == " OR "`,
			want: `label == "AND" || label == " OR "`,
		},
		{
			name: "identifier boundaries respected",
			in:   `ANDROID == "1" AND ORDER == "2"`,
			want: `ANDROID == "1" && ORDER == "2"`,
		},
		{
			name: "whitespace collapsed",
			in:   "severity==\"critical\"    AND\t\n environment==\"production\"  ",
			want: `severity=="critical" && environment=="production"`,
		},
		{
			name: "idempotent",
			in:   `a == "1" AND b == "2" AND c == "3"`,
			want: CanonicalRouteExpression(CanonicalRouteExpression(`a == "1" AND b == "2" AND c == "3"`)),
		},
		{
			name: "empty string",
			in:   ``,
			want: ``,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := CanonicalRouteExpression(tc.in); got != tc.want {
				t.Errorf("CanonicalRouteExpression(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
