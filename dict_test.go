package fate

import "testing"

func TestDict(t *testing.T) {
	var tests = []struct {
		strs     []string
		q        string
		expected token
	}{
		{[]string{}, "foo", 0},
		{[]string{"foo"}, "foo", 0},
		{[]string{"foo"}, "bar", 1},
	}

	for ti, tt := range tests {
		d := newDict()
		for _, str := range tt.strs {
			d.ID(str)
		}

		res := d.ID(tt.q)
		if res != tt.expected {
			t.Errorf("[%d] Id(%q) => %d, want %d", ti, tt.q, res, tt.expected)
		}
	}
}
