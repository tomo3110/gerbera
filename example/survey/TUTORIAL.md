# Survey Form Tutorial

## Overview

Build a survey form with dropdowns and checkboxes. This tutorial covers the following features:

- `Select`, `Option` — Dropdown lists
- `InputCheckbox` — Checkbox inputs
- `Input` — Text/email inputs
- `Submit` — Submit button
- `Unless` — Negative conditional rendering
- `If` — Conditional rendering
- `Name`, `ID` — Form attributes
- `Ol`, `Li` — Ordered lists

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Data Model and Storage

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

Responses are stored in an in-memory slice. A `sync.Mutex` is used for safe concurrent access.

## Step 2: Text Input — Input

```go
gd.Input(
	gp.Attr("type", "text"),
	gp.Class("form-control"),
	gp.ID("name"),
	gp.Name("name"),
	gp.Attr("required", "required"),
),
```

- `gd.Input(...)` generates an `<input>` void element
- `gp.ID("name")` sets `id="name"`
- `gp.Name("name")` sets `name="name"` (the key used on form submission)
- `type` is specified via `gp.Attr`

## Step 3: Dropdown — Select / Option

```go
gd.Select(
	gp.Class("form-control"),
	gp.ID("language"),
	gp.Name("language"),
	gd.Option(gp.Attr("value", ""), gp.Value("Please select")),
	gd.Option(gp.Attr("value", "Go"), gp.Value("Go")),
	gd.Option(gp.Attr("value", "TypeScript"), gp.Value("TypeScript")),
	gd.Option(gp.Attr("value", "Python"), gp.Value("Python")),
	gd.Option(gp.Attr("value", "Rust"), gp.Value("Rust")),
),
```

- `gd.Select(...)` generates a `<select>` element
- `gd.Option(...)` represents each choice. Use `gp.Attr("value", ...)` for the submitted value and `gp.Value(...)` for the display text

## Step 4: Checkbox — InputCheckbox

```go
gd.InputCheckbox(false, "web",
	gp.Name("interests"),
	gp.ID("interest-web"),
),
gd.Label(gp.Attr("for", "interest-web"), gp.Value("Web Development")),
```

- `gd.InputCheckbox(initCheck, value, children...)` — The 1st argument is the initial checked state, the 2nd is the `value` attribute
- Checkboxes sharing the same `name` are received as a slice on form submission

## Step 5: Submit Button — Submit

```go
gd.Submit(
	gp.Class("btn", "btn-primary"),
	gp.Value("Submit"),
),
```

`gd.Submit(...)` generates an `<input>` element with `type="submit"` set.

## Step 6: Handling Form Submission

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

Multiple checkbox values are retrieved as a slice using `r.Form["interests"]`.

## Step 7: Displaying "No Results" with Unless

```go
ge.Unless(len(list) > 0,
	gd.P(gp.Class("text-muted"), gp.Value("No responses yet.")),
),
```

- `ge.Unless(cond, component)` renders when the condition is **false**
- It is the inverse of `ge.If`, suitable for "show if not" patterns

## Step 8: Displaying Results as an Ordered List

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
					g.Literal(" — Favorite language: "+lang),
				),
			)
		}),
	),
),
```

`gd.Ol` generates an `<ol>` ordered list. It shares the same interface as `gd.Ul` (unordered list).

## Running

```bash
go run example/survey/survey.go
```

Open the following URLs in your browser:
- http://localhost:8830 — Survey form
- http://localhost:8830/results — Response results

Fill in the form, submit it, and verify the results appear on the results page.

## Exercises

1. Display the results in a table format (`gd.Table`)
2. Add radio buttons (`gp.Attr("type", "radio")`)
3. Use `gs.Style` to create a CSS-based bar chart based on response counts

## API Reference

| Function | Description |
|----------|-------------|
| `gd.Select(children...)` | `<select>` element |
| `gd.Option(children...)` | `<option>` element |
| `gd.InputCheckbox(check, value, children...)` | `<input type="checkbox">` element |
| `gd.Input(children...)` | `<input>` element |
| `gd.Submit(children...)` | `<input type="submit">` element |
| `gd.Ol(children...)` | `<ol>` ordered list |
| `gd.Li(children...)` | `<li>` list item |
| `gp.ID(text)` | Sets the `id` attribute |
| `gp.Name(text)` | Sets the `name` attribute |
| `ge.Unless(cond, component)` | Renders when condition is false |
| `ge.If(cond, true, false...)` | Conditional branching |
