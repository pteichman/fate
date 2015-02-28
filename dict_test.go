package fate

import (
	"reflect"
	"strings"
	"testing"
)

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

func TestSyndict(t *testing.T) {
	var tests = []struct {
		strs     []string
		query    string
		expected tokset
	}{
		{[]string{"foo", "Foo", "bar", "baz"}, "FOO", tokset{0, 1}},
	}

	for ti, tt := range tests {
		d := newSyndict(strings.ToLower)

		for _, str := range tt.strs {
			d.ID(str)
		}

		res := d.Syns(tt.query)
		if !reflect.DeepEqual(res, tt.expected) {
			t.Errorf("[%d] Get(%q) => %d, want %d", ti, tt.query, res, tt.expected)
		}
	}
}
