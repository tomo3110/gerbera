# アンケートフォーム チュートリアル

## 概要

ドロップダウン・チェックボックスを含むアンケートフォームを構築します。このチュートリアルでは以下の機能を学びます。

- `Select`, `Option` — ドロップダウンリスト
- `InputCheckbox` — チェックボックス入力
- `Input` — テキスト/メール入力
- `Submit` — 送信ボタン
- `Unless` — 否定条件の表示
- `If` — 条件付き表示
- `Name`, `ID` — フォーム属性
- `Ol`, `Li` — 番号付きリスト

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: データモデルとストレージ

```go
type SurveyResponse struct {
	Name      string
	Email     string
	Language  string
	Interests []string
}

func (s *SurveyResponse) ToMap() g.Map {
	return g.Map{
		"name":      s.Name,
		"email":     s.Email,
		"language":  s.Language,
		"interests": s.Interests,
	}
}

var (
	responses []*SurveyResponse
	mu        sync.Mutex
)
```

インメモリのスライスで回答を保持します。並行アクセスに備えて `sync.Mutex` を使います。

## ステップ 2: テキスト入力 — Input

```go
gd.Input(
	gp.Type("text"),
	gp.Class("form-control"),
	gp.ID("name"),
	gp.Name("name"),
	gp.Required(true),
),
```

- `gd.Input(...)` は `<input>` void要素を生成します
- `gp.ID("name")` は `id="name"` を設定
- `gp.Name("name")` は `name="name"` を設定（フォーム送信時のキー）
- `gp.Type("text")` で入力タイプを設定し、`gp.Required(true)` でフィールドを必須にします

## ステップ 3: ドロップダウン — Select / Option

```go
gd.Select(
	gp.Class("form-control"),
	gp.ID("language"),
	gp.Name("language"),
	gd.Option(gp.Attr("value", ""), gp.Value("選択してください")),
	gd.Option(gp.Attr("value", "Go"), gp.Value("Go")),
	gd.Option(gp.Attr("value", "TypeScript"), gp.Value("TypeScript")),
	gd.Option(gp.Attr("value", "Python"), gp.Value("Python")),
	gd.Option(gp.Attr("value", "Rust"), gp.Value("Rust")),
),
```

- `gd.Select(...)` は `<select>` 要素を生成
- `gd.Option(...)` は各選択肢。`gp.Attr("value", ...)` で送信値を、`gp.Value(...)` で表示テキストを設定

## ステップ 4: チェックボックス — InputCheckbox

```go
gd.InputCheckbox(false, "web",
	gp.Name("interests"),
	gp.ID("interest-web"),
),
gd.Label(gp.For("interest-web"), gp.Value("Web 開発")),
```

- `gd.InputCheckbox(initCheck, value, children...)` — 第1引数は初期チェック状態、第2引数は `value` 属性
- `gp.For("interest-web")` で `for` 属性を使いラベルとチェックボックスを紐付けます
- 同じ `name` を持つチェックボックスは、フォーム送信時にスライスとして受け取れます

## ステップ 5: 送信ボタン — Submit

```go
gd.Submit(
	gp.Class("btn", "btn-primary"),
	gp.Value("送信"),
),
```

`gd.Submit(...)` は `type="submit"` が設定された `<input>` 要素を生成します。

## ステップ 6: フォーム送信の処理

```go
func submitHandle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := &SurveyResponse{
		Name:      r.FormValue("name"),
		Email:     r.FormValue("email"),
		Language:  r.FormValue("language"),
		Interests: r.Form["interests"],
	}
	mu.Lock()
	responses = append(responses, resp)
	mu.Unlock()
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}
```

チェックボックスの複数選択値は `r.Form["interests"]` でスライスとして取得します。

## ステップ 7: Unless で「結果なし」を表示

```go
ge.Unless(len(list) > 0,
	gd.P(gp.Class("text-muted"), gp.Value("まだ回答がありません。")),
),
```

- `ge.Unless(cond, component)` は条件が**偽**のときに描画します
- `ge.If` の逆です。「〜でなければ表示」というパターンに適しています

## ステップ 8: 結果の番号付きリスト表示

```go
ge.If(len(list) > 0,
	gd.Ol(
		ge.Each(list, func(item g.ConvertToMap) g.ComponentFunc {
			m := item.ToMap()
			name := m.Get("name").(string)
			lang := m.Get("language").(string)
			return gd.Li(
				gd.P(
					gd.B(gp.Value(name)),
					g.Literal(" — 好きな言語: "+lang),
				),
			)
		}),
	),
),
```

`gd.Ol` は `<ol>` 番号付きリストを生成します。`gd.Ul`（箇条書き）と同じインターフェースです。

## 実行方法

```bash
go run example/survey/survey.go
```

ブラウザで以下の URL を確認します:
- http://localhost:8830 — アンケートフォーム
- http://localhost:8830/results — 回答結果一覧

フォームに回答を入力して送信し、結果ページに反映されることを確認しましょう。

## 発展課題

1. 回答結果をテーブル形式（`gd.Table`）で表示してみましょう
2. ラジオボタン（`gp.Attr("type", "radio")`）を追加してみましょう
3. `gs.Style` を使って回答数に応じたバーチャートを CSS で表現してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gd.Select(children...)` | `<select>` 要素 |
| `gd.Option(children...)` | `<option>` 要素 |
| `gd.InputCheckbox(check, value, children...)` | `<input type="checkbox">` 要素 |
| `gd.Input(children...)` | `<input>` 要素 |
| `gd.Submit(children...)` | `<input type="submit">` 要素 |
| `gd.Ol(children...)` | `<ol>` 番号付きリスト |
| `gd.Li(children...)` | `<li>` リスト項目 |
| `gp.ID(text)` | `id` 属性を設定 |
| `gp.Name(text)` | `name` 属性を設定 |
| `ge.Unless(cond, component)` | 条件が偽のとき描画 |
| `ge.If(cond, true, false...)` | 条件分岐 |
