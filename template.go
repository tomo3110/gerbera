package gerbera

import (
	"bytes"
	"fmt"
	"io"
)

type Template struct {
	Lang string
	el   *Element
	buf  *bytes.Buffer
}

func (t *Template) init() {
	if t.el == nil {
		t.el = &Element{}
	}
	t.el.TagName = "html"
	t.el.Index = 0
	t.el.Value = ""
	t.el.Attr = make(AttrMap)
	if len(t.Lang) > 0 {
		t.el.Attr["lang"] = t.Lang
	} else {
		t.el.Attr["lang"] = "en"
	}
	t.el.ClassNames = make(ClassMap)
	t.el.Children = make([]*Element, 0)
	t.buf = &bytes.Buffer{}
}

func (t *Template) Mount(fn ...ComponentFunc) (err error) {
	t.init()
	t.buf.Reset()
	if t.el, err = Parse(t.el, fn...); err != nil {
		return err
	}
	if err := Render(t.buf, t.el); err != nil {
		return err
	}
	return nil
}

func (t *Template) String() string {
	return t.buf.String()
}

func (t *Template) Bytes() []byte {
	return t.buf.Bytes()
}

func (t *Template) Read(b []byte) (int, error) {
	if t.buf.Len() > 0 {
		return t.buf.Read(b)
	}
	if &t.el == nil {
		t.init()
	}
	if err := Render(t.buf, t.el); err != nil {
		return 0, err
	}
	return t.buf.Read(b)
}

func (t *Template) Execute(w io.Writer, c ...ComponentFunc) error {
	if err := t.Mount(c...); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, t); err != nil {
		return err
	}
	fmt.Println(t.el)
	t.el = nil
	return nil
}
