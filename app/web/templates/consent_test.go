package templates

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func renderConsent(t *testing.T) string {
	t.Helper()

	var buf bytes.Buffer
	if err := Consent().Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render the consent markup: %v", err)
	}

	return buf.String()
}

func TestConsent_RendersBothDialogsAndTheButton(t *testing.T) {
	html := renderConsent(t)

	for _, id := range []string{`id="cookie-consent"`, `id="ad-consent"`, `id="ad-consent-button"`} {
		if !strings.Contains(html, id) {
			t.Errorf("consent markup is missing %s", id)
		}
	}

	// Every action the script dispatches on has to exist in the markup, and
	// vice versa: a typo in either half silently breaks a button.
	for _, action := range []string{"accept-all", "necessary-only", "open", "accept-ads", "decline-ads", "close"} {
		if !strings.Contains(html, `data-consent-action="`+action+`"`) {
			t.Errorf("consent markup is missing the %q action", action)
		}
	}
}

// Nothing may be visible or consented to before the visitor has answered, so
// neither dialog carries the class that opens it and the slots start hidden.
func TestConsent_StartsClosed(t *testing.T) {
	html := renderConsent(t)

	if strings.Contains(html, "is-open") {
		t.Error("a consent dialog is rendered already open")
	}

	if strings.Contains(html, "is-visible") {
		t.Error("the consent button is rendered visible; consent.js reveals it")
	}
}

func renderAdSlot(t *testing.T, id string, variant AdVariant) string {
	t.Helper()

	var buf bytes.Buffer
	if err := AdSlot(id, variant).Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render the %s ad slot: %v", variant, err)
	}

	return buf.String()
}

func TestAdSlot_CarriesItsIdAndVariant(t *testing.T) {
	for _, variant := range []AdVariant{AdRail, AdBanner, AdInline} {
		html := renderAdSlot(t, "ad-"+string(variant), variant)

		if !strings.Contains(html, `id="ad-`+string(variant)+`"`) {
			t.Errorf("the %s slot ignores the id it was given", variant)
		}

		// The variant class is what the stylesheet keys the placement and the
		// breakpoints off, so losing it silently makes every slot identical.
		if !strings.Contains(html, variant.Class()) {
			t.Errorf("the %s slot is missing the %s class", variant, variant.Class())
		}

		if !strings.Contains(html, "data-ad-slot") {
			t.Errorf("the %s slot is missing data-ad-slot, so consent.js cannot find it", variant)
		}
	}
}

// Nothing about a slot may depend on a class that consent.js toggles: the
// stylesheet hides every [data-ad-slot] until the document is marked ads-on,
// and a stray `hidden` here would fight the breakpoint rules.
func TestAdSlot_HiddenByTheStylesheetAndEmpty(t *testing.T) {
	for _, variant := range []AdVariant{AdRail, AdBanner, AdInline} {
		html := renderAdSlot(t, "ad-"+string(variant), variant)

		if strings.Contains(html, `class="hidden`) || strings.Contains(html, ` hidden `) {
			t.Errorf("the %s slot carries a hidden class; visibility belongs to the stylesheet", variant)
		}

		if !strings.Contains(html, `aria-hidden="true"`) {
			t.Errorf("the %s slot must start aria-hidden", variant)
		}

		// The slot must not reference an ad network: consent decides whether
		// one is ever contacted, and markup in the page would bypass that.
		for _, host := range []string{"googlesyndication", "doubleclick", "adsbygoogle", "<script"} {
			if strings.Contains(html, host) {
				t.Errorf("the %s slot ships %q in the markup, which loads before consent", variant, host)
			}
		}
	}
}

// The cookie dialog is where most visitors will first look for the policy, so
// the link has to survive edits to the wording around it.
func TestConsent_CookieDialogLinksToThePolicy(t *testing.T) {
	if !strings.Contains(renderConsent(t), `href="/privacy"`) {
		t.Error("the cookie dialog does not link to the privacy policy")
	}
}
