package gerbera

import "testing"

type testCaseMap struct {
	m    Map
	key  string
	want interface{}
}

func TestMap_Get(t *testing.T) {
	cases := []testCaseMap{
		{
			m:    Map{"test": "result"},
			key:  "test",
			want: "result",
		},
		{
			m:    Map{"test": "result"},
			key:  "test1",
			want: "",
		},
	}
	for i, c := range cases {
		res := c.m.Get(c.key)
		if res != c.want {
			t.Errorf("#%d: result = %s, want = %v\n", i, res, c.want)
		}
	}
}
