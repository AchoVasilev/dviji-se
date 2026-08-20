package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func fragmentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<div>fragment</div>"))
	})
}

// Pasting a fragment URL into the address bar would otherwise return unstyled
// markup with no layout.
func TestFragmentOnly_RedirectsDirectVisits(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/blog/recent", nil)
	w := httptest.NewRecorder()

	FragmentOnly("/blog", fragmentHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}

	if location := w.Header().Get("Location"); location != "/blog" {
		t.Errorf("Location = %q, want /blog", location)
	}
}

// A search must keep its terms when the visitor lands on the full page.
func TestFragmentOnly_PreservesQueryString(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/blog/search/suggestions?q=%D1%84%D0%B8%D1%82", nil)
	w := httptest.NewRecorder()

	FragmentOnly("/blog/search", fragmentHandler()).ServeHTTP(w, req)

	want := "/blog/search?q=%D1%84%D0%B8%D1%82"
	if location := w.Header().Get("Location"); location != want {
		t.Errorf("Location = %q, want %q", location, want)
	}
}

func TestFragmentOnly_ServesFragmentToHTMX(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/blog/recent", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	FragmentOnly("/blog", fragmentHandler()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	if body := w.Body.String(); body != "<div>fragment</div>" {
		t.Errorf("body = %q, want the fragment", body)
	}
}

// Fragments must not turn up in search results as partial pages.
func TestFragmentOnly_MarksResponsesNoindex(t *testing.T) {
	for _, htmx := range []bool{true, false} {
		req := httptest.NewRequest(http.MethodGet, "/blog/recent", nil)
		if htmx {
			req.Header.Set("HX-Request", "true")
		}

		w := httptest.NewRecorder()
		FragmentOnly("/blog", fragmentHandler()).ServeHTTP(w, req)

		if got := w.Header().Get("X-Robots-Tag"); got != "noindex" {
			t.Errorf("htmx=%v: X-Robots-Tag = %q, want noindex", htmx, got)
		}
	}
}
