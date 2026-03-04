(function() {
    var mapEl = document.getElementById('gpx-map');
    if (!mapEl) return;

    var gpxUrl = mapEl.getAttribute('data-gpx-url');
    if (!gpxUrl) return;

    var map = L.map('gpx-map');
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        maxZoom: 18
    }).addTo(map);

    map.setView([42.7, 25.5], 8);

    var noIcon = L.divIcon({ className: '', iconSize: [0, 0] });
    var gpx = new L.GPX(gpxUrl, {
        async: true,
        marker_options: {
            clickable: false,
            startIconUrl: null,
            endIconUrl: null,
            shadowUrl: null
        },
        markers: {
            startIcon: null,
            endIcon: null,
            wptIcons: { '': noIcon },
            wptTypeIcons: { '': noIcon },
            pointMatchers: [{ regex: /.*/, icon: noIcon }]
        },
        polyline_options: {
            color: '#dc2626',
            weight: 4,
            opacity: 0.8,
            lineCap: 'round'
        }
    });

    gpx.on('loaded', function(e) {
        map.fitBounds(e.target.getBounds(), { padding: [30, 30] });
    });

    gpx.addTo(map);
})();
