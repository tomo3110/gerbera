package main

import (
	"database/sql"
	"net/url"
	"strings"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)

type SearchView struct {
	baseView

	searchQuery string
	searchUsers []User
	searchPosts []PostWithMeta
}

func NewSearchView(db *sql.DB, hub *Hub) *SearchView {
	return &SearchView{baseView: baseView{db: db, hub: hub}}
}

func (v *SearchView) Mount(params gl.Params) error {
	if err := v.mountBase(params); err != nil {
		return err
	}
	if kw := params.Query.Get("keyword"); kw != "" {
		v.searchQuery = kw
		v.searchUsers, _ = dbSearchUsers(v.db, kw, 10)
		v.searchPosts, _ = dbSearchPosts(v.db, kw, v.userID, 20)
	}
	return nil
}

func (v *SearchView) Unmount() {
	v.unmountBase()
}

func (v *SearchView) HandleEvent(event string, payload gl.Payload) error {
	if handled, err := v.handlePostAction(event, payload); handled {
		if err != nil {
			return err
		}
		if v.searchQuery != "" {
			v.searchPosts, _ = dbSearchPosts(v.db, v.searchQuery, v.userID, 20)
		}
		return nil
	}

	switch event {
	case "searchInput":
		v.searchQuery = payload["value"]
		if strings.TrimSpace(v.searchQuery) == "" {
			v.searchUsers = nil
			v.searchPosts = nil
			v.ReplacePatch(v.buildSearchPath())
			return nil
		}
		v.searchUsers, _ = dbSearchUsers(v.db, v.searchQuery, 10)
		v.searchPosts, _ = dbSearchPosts(v.db, v.searchQuery, v.userID, 20)
		v.ReplacePatch(v.buildSearchPath())
	}
	return nil
}

func (v *SearchView) HandleParams(path string, params url.Values) error {
	if kw := params.Get("keyword"); kw != "" {
		v.searchQuery = kw
		v.searchUsers, _ = dbSearchUsers(v.db, kw, 10)
		v.searchPosts, _ = dbSearchPosts(v.db, kw, v.userID, 20)
	} else {
		v.searchQuery = ""
		v.searchUsers = nil
		v.searchPosts = nil
	}
	return nil
}

func (v *SearchView) HandleInfo(msg any) error {
	v.handleBaseInfo(msg)
	if _, ok := msg.(NewMessageNotif); ok {
		v.showToast("New message received", "info")
	}
	return nil
}

func (v *SearchView) buildSearchPath() string {
	if v.searchQuery != "" {
		q := url.Values{}
		q.Set("keyword", v.searchQuery)
		return "/search?" + q.Encode()
	}
	return "/search"
}

func (v *SearchView) Render() g.Components {
	var results g.Components

	if v.searchQuery != "" {
		if len(v.searchUsers) > 0 {
			results = append(results,
				gd.Div(gp.Class("search-section-label"), gp.Value("People")),
			)
			for _, u := range v.searchUsers {
				results = append(results, renderSearchUserItem(u))
			}
		}

		if len(v.searchPosts) > 0 {
			results = append(results,
				gd.Div(gp.Class("search-section-label"), gp.Value("Posts")),
			)
			for _, p := range v.searchPosts {
				results = append(results, postCard(p))
			}
		}

		if len(v.searchUsers) == 0 && len(v.searchPosts) == 0 {
			results = append(results, gu.EmptyState("No results found"))
		}
	}

	return g.Components{
		gd.Body(
			gd.Div(gp.Attr("style", "padding: var(--g-space-md) 0; font-size: 1.1rem; font-weight: 700"),
				gp.Value("Search"),
			),
			gu.Card(
				gd.Div(gp.Class("search-input-wrap"),
					gu.FormInput("search",
						gp.Attr("value", v.searchQuery),
						gp.Placeholder("Search people and posts..."),
						gl.Input("searchInput"),
						gl.Debounce(300),
					),
				),
				gd.Div(append(g.Components{gp.Key("sr:" + v.searchQuery)}, results...)...),
			),
			gul.Toast(v.toastVisible, v.toastMessage, v.toastVariant, "dismissToast"),
		),
	}
}
