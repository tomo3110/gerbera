# ダッシュボード チュートリアル

## 概要

社員データの一覧・詳細を表示するダッシュボードを構築します。このチュートリアルでは以下の機能を学びます。

- `styles.Style` — インライン CSS の適用
- テーブル要素: `Table`, `Thead`, `Tbody`, `Tfoot`, `Tr`, `Th`, `Td`
- `Each` + `ConvertToMap` — データ駆動のリスト描画
- `Map` 型の直接利用
- 複数ルート構成

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: データモデルの定義

```go
type Record struct {
	Name       string
	Department string
	Role       string
	Salary     int
}

func (r *Record) ToMap() g.Map {
	return g.Map{
		"name":       r.Name,
		"department": r.Department,
		"role":       r.Role,
		"salary":     r.Salary,
	}
}
```

`ConvertToMap` インターフェースを実装することで、`ge.Each()` でデータ駆動のテーブル行を生成できるようになります。

## ステップ 2: インライン CSS — styles.Style

```go
import gs "github.com/tomo3110/gerbera/styles"

gd.H2(
	gs.Style(g.StyleMap{"margin-top": "20px", "margin-bottom": "20px"}),
	gp.Value("社員ダッシュボード"),
),
```

- `gs.Style(styleMap)` は `style` 属性を設定する `ComponentFunc` を返します
- `g.StyleMap` は `map[string]any` のエイリアスです
- CSS プロパティ名をキーに、値を指定します

## ステップ 3: テーブルの構築

```go
gd.Table(
	gp.Class("table", "table-striped"),
	gd.Thead(
		gd.Tr(
			gd.Th(gp.Value("名前")),
			gd.Th(gp.Value("部署")),
			gd.Th(gp.Value("役職")),
			gd.Th(gp.Value("年収")),
		),
	),
	gd.Tbody(...),
	gd.Tfoot(...),
)
```

テーブル要素は以下の階層で構成されます:
- `Table` → `Thead` / `Tbody` / `Tfoot` → `Tr` → `Th` / `Td`
- すべて標準の `ComponentFunc` パターンに従います

## ステップ 4: Each でデータ行を生成

```go
list := make([]g.ConvertToMap, len(records))
for i, r := range records {
	list[i] = r
}

gd.Tbody(
	ge.Each(list, func(item g.ConvertToMap) g.ComponentFunc {
		m := item.ToMap()
		name := m.Get("name").(string)
		dept := m.Get("department").(string)
		role := m.Get("role").(string)
		salary := m.Get("salary").(int)
		return gd.Tr(
			gd.Td(gp.Value(name)),
			gd.Td(gp.Value(dept)),
			gd.Td(gp.Value(role)),
			gd.Td(gp.Value(formatYen(salary))),
		)
	}),
),
```

`ge.Each` のコールバック内で `ToMap().Get(key)` を使い、各フィールドを取得して型アサーションします。

## ステップ 5: フッター行と colspan

```go
gd.Tfoot(
	gd.Tr(
		gd.Td(
			gp.Attr("colspan", "3"),
			gp.Value("合計"),
		),
		gd.Td(gp.Value(formatYen(totalSalary()))),
		gd.Td(),
	),
),
```

`gp.Attr("colspan", "3")` で列を結合します。`gp.Attr` はあらゆる HTML 属性に使えます。

## ステップ 6: 複数ルートの構成

```go
func main() {
	http.HandleFunc("/detail/", detailHandle)
	http.HandleFunc("/", listHandle)
	log.Fatal(http.ListenAndServe(":8820", nil))
}
```

`g.Handler()` は静的ページの配信に推奨されるアプローチで、`g.HandlerFunc()` はリクエストに依存する動的ページ向けです。複数のルートが必要な場合は標準の `http.HandleFunc` を使い、各ハンドラ内で `g.ExecuteTemplate` を呼び出します。

## 実行方法

```bash
go run example/dashboard/dashboard.go
```

ブラウザで以下の URL を確認します:
- http://localhost:8820 — 社員一覧テーブル
- http://localhost:8820/detail/佐藤花子 — 個人詳細ページ

## 発展課題

1. テーブルにソート機能を追加してみましょう（クエリパラメータで並び替え）
2. `ge.If` を使って、年収が一定額以上の行を強調表示してみましょう
3. `gs.Style` で条件に応じた色分けを実装してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gs.Style(styleMap)` | インライン CSS を設定 |
| `g.StyleMap` | `map[string]any` 型（CSS プロパティマップ） |
| `gd.Table(children...)` | `<table>` 要素 |
| `gd.Thead(children...)` | `<thead>` 要素 |
| `gd.Tbody(children...)` | `<tbody>` 要素 |
| `gd.Tfoot(children...)` | `<tfoot>` 要素 |
| `gd.Tr(children...)` | `<tr>` 要素 |
| `gd.Th(children...)` | `<th>` 要素 |
| `gd.Td(children...)` | `<td>` 要素 |
| `ge.Each(list, callback)` | リスト反復 |
| `g.ConvertToMap` | `ToMap() Map` インターフェース |
