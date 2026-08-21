package imageutils

import (
	"strings"
	"testing"
)

const cloudinaryURL = "https://res.cloudinary.com/demo/image/upload/v1691940539/dviji-se/photo.jpg"

func TestResized_InsertsTheTransformation(t *testing.T) {
	got := Resized(cloudinaryURL, 800)
	want := "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,c_limit,w_800/v1691940539/dviji-se/photo.jpg"

	if got != want {
		t.Errorf("Resized() = %q, want %q", got, want)
	}
}

// The cover field is a plain URL input, so an author can paste an address from
// anywhere. Rewriting one of those would produce a 404 instead of an image.
func TestResized_LeavesForeignURLsAlone(t *testing.T) {
	for _, url := range []string{
		"https://example.com/photo.jpg",
		"/static/img/running-man.png",
		"https://res.cloudinary.com/demo/raw/upload/v1/file.gpx",
		"",
	} {
		if got := Resized(url, 800); got != url {
			t.Errorf("Resized(%q) = %q, want it unchanged", url, got)
		}
	}
}

// Transforming an already transformed URL would nest a second transformation,
// billing another derived asset for the same picture.
func TestResized_IsIdempotent(t *testing.T) {
	once := Resized(cloudinaryURL, 800)

	if twice := Resized(once, 800); twice != once {
		t.Errorf("Resized() applied twice = %q, want %q", twice, once)
	}
}

func TestSrcset_ListsEveryWidth(t *testing.T) {
	got := Srcset(cloudinaryURL, []int{400, 800})

	for _, want := range []string{"c_limit,w_400/", " 400w", "c_limit,w_800/", " 800w"} {
		if !strings.Contains(got, want) {
			t.Errorf("Srcset() = %q, missing %q", got, want)
		}
	}

	if strings.Count(got, ",") < 1 || !strings.Contains(got, ", ") {
		t.Errorf("Srcset() = %q, want comma separated candidates", got)
	}
}

// An untransformable URL has exactly one candidate, which is what the plain src
// already provides; emitting it as a srcset only adds noise.
func TestSrcset_EmptyForForeignURLs(t *testing.T) {
	if got := Srcset("https://example.com/photo.jpg", CardWidths); got != "" {
		t.Errorf("Srcset() = %q, want empty", got)
	}
}

func TestRewriteContentImages(t *testing.T) {
	html := `<p>Преди</p><img src="` + cloudinaryURL + `" alt="Снимка"><p>След</p>`

	got := RewriteContentImages(html)

	if !strings.Contains(got, "f_auto,q_auto,c_limit,w_1600/v1691940539") {
		t.Errorf("RewriteContentImages() did not optimize the image: %q", got)
	}

	// The surrounding markup is the author's; only the URL may change.
	for _, want := range []string{"<p>Преди</p>", `alt="Снимка"`, "<p>След</p>"} {
		if !strings.Contains(got, want) {
			t.Errorf("RewriteContentImages() lost %q", want)
		}
	}
}

func TestRewriteContentImages_LeavesOtherContentAlone(t *testing.T) {
	for _, html := range []string{
		`<img src="https://example.com/image/upload/photo.jpg">`,
		`<p>Текст без снимки</p>`,
		"",
	} {
		if got := RewriteContentImages(html); got != html {
			t.Errorf("RewriteContentImages(%q) = %q, want it unchanged", html, got)
		}
	}
}

func TestRewriteContentImages_IsIdempotent(t *testing.T) {
	once := RewriteContentImages(`<img src="` + cloudinaryURL + `">`)

	if twice := RewriteContentImages(once); twice != once {
		t.Errorf("RewriteContentImages() applied twice = %q, want %q", twice, once)
	}
}
