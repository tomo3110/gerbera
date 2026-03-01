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

func TestSpan(t *testing.T) {
	testCommon(t, Span(), "span")
}

func TestDetails(t *testing.T) {
	testCommon(t, Details(), "details")
}

func TestSummary(t *testing.T) {
	testCommon(t, Summary(), "summary")
}

func TestStyle(t *testing.T) {
	testCommon(t, Style(), "style")
}

func TestDialog(t *testing.T) {
	testCommon(t, Dialog(), "dialog")
}

func TestProgress(t *testing.T) {
	testCommon(t, Progress(), "progress")
}

func TestMeter(t *testing.T) {
	testCommon(t, Meter(), "meter")
}

func TestTime(t *testing.T) {
	testCommon(t, Time(), "time")
}

func TestMark(t *testing.T) {
	testCommon(t, Mark(), "mark")
}

func TestAbbr(t *testing.T) {
	testCommon(t, Abbr(), "abbr")
}

func TestFigure(t *testing.T) {
	testCommon(t, Figure(), "figure")
}

func TestFigcaption(t *testing.T) {
	testCommon(t, Figcaption(), "figcaption")
}

func TestVideo(t *testing.T) {
	testCommon(t, Video(), "video")
}

func TestAudio(t *testing.T) {
	testCommon(t, Audio(), "audio")
}

func TestSource(t *testing.T) {
	testCommon(t, Source(), "source")
}

func TestCanvas(t *testing.T) {
	testCommon(t, Canvas(), "canvas")
}

func TestPicture(t *testing.T) {
	testCommon(t, Picture(), "picture")
}

func TestTmpl(t *testing.T) {
	testCommon(t, Tmpl(), "template")
}

func TestSlot(t *testing.T) {
	testCommon(t, Slot(), "slot")
}

func TestH6(t *testing.T) {
	testCommon(t, H6(), "h6")
}

func TestPre(t *testing.T) {
	testCommon(t, Pre(), "pre")
}

func TestBlockquote(t *testing.T) {
	testCommon(t, Blockquote(), "blockquote")
}

func TestEm(t *testing.T) {
	testCommon(t, Em(), "em")
}

func TestStrong(t *testing.T) {
	testCommon(t, Strong(), "strong")
}

func TestSmall(t *testing.T) {
	testCommon(t, Small(), "small")
}

func TestSub(t *testing.T) {
	testCommon(t, Sub(), "sub")
}

func TestSup(t *testing.T) {
	testCommon(t, Sup(), "sup")
}

func TestDl(t *testing.T) {
	testCommon(t, Dl(), "dl")
}

func TestDt(t *testing.T) {
	testCommon(t, Dt(), "dt")
}

func TestDd(t *testing.T) {
	testCommon(t, Dd(), "dd")
}

func TestFieldset(t *testing.T) {
	testCommon(t, Fieldset(), "fieldset")
}

func TestLegend(t *testing.T) {
	testCommon(t, Legend(), "legend")
}

func TestCaption(t *testing.T) {
	testCommon(t, Caption(), "caption")
}

func TestColgroup(t *testing.T) {
	testCommon(t, Colgroup(), "colgroup")
}

func TestCol(t *testing.T) {
	testCommon(t, Col(), "col")
}

func TestIframe(t *testing.T) {
	testCommon(t, Iframe(), "iframe")
}

func TestEmbedEl(t *testing.T) {
	testCommon(t, EmbedEl(), "embed")
}

func TestObject(t *testing.T) {
	testCommon(t, Object(), "object")
}

func TestParam(t *testing.T) {
	testCommon(t, Param(), "param")
}

func TestOutput(t *testing.T) {
	testCommon(t, Output(), "output")
}

func TestDatalist(t *testing.T) {
	testCommon(t, Datalist(), "datalist")
}

func TestOptgroup(t *testing.T) {
	testCommon(t, Optgroup(), "optgroup")
}

func TestRuby(t *testing.T) {
	testCommon(t, Ruby(), "ruby")
}

func TestRt(t *testing.T) {
	testCommon(t, Rt(), "rt")
}

func TestRp(t *testing.T) {
	testCommon(t, Rp(), "rp")
}

func TestWbr(t *testing.T) {
	testCommon(t, Wbr(), "wbr")
}

func TestMapEl(t *testing.T) {
	testCommon(t, MapEl(), "map")
}

func TestNoscript(t *testing.T) {
	testCommon(t, Noscript(), "noscript")
}
