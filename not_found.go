package tigertonic

import (
	"fmt"
	"net/http"
)

// NotFoundHandler responds 404 to every request, possibly with a JSON body.
type NotFoundHandler struct{}

func (NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	description := fmt.Sprintf("%s %s not found", r.Method, r.URL.Path)
	if acceptJSON(r) {
		/*
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
				var e string
				if SnakeCaseHTTPEquivErrors {
					e = "not_found"
				} else {
					e = "tigertonic.NotFound"
				}
					if err := json.NewEncoder(w).Encode(map[string]string{
						"description": description,
						"error":       e,
					}); nil != err {
						log.Println(err)
					}*/

		WriteJSONError(w, NewMethodNotFoundError(description))

	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, description)
	}
}
