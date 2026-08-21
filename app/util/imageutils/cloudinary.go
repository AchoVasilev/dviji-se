// Package imageutils rewrites Cloudinary delivery URLs so images are served at
// the size the page actually paints them.
//
// The transformation happens on delivery rather than on upload: the original
// stays in Cloudinary, so a crop can be redone and a future format can be
// adopted without re-uploading anything. Only the bytes on the wire change.
package imageutils

import (
	"fmt"
	"strconv"
	"strings"
)

// uploadMarker is the segment every Cloudinary delivery URL carries, and the
// point transformations are inserted at:
//
//	https://res.cloudinary.com/<cloud>/image/upload/v123/folder/name.jpg
//	                                        ^ here
const uploadMarker = "/image/upload/"

// autoFormat lets Cloudinary pick the format and quality per browser: WebP or
// AVIF where they are supported, the original format where they are not.
const autoFormat = "f_auto,q_auto"

// CardWidths are the widths offered for images rendered at card size. The
// largest covers a card on a 2x screen; the smallest a narrow phone.
var CardWidths = []int{400, 800, 1200}

// HeroWidths are the widths offered for the full width header of an article.
var HeroWidths = []int{640, 1024, 1600, 2000}

// ContentWidth caps images inside the article body, which is never wider than
// 848 CSS pixels.
const ContentWidth = 1600

// Resized returns the URL for a single width. A URL that is not a Cloudinary
// delivery URL is returned unchanged: the cover field accepts any address the
// author pastes, and rewriting a foreign one would produce a broken link.
//
// c_limit only ever shrinks, so an image smaller than the requested width is
// served as it is rather than upscaled into blur.
func Resized(rawURL string, width int) string {
	return transform(rawURL, fmt.Sprintf("%s,c_limit,w_%d", autoFormat, width))
}

// Optimized returns the URL with format and quality left to Cloudinary but no
// resizing, for images whose displayed size is not known here.
func Optimized(rawURL string) string {
	return transform(rawURL, autoFormat)
}

// Srcset builds a srcset value for the given widths. It returns an empty
// string when the URL cannot be transformed, which keeps the attribute off the
// element entirely rather than emitting a srcset of one useless candidate.
func Srcset(rawURL string, widths []int) string {
	if !isCloudinary(rawURL) {
		return ""
	}

	candidates := make([]string, 0, len(widths))
	for _, width := range widths {
		candidates = append(candidates, Resized(rawURL, width)+" "+strconv.Itoa(width)+"w")
	}

	return strings.Join(candidates, ", ")
}

// RewriteContentImages points the images inside stored post HTML at optimized
// URLs. The editor saves whatever URL the upload returned, so this is the only
// place the article body can be improved without rewriting the database.
//
// Only the Cloudinary delivery prefix is touched; the surrounding markup is
// left exactly as the author saved it.
func RewriteContentImages(html string) string {
	params := fmt.Sprintf("%s,c_limit,w_%d", autoFormat, ContentWidth)

	for _, prefix := range cloudinaryPrefixes(html) {
		from := prefix + uploadMarker

		// Rebuilt occurrence by occurrence rather than with ReplaceAll: a body
		// that already carries the transformation - saved after an earlier
		// render, or pasted by the author - must not have a second one nested
		// inside it, which would bill another derived asset for the same
		// picture.
		var out strings.Builder

		for rest := html; ; {
			at := strings.Index(rest, from)
			if at < 0 {
				out.WriteString(rest)
				break
			}

			out.WriteString(rest[:at])
			rest = rest[at+len(from):]

			out.WriteString(from)
			if !strings.HasPrefix(rest, params+"/") {
				out.WriteString(params + "/")
			}
		}

		html = out.String()
	}

	return html
}

func transform(rawURL string, params string) string {
	if !isCloudinary(rawURL) {
		return rawURL
	}

	// Already carries this transformation, so inserting it again would nest a
	// second one and cost an extra derived asset for the same picture.
	if strings.Contains(rawURL, uploadMarker+params) {
		return rawURL
	}

	return strings.Replace(rawURL, uploadMarker, uploadMarker+params+"/", 1)
}

func isCloudinary(rawURL string) bool {
	return strings.HasPrefix(rawURL, "https://res.cloudinary.com/") && strings.Contains(rawURL, uploadMarker)
}

// cloudinaryPrefixes finds the distinct URL prefixes present in the HTML, so
// the replacement is anchored to a real Cloudinary host rather than to any
// occurrence of "/image/upload/".
func cloudinaryPrefixes(html string) []string {
	const host = "https://res.cloudinary.com/"

	var prefixes []string

	for rest := html; ; {
		start := strings.Index(rest, host)
		if start < 0 {
			break
		}

		rest = rest[start:]

		end := strings.Index(rest, uploadMarker)
		if end < 0 {
			break
		}

		prefix := rest[:end]
		if !strings.Contains(prefix, "\"") && !strings.Contains(prefix, " ") {
			prefixes = appendUnique(prefixes, prefix)
		}

		rest = rest[end+len(uploadMarker):]
	}

	return prefixes
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}

	return append(values, value)
}
