// Cookie and advertising consent.
//
// The public Content-Security-Policy is script-src 'self' with no
// 'unsafe-inline', so every listener here is bound from this file - inline
// handlers in the markup would be blocked.
//
// One cookie holds the answer:
//   consent=necessary  only the cookies the site needs to work
//   consent=ads        the above plus advertising
// Absent means unanswered, which is treated as a refusal until the visitor
// says otherwise.

(function () {
	'use strict';

	const COOKIE_NAME = 'consent';
	const CONSENT_NECESSARY = 'necessary';
	const CONSENT_ADS = 'ads';

	// A year is the usual ceiling for a consent record: long enough not to
	// nag, short enough that the choice is revisited.
	const COOKIE_MAX_AGE = 365 * 24 * 60 * 60;

	function readConsent() {
		const match = document.cookie.match(/(?:^|;\s*)consent=([^;]*)/);
		if (!match) {
			return '';
		}

		const value = decodeURIComponent(match[1]);

		return value === CONSENT_ADS || value === CONSENT_NECESSARY ? value : '';
	}

	function writeConsent(value) {
		// Secure is only added over HTTPS: on a plain-HTTP localhost the
		// browser would drop the cookie and the dialog would never go away.
		const secure = window.location.protocol === 'https:' ? '; Secure' : '';
		document.cookie = `${COOKIE_NAME}=${encodeURIComponent(value)}; Path=/; Max-Age=${COOKIE_MAX_AGE}; SameSite=Lax${secure}`;
	}

	function adsAllowed() {
		return readConsent() === CONSENT_ADS;
	}

	let lastFocused = null;

	function openDialog(dialog) {
		if (!dialog) {
			return;
		}

		lastFocused = document.activeElement;
		dialog.classList.add('is-open');

		const focusTarget = dialog.querySelector('.consent-accept');
		if (focusTarget) {
			focusTarget.focus();
		}
	}

	function closeDialog(dialog) {
		if (!dialog) {
			return;
		}

		dialog.classList.remove('is-open');

		if (lastFocused && typeof lastFocused.focus === 'function') {
			lastFocused.focus();
		}

		lastFocused = null;
	}

	// applyConsent is the only place that reveals or hides advertising, so it
	// can be re-run after an htmx swap brings new slots into the page.
	//
	// The switch is a single class on the document rather than one per slot:
	// which placement is appropriate depends on the viewport, and that belongs
	// to the .ad-slot-* media queries in the stylesheet, not to this file.
	function applyConsent() {
		const allowed = adsAllowed();
		const slots = document.querySelectorAll('[data-ad-slot]');

		document.documentElement.classList.toggle('ads-on', allowed);

		slots.forEach(function (slot) {
			slot.setAttribute('aria-hidden', allowed ? 'false' : 'true');
		});

		// The ad network is loaded by whoever registers this hook, and only
		// ever from here - so an unconsented visitor never fetches it. Nothing
		// registers it yet, which is why the slots stay empty placeholders.
		if (allowed && window.dvijiSe && typeof window.dvijiSe.renderAds === 'function') {
			window.dvijiSe.renderAds(slots);
		}

		document.dispatchEvent(new CustomEvent('consentchange', {
			detail: { consent: readConsent(), adsAllowed: allowed },
		}));
	}

	function decide(value) {
		writeConsent(value);
		applyConsent();
	}

	function handleAction(action, cookieDialog, adDialog) {
		switch (action) {
			case 'accept-all':
				decide(CONSENT_ADS);
				closeDialog(cookieDialog);
				break;

			case 'necessary-only':
				decide(CONSENT_NECESSARY);
				closeDialog(cookieDialog);
				break;

			case 'open':
				openDialog(adDialog);
				break;

			case 'accept-ads':
				decide(CONSENT_ADS);
				closeDialog(adDialog);
				break;

			// Declining hides the ads and closes the dialog, but leaves the
			// button in place so the answer can be changed later.
			case 'decline-ads':
				decide(CONSENT_NECESSARY);
				closeDialog(adDialog);
				break;

			case 'close':
				closeDialog(adDialog);
				break;
		}
	}

	function init() {
		const cookieDialog = document.getElementById('cookie-consent');
		const adDialog = document.getElementById('ad-consent');
		const adButton = document.getElementById('ad-consent-button');

		if (!cookieDialog && !adDialog && !adButton) {
			return;
		}

		document.addEventListener('click', function (event) {
			const trigger = event.target.closest('[data-consent-action]');
			if (trigger) {
				handleAction(trigger.dataset.consentAction, cookieDialog, adDialog);
				return;
			}

			// A click on the backdrop closes the ad dialog only. The cookie
			// dialog needs an answer, and dismissing it would record none.
			if (event.target === adDialog) {
				closeDialog(adDialog);
			}
		});

		document.addEventListener('keydown', function (event) {
			if (event.key === 'Escape') {
				closeDialog(adDialog);
			}
		});

		if (adButton) {
			adButton.classList.add('is-visible');
		}

		// Unanswered means the cookie dialog leads; the ad question is part of
		// it, so the floating button only matters afterwards.
		if (readConsent() === '') {
			openDialog(cookieDialog);
		}

		applyConsent();

		// Slots can arrive with swapped-in content, so consent is re-applied
		// to whatever is in the page afterwards.
		document.body.addEventListener('htmx:afterSwap', applyConsent);
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', init);
	} else {
		init();
	}
})();
