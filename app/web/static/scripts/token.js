'use strict'

document.addEventListener('DOMContentLoaded', function() {
	const metaCsrf = document.querySelector('meta[name="csrf-token"]');
	if (metaCsrf) {
		const token = metaCsrf.content;
		document.body.addEventListener('htmx:configRequest', function(ev) {
			ev.detail.headers['X-CSRF-Token'] = token;
		});
	}
});
