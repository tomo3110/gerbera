package dom

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func testCommon(t *testing.T, c gerbera.ComponentFunc, wantTagName string) {
	parent := &gerbera.Element{TagName: "div"}
	if err := c(parent); err != nil {
		t.Error(err.Error())
	}
	target := parent.Children[0]
	if target == nil {
		t.Error("要素が正常に追加されていません")
	}
	if target.TagName != wantTagName {
		t.Errorf("要素名が異なります: want = %s, result = %s\n", wantTagName, target.TagName)
	}
}

func TestB(t *testing.T) {
	testCommon(t, B(), "b")
}

func TestA(t *testing.T) {
	testCommon(t, A(), "a")
}

func TestP(t *testing.T) {
	testCommon(t, P(), "p")
}

func TestBr(t *testing.T) {
	testCommon(t, Br(), "br")
}

func TestArticle(t *testing.T) {
	testCommon(t, Article(), "article")
}

func TestAddress(t *testing.T) {
	testCommon(t, Address(), "address")
}

func TestAside(t *testing.T) {
	testCommon(t, Aside(), "aside")
}

func TestArea(t *testing.T) {
	testCommon(t, Area(), "area")
}

func TestButton(t *testing.T) {
	testCommon(t, Button(), "button")
}

func TestBody(t *testing.T) {
	testCommon(t, Body(), "body")
}

func TestCode(t *testing.T) {
	testCommon(t, Code(), "code")
}

func TestDiv(t *testing.T) {
	testCommon(t, Div(), "div")
}

func TestFooter(t *testing.T) {
	testCommon(t, Footer(), "footer")
}

func TestForm(t *testing.T) {
	testCommon(t, Form(), "form")
}

func TestH1(t *testing.T) {
	testCommon(t, H1(), "h1")
}

func TestH2(t *testing.T) {
	testCommon(t, H2(), "h2")
}

func TestH3(t *testing.T) {
	testCommon(t, H3(), "h3")
}

func TestH4(t *testing.T) {
	testCommon(t, H4(), "h4")
}

func TestH5(t *testing.T) {
	testCommon(t, H5(), "h5")
}

func TestHr(t *testing.T) {
	testCommon(t, Hr(), "hr")
}

func TestHead(t *testing.T) {
	testCommon(t, Head(), "head")
}

func TestHeader(t *testing.T) {
	testCommon(t, Header(), "header")
}

func TestHTML(t *testing.T) {
	testCommon(t, HTML("ja"), "html")
}

func TestImg(t *testing.T) {
	testCommon(t, Img("./"), "img")
}

func TestInput(t *testing.T) {
	testCommon(t, Input(), "input")
}

func TestInputCheckbox(t *testing.T) {
	testCommon(t, InputCheckbox(true, 1), "input")
}

func TestLabel(t *testing.T) {
	testCommon(t, Label(), "label")
}

func TestLi(t *testing.T) {
	testCommon(t, Li(), "li")
}

func TestLink(t *testing.T) {
	testCommon(t, Link(), "link")
}

func TestMainElement(t *testing.T) {
	testCommon(t, Main(), "main")
}

func TestMeta(t *testing.T) {
	testCommon(t, Meta(), "meta")
}

func TestNav(t *testing.T) {
	testCommon(t, Nav(), "nav")
}

func TestOl(t *testing.T) {
	testCommon(t, Ol(), "ol")
}

func TestOption(t *testing.T) {
	testCommon(t, Option(), "option")
}

func TestScript(t *testing.T) {
	testCommon(t, Script("./"), "script")
}

func TestSection(t *testing.T) {
	testCommon(t, Section(), "section")
}

func TestSelect(t *testing.T) {
	testCommon(t, Select(), "select")
}

func TestSubmit(t *testing.T) {
	testCommon(t, Submit(), "button")
}

func TestTable(t *testing.T) {
	testCommon(t, Table(), "table")
}

func TestTbody(t *testing.T) {
	testCommon(t, Tbody(), "tbody")
}

func TestThead(t *testing.T) {
	testCommon(t, Thead(), "thead")
}

func TestTfoot(t *testing.T) {
	testCommon(t, Tfoot(), "tfoot")
}

func TestTd(t *testing.T) {
	testCommon(t, Td(), "td")
}

func TestTh(t *testing.T) {
	testCommon(t, Th(), "th")
}

func TestTr(t *testing.T) {
	testCommon(t, Tr(), "tr")
}

func TestTextarea(t *testing.T) {
	testCommon(t, Textarea(), "textarea")
}

func TestTitle(t *testing.T) {
	testCommon(t, Title(""), "title")
}

func TestUl(t *testing.T) {
	testCommon(t, Ul(), "ul")
}
