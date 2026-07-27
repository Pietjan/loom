// Applies the stored theme and wires the header toggle.
//
// Inlined into <head> by pages/scripts.go, and blocking: the theme has to be
// on <html> before the first paint, or a light-themed frame shows before the
// dark one. Inline is what makes that cheap - an external script would block
// the paint on a round trip instead. The toggle half waits for the DOM, since
// the button it binds to has not been parsed yet at that point.
(function () {
	var root = document.documentElement;

	function stored(key) {
		try {
			return localStorage.getItem(key);
		} catch (e) {
			return null; // private mode, or storage disabled
		}
	}

	function store(key, value) {
		try {
			if (value) {
				localStorage.setItem(key, value);
			} else {
				localStorage.removeItem(key);
			}
		} catch (e) {}
	}

	// loom's dark CSS keys off .dark and gates its prefers-color-scheme rules
	// behind :not(.light), so an explicit .light wins even on a dark OS.
	// Exactly one of the two classes is always set.
	function apply(dark) {
		root.classList.toggle('dark', dark);
		root.classList.toggle('light', !dark);
	}

	var params = new URLSearchParams(location.search);

	// ?theme=dark|light is the same bargain as the colors below: a param wins
	// and is remembered, an empty one hands the choice back to the OS. It is
	// what lets one link carry a whole theme rather than half of one.
	var mode = params.get('theme');
	if (mode !== null) {
		mode = mode.trim().toLowerCase();
		store('theme', mode === 'dark' || mode === 'light' ? mode : '');
	}

	var saved = stored('theme');
	apply(saved ? saved === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches);

	// Accent and base come from the query string - ?accent=orange&base=neutral
	// - and land on the html[data-accent] / html[data-base] rules that
	// cmd/css/themes generates. A param is remembered, because the sidebar
	// links carry no query string and the choice would otherwise last exactly
	// one click; an empty one (?accent=) clears it again.
	//
	// The value is only shape-checked. The stylesheet is the list of real
	// accents, and one it does not know matches no rule, leaving the default
	// theme in place - which beats keeping a second copy of the palette names
	// here, out of step with the Go table that generates them.
	function token(name) {
		var value = params.get(name);
		if (value === null) return stored(name);
		value = value.trim().toLowerCase();
		if (!/^[a-z]+$/.test(value)) value = '';
		store(name, value);
		return value;
	}

	['accent', 'base'].forEach(function (name) {
		var value = token(name);
		if (value) {
			root.setAttribute('data-' + name, value);
		} else {
			root.removeAttribute('data-' + name);
		}
	});

	function choose(dark) {
		apply(dark);
		store('theme', dark ? 'dark' : 'light');
	}

	document.addEventListener('DOMContentLoaded', function () {
		var toggle = document.querySelector('[data-theme-toggle]');
		if (toggle) {
			toggle.addEventListener('click', function () {
				choose(!root.classList.contains('dark'));
			});
		}

		// The theme page's light/dark pair. They are links, so they still work
		// as a page load; caught here they switch without one, and the URL is
		// kept in step so it stays the shareable form of what is on screen.
		// Only the theme param is touched - the picker owns the other two.
		var modes = document.querySelectorAll('[data-theme-mode]');
		for (var i = 0; i < modes.length; i++) {
			modes[i].addEventListener('click', function (event) {
				if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) return;
				event.preventDefault();
				var mode = this.getAttribute('data-theme-mode');
				choose(mode === 'dark');
				var next = new URLSearchParams(location.search);
				next.set('theme', mode);
				history.replaceState(null, '', location.pathname + '?' + next.toString() + location.hash);
			});
		}
	});
})();
