package gerbera

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestTag(t *testing.T) {
	list := []struct {
		name string
		want *Element
		err  bool
	}{
		{name: "body", want: &Element{}, err: false},
		{name: "p", want: &Element{}, err: false},
		{name: "", want: &Element{}, err: true},
	}
	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			parent := &Element{TagName: "div"}
			err := Tag(item.name)(parent)
			if err != nil {
				t.Error(err)
			}
			for _, c := range parent.Children {
				if c.TagName != item.name {
					if item.err {
						continue
					}
					t.Error("TagNameが異なります")
				}
				if len(c.Children) != 0 {
					if item.err {
						continue
					}
					t.Error("要素が空であるはず")
				}
			}
		})
	}
}

func TestSkip(t *testing.T) {
	p := &Element{TagName: "div"}
	if err := Skip()(p); err != nil {
		t.Error(err.Error())
	}
	if len(p.Children) != 0 {
		t.Errorf("子要素が追加されている: want = %d, result = %d\n", 0, len(p.Children))
	}
}

func TestExecuteTemplate(t *testing.T) {
	buf := &bytes.Buffer{}
	if err := ExecuteTemplate(buf, "en"); err != nil {
		t.Error(err)
	}
	res := buf.String()
	buf.Reset()
	resArr := strings.Split(res, "\n")
	if len(resArr) != 2 {
		t.Errorf("出力されたHTMLの行数が異なります: want = 2, result = %d", len(resArr))
	}
	if resArr[0] != "<!DOCTYPE html>" {
		t.Errorf("DOCTYPE宣言が記述されていない: want = <!DOCTYPE html>, result = %s\n", resArr[0])
	}
	if resArr[1] != "<html lang=\"en\"></html>" {
		t.Errorf("html要素が異なります: want = <html lang=\"en\"></html>, result = %s", resArr[1])
	}
}

func TestIsEmptyElement(t *testing.T) {
	cases := []struct {
		tagName string
		isEmpty bool
		err     bool
	}{
		{tagName: "div", isEmpty: false, err: false},
	}
	for _, c := range cases {
		is := isEmptyElement(c.tagName)
		if c.isEmpty {
			if !is && !c.err {
				t.Errorf("")
			}
		} else {
			if is && !c.err {
				t.Errorf("")
			}
		}
	}
}

func TestWriteIndent(t *testing.T) {
	cases := []struct {
		count int
		str   string
	}{
		{count: 0, str: ""},
		{count: 1, str: " "},
		{count: 4, str: "    "},
		{count: 10, str: "          "},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		bw := bufio.NewWriter(&buf)
		writeIndent(bw, c.count)
		bw.Flush()
		if buf.String() != c.str {
			t.Errorf("writeIndent(%d): want = %q, result = %q\n", c.count, c.str, buf.String())
		}
	}
}
