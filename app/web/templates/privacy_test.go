package templates

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func renderPrivacy(t *testing.T) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Privacy().Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render the privacy policy: %v", err)
	}

	return buf.String()
}

// The policy is a legal document: a section quietly disappearing is the kind
// of edit nothing else would catch.
func TestPrivacy_CoversEveryRequiredSection(t *testing.T) {
	html := renderPrivacy(t)

	for _, heading := range []string{
		"Какво събираме",
		"Бисквитки",
		"Реклами",
		"Кой друг вижда данните",
		"Колко дълго",
		"Твоите права",
		"Контакт",
	} {
		if !strings.Contains(html, heading) {
			t.Errorf("the policy is missing the %q section", heading)
		}
	}

	// Naming the supervisory authority is what tells a reader where to
	// complain, and it is easy to lose in a rewrite.
	if !strings.Contains(html, "Комисията за защита на личните данни") {
		t.Error("the policy does not name the supervisory authority")
	}

	if !strings.Contains(html, privacyUpdated) {
		t.Error("the policy does not show when it was last changed")
	}
}

func TestPrivacy_ShowsTheContactAddress(t *testing.T) {
	// TestMain sets APP_BASE_URL to https://dviji.se and the address is
	// derived from it when PRIVACY_CONTACT_EMAIL is unset.
	if got := os.Getenv("PRIVACY_CONTACT_EMAIL"); got != "" {
		t.Skipf("PRIVACY_CONTACT_EMAIL is set to %q in this environment", got)
	}

	html := renderPrivacy(t)

	if !strings.Contains(html, "mailto:privacy@dviji.se") {
		t.Error("the policy does not offer a contact address")
	}
}

// Canonical and og:url are absolute and built from APP_BASE_URL, so a policy
// without a path would be indexed under whatever URL it was reached by.
func TestPrivacy_IsIndexableAtItsOwnURL(t *testing.T) {
	html := renderPrivacy(t)

	if !strings.Contains(html, `rel="canonical" href="https://dviji.se/privacy"`) {
		t.Error("the policy does not declare its canonical URL")
	}
}
