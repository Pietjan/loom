// Makes the theme picker instant, and marks what is selected.
//
// The swatches are ordinary links, so this script is an enhancement twice
// over: without it a click still loads the page and theme.js applies the query
// string. What it adds is switching without a reload - every theme is CSS
// custom properties on <html>, so setting the attribute is the whole job - and
// the things a static page cannot know: which chip is current, and the command
// that reproduces what you are looking at.
//
// It reads the theme's own name from --loom-accent-name / --loom-base-name
// rather than keeping a list. That is what makes the untouched page markable
// (no attribute is set, but a theme is still in force) and what the Base chip
// needs, since "base" is whichever neutral is currently in play.
//
// Inlined at the end of the page by pages/scripts.go, after the chips it binds
// to. It shares theme.js's storage keys ('accent', 'base'); the picker writes
// them, and theme.js is what reads them back on the next page.
(function () {
	var root = document.documentElement;
	var chips = Array.prototype.slice.call(document.querySelectorAll('[data-theme-swatch]'));
	if (!chips.length) return;

	var command = document.querySelector('[data-theme-command]');
	// The Base chip has no color of its own - it is the monochrome accent, and
	// which neutral that is comes from the base row.
	var monoChip = document.querySelector('[data-theme-swatch="accent"][data-theme-value=""]');
	var monoFill = monoChip && monoChip.querySelector('[data-theme-swatch-fill]');
	// The neutrals are read off the base row instead of listed again here.
	var neutrals = chips
		.filter(function (chip) {
			return chip.getAttribute('data-theme-swatch') === 'base' && chip.getAttribute('data-theme-value');
		})
		.map(function (chip) {
			return chip.getAttribute('data-theme-value');
		});

	function name(which) {
		return getComputedStyle(root).getPropertyValue('--loom-' + which + '-name').trim().replace(/^["']|["']$/g, '');
	}

	function isMono(accent) {
		return neutrals.indexOf(accent) !== -1;
	}

	function set(kind, value) {
		if (value) {
			root.setAttribute('data-' + kind, value);
		} else {
			root.removeAttribute('data-' + kind);
		}
		try {
			if (value) {
				localStorage.setItem(kind, value);
			} else {
				localStorage.removeItem(kind);
			}
		} catch (e) {} // private mode, or storage disabled
	}

	// Marks are derived from what the stylesheet resolved to, not from the
	// click that caused it, so they stay right whichever way the theme was set
	// - a chip, a shared link, or what the last visit left in storage.
	function mark() {
		var accent = name('accent');
		var base = name('base');
		var linked = !root.hasAttribute('data-base');

		for (var i = 0; i < chips.length; i++) {
			var chip = chips[i];
			var value = chip.getAttribute('data-theme-value');
			var selected;
			if (chip.getAttribute('data-theme-swatch') === 'accent') {
				selected = value ? value === accent : isMono(accent);
			} else {
				selected = value ? value === base : linked;
			}
			chip.toggleAttribute('data-selected', selected);
			// A row of choices with one of them on: aria-pressed says that
			// about a link, where :checked has nothing to attach to.
			chip.setAttribute('aria-pressed', selected ? 'true' : 'false');
		}

		if (monoFill) {
			monoFill.className = monoFill.className.replace(/loom-swatch-\S+/, 'loom-swatch-' + base);
		}
	}

	// The command names the accent even when it is the default: copied out of
	// the page, it still has to say what it produces. -base is only there when
	// it was chosen, since an accent already carries the pairing.
	function describe() {
		if (!command) return;
		command.textContent = 'go run github.com/pietjan/loom/cmd/css -accent ' + name('accent') +
			(root.hasAttribute('data-base') ? ' -base ' + root.getAttribute('data-base') : '') +
			' -o css/input.css';
	}

	// The URL is kept in step so it stays the shareable form of the page, but
	// with replaceState: a row of swatches is one decision being made, and
	// pushing each try onto the history means the back button has to walk out
	// of them one at a time.
	function url() {
		var params = new URLSearchParams(location.search);
		params.set('accent', name('accent'));
		root.hasAttribute('data-base') ? params.set('base', root.getAttribute('data-base')) : params.delete('base');
		history.replaceState(null, '', location.pathname + '?' + params.toString() + location.hash);
	}

	function choose(kind, value) {
		if (kind === 'accent') {
			if (value) {
				set('accent', value);
				return;
			}
			// Base means the monochrome accent built from the neutral in play,
			// and that accent pairs with itself - so an override naming the
			// same neutral is saying nothing, and comes off. Read it first:
			// clearing it is what changes the answer.
			var mono = name('base');
			set('base', '');
			set('accent', mono);
			return;
		}
		// The link: back to whatever the accent pairs with, accent untouched.
		if (!value) {
			set('base', '');
			return;
		}
		// Picking a neutral while on a monochrome accent moves the accent with
		// it - otherwise Base stays selected while showing some other neutral's
		// dark surface, which is a theme nobody asked for. The accent carries
		// the pairing on its own, so the override comes back off.
		if (isMono(name('accent'))) {
			set('base', '');
			set('accent', value);
			return;
		}
		set('base', value);
	}

	for (var i = 0; i < chips.length; i++) {
		chips[i].addEventListener('click', function (event) {
			// Let modified clicks through: opening a theme in a new tab is a
			// reasonable way to compare two of them side by side.
			if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) return;
			event.preventDefault();
			choose(this.getAttribute('data-theme-swatch'), this.getAttribute('data-theme-value'));
			mark();
			describe();
			url();
		});
	}

	mark();
	describe();
})();
