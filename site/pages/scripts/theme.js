// Applies the stored theme and wires the header toggle.
//
// Inlined into <head> by pages/scripts.go, and blocking: the theme has to be
// on <html> before the first paint, or a light-themed frame shows before the
// dark one. Inline is what makes that cheap - an external script would block
// the paint on a round trip instead. The toggle half waits for the DOM, since
// the button it binds to has not been parsed yet at that point.
(function () {
	var root = document.documentElement;

	// loom's dark CSS keys off .dark and gates its prefers-color-scheme rules
	// behind :not(.light), so an explicit .light wins even on a dark OS.
	// Exactly one of the two classes is always set.
	function apply(dark) {
		root.classList.toggle('dark', dark);
		root.classList.toggle('light', !dark);
	}

	function stored() {
		try {
			return localStorage.getItem('theme');
		} catch (e) {
			return null; // private mode, or storage disabled
		}
	}

	var saved = stored();
	apply(saved ? saved === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches);

	document.addEventListener('DOMContentLoaded', function () {
		var toggle = document.querySelector('[data-theme-toggle]');
		if (!toggle) return;
		toggle.addEventListener('click', function () {
			var dark = !root.classList.contains('dark');
			apply(dark);
			try {
				localStorage.setItem('theme', dark ? 'dark' : 'light');
			} catch (e) {}
		});
	});
})();
