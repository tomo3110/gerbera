package gerbera

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any { return &bytes.Buffer{} },
}

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
	t.el.Value = ""
	lang := t.Lang
	if lang == "" {
		lang = "en"
	}
	t.el.Attr = AttrMap{"lang": lang}
	t.el.ClassNames = nil
	t.el.Children = nil
	if t.buf == nil {
		t.buf = bufPool.Get().(*bytes.Buffer)
	}
	t.buf.Reset()
}

func (t *Template) Mount(fn ...ComponentFunc) error {
	t.init()
	t.el = Parse(t.el, fn...)
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
	if t.el == nil {
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
	bufPool.Put(t.buf)
	t.buf = nil
	t.el = nil
	return nil
}
