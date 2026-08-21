package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"server/util/imageutils"
)

func renderImage(t *testing.T, img ResponsiveImage) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Image(img).Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render the image: %v", err)
	}

	return buf.String()
}

const cloudinaryCover = "https://res.cloudinary.com/demo/image/upload/v1/dviji-se/cover.jpg"

func TestImage_OffersEveryWidthWithSizes(t *testing.T) {
	html := renderImage(t, ResponsiveImage{
		URL:           cloudinaryCover,
		Alt:           "Заглавие",
		Class:         "w-full",
		Sizes:         "100vw",
		Widths:        imageutils.CardWidths,
		FallbackWidth: 800,
	})

	if !strings.Contains(html, "w_800/") {
		t.Error("the plain src is not resized to the fallback width")
	}

	for _, width := range imageutils.CardWidths {
		if !strings.Contains(html, "400w") && width == 400 {
			t.Errorf("srcset is missing the %dw candidate", width)
		}
	}

	// srcset without sizes makes the browser assume 100vw and over-fetch on
	// every layout narrower than the viewport.
	if !strings.Contains(html, `sizes="100vw"`) {
		t.Error("srcset was emitted without sizes")
	}
}

// A pasted address cannot be transformed, and a srcset with one candidate
// only repeats what src already says.
func TestImage_NoSrcsetForForeignURLs(t *testing.T) {
	html := renderImage(t, ResponsiveImage{
		URL:           "https://example.com/photo.jpg",
		Alt:           "Външна",
		Widths:        imageutils.CardWidths,
		FallbackWidth: 800,
	})

	if strings.Contains(html, "srcset") {
		t.Errorf("a foreign URL got a srcset: %s", html)
	}

	if !strings.Contains(html, `src="https://example.com/photo.jpg"`) {
		t.Errorf("a foreign URL was rewritten: %s", html)
	}
}

// Lazy loading anything above the fold delays the largest paint, so it has to
// stay opt-in rather than a default.
func TestImage_LazyOnlyWhenAsked(t *testing.T) {
	eager := renderImage(t, ResponsiveImage{URL: cloudinaryCover, FallbackWidth: 2000})
	if strings.Contains(eager, `loading="lazy"`) {
		t.Error("an image is lazy without being asked")
	}

	lazy := renderImage(t, ResponsiveImage{URL: cloudinaryCover, FallbackWidth: 800, Lazy: true})
	if !strings.Contains(lazy, `loading="lazy"`) {
		t.Error("Lazy was ignored")
	}
}

func TestImage_ThumbnailNeedsNoSizes(t *testing.T) {
	html := renderImage(t, ResponsiveImage{URL: cloudinaryCover, FallbackWidth: 80})

	if strings.Contains(html, "sizes=") {
		t.Error("a single width image should not carry sizes")
	}

	if !strings.Contains(html, "w_80/") {
		t.Error("the thumbnail is not resized")
	}
}
