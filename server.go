package gerbera

import "net/http"

func NewServeMux(c ...ComponentFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := ExecuteTemplate(w, "ja", c...); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
	})
	return mux
}
