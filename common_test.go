package gerbera

import (
	"bytes"
	"reflect"
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
			err := Tag(item.name, []ComponentFunc{})(parent)
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
	if len(resArr) != 3 {
		t.Errorf("出力されたHTMLの行数が異なります: want = 3, result = %d", len(resArr))
	}
	if resArr[0] != "<!DOCTYPE html>" {
		t.Errorf("DOCTYPE宣言が記述されていない: want = <!DOCTYPE html>, result = %s\n", resArr[0])
	}
	if resArr[1] != "<html lang=\"en\">" {
		t.Errorf("html要素の開始タグが異なります: want = <html lang=\"en\">, result = %s", resArr[1])
	}
	if resArr[2] != "</html>" {
		t.Errorf("html要素の終了タグが異なります: want = </html>, result = %s", resArr[2])
	}
}

func TestCnvToSlice(t *testing.T) {
	s := CnvToSlice(Tag("h1", []ComponentFunc{}))
	typeInfo := reflect.TypeOf(s)
	if typeInfo.String() != "[]gerbera.ComponentFunc" {
		t.Errorf("型名が異なります: want = []gerbera.ComponentFunc, result = %s", typeInfo.String())
	}
	if len(s) != 1 {
		t.Errorf("子要素数が異なります")
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

func TestBytesRepeat(t *testing.T) {
	cases := []struct {
		b      string
		count  int
		str    string
		wcount int
	}{
		{count: 5, b: "b", str: "bbbbb", wcount: 5},
		{count: 7, b: "by", str: "bybybybybybyby", wcount: 14},
		{count: 3, b: "hoge", str: "hogehogehoge", wcount: 12},
		{count: 1, b: "bc", str: "bc", wcount: 2},
		{count: 0, b: "abc", str: "", wcount: 0},
	}
	buf := &bytes.Buffer{}
	for _, c := range cases {
		buf.Reset()
		bytesRepeat(buf, []byte(c.b), c.count)
		if buf.Len() != c.wcount {
			t.Errorf("byte列の%sを%d回繰り返す場合の文字数: want = %d, result = %d\n", c.b, c.wcount, c.count, buf.Len())
		}
		if buf.String() != c.str {
			t.Errorf("byte列の%sを%d回繰り返す場合の文字列: want = %s, result = %s\n", c.b, c.count, c.str, buf.String())
		}
	}
}
