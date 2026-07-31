// Drives the launcher demo on the component page. This is the site's
// script, not the library's - loom ships the markup and the ids and leaves
// this part to you. It is here so the page demonstrates the seam rather
// than describing it.
//
// Everything is delegated from document: handlers bind once, no element is
// ever touched at setup, and a panel swapped in later works with no
// re-init. That is the shape to copy for htmx or datastar, where binding
// per node is how you end up with listeners on nodes that no longer exist.
//
// Arrow keys move real focus rather than an aria-activedescendant cursor.
// The rows are links, so focus is what a link wants: Enter activates
// natively, the focus ring is already in the component recipe, and no ARIA
// role has to be claimed to explain any of it.
//
// Which rows are visible is this script's job; whether the panel or the
// empty message is showing follows from that in CSS, keyed on
// [data-ui=launcher-item]:not([hidden]).
(function () {
	function root(target) {
		return target.closest && target.closest('[data-ui=launcher]');
	}

	function shown(el) {
		return Array.prototype.slice
			.call(el.querySelectorAll('[data-ui=launcher-item]'))
			.filter(function (r) {
				return !r.hidden;
			});
	}

	function filter(el) {
		var input = el.querySelector('[data-ui=launcher-input]');
		var q = input.value.trim().toLowerCase();
		Array.prototype.forEach.call(
			el.querySelectorAll('[data-ui=launcher-item]'),
			function (r) {
				r.hidden = q !== '' && r.textContent.toLowerCase().indexOf(q) === -1;
			}
		);
	}

	function move(el, by) {
		var rows = shown(el);
		if (!rows.length) return;
		var at = rows.indexOf(document.activeElement);
		// From the input (at === -1) both directions enter the list: down at
		// the top, up at the bottom.
		var next = at === -1 ? (by > 0 ? 0 : rows.length - 1) : at + by;
		if (next < 0 || next >= rows.length) {
			el.querySelector('[data-ui=launcher-input]').focus();
			return;
		}
		rows[next].focus();
		rows[next].scrollIntoView({ block: 'nearest' });
	}

	document.addEventListener('input', function (e) {
		var el = root(e.target);
		if (el && e.target.matches('[data-ui=launcher-input]')) filter(el);
	});

	// The shortcut is the one part neither the platform nor loom supplies.
	// Everything it opens - the dialog, the backdrop, light dismiss, Escape
	// to close - is already there; this only calls showModal.
	//
	// Bound on document like the rest, so it keeps working if the dialog is
	// swapped in later. It looks the element up per press rather than
	// holding a reference, for the same reason.
	document.addEventListener('keydown', function (e) {
		if (e.key !== 'k' || !(e.metaKey || e.ctrlKey)) return;
		var dialog = document.querySelector('[data-launcher-hotkey]');
		if (!dialog) return;
		e.preventDefault();
		if (dialog.open) {
			dialog.close();
			return;
		}
		dialog.showModal();
		// Reopening should not inherit the last search. showModal focuses
		// the first focusable element, which is the query field, so there
		// is nothing to focus by hand.
		var el = dialog.querySelector('[data-ui=launcher]');
		if (el) {
			el.querySelector('[data-ui=launcher-input]').value = '';
			filter(el);
		}
	});

	document.addEventListener('keydown', function (e) {
		var el = root(e.target);
		if (!el) return;

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			move(el, 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			move(el, -1);
		} else if (e.key === 'Escape') {
			var input = el.querySelector('[data-ui=launcher-input]');
			if (!input.value) return;
			e.preventDefault();
			input.value = '';
			filter(el);
			input.focus();
		}
	});
})();
