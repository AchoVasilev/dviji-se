package middleware

import (
	"net/http"
	"server/util/httputils"
)

// FragmentOnly guards an endpoint that renders an HTMX fragment rather than a
// whole document. Reaching one directly - by pasting the URL, opening it in a
// new tab, or by a crawler following it - would otherwise return unstyled
// markup with no <html>, stylesheet or navigation.
//
// Requests without the HTMX header are sent to the page that hosts the
// fragment, carrying the query string so a search keeps its terms. The noindex
// header keeps the fragment out of search results even when it is fetched
// directly.
func FragmentOnly(hostPage string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex")

		if httputils.IsHTMXRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		target := hostPage
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		http.Redirect(w, r, target, http.StatusSeeOther)
	})
}
