// Filters the icon browser grid as you type. Progressive enhancement: with
// the script blocked, the grid still renders every icon and the browser's own
// find-in-page still works — only the search box goes away.
(function () {
	var root = document.querySelector('[data-icon-browser]');
	if (!root) return;

	var input = root.querySelector('[data-icon-search]');
	var count = root.querySelector('[data-icon-count]');
	var empty = root.querySelector('[data-icon-empty]');
	var cells = Array.prototype.slice.call(root.querySelectorAll('[data-icon]'));
	if (!input || !cells.length) return;

	// The search box is useless without this script, so it ships hidden and is
	// revealed here rather than flashing a dead control.
	input.closest('[data-icon-search-field]').hidden = false;

	// Each cell is matched against its constant name and its kebab-case value
	// ("ArrowLeft arrow-left"), so both "arrowleft" and "arrow-left" hit.
	var haystacks = cells.map(function (cell) {
		return cell.getAttribute('data-icon').toLowerCase();
	});

	function apply() {
		var query = input.value.trim().toLowerCase().replace(/[\s-]/g, '');
		var shown = 0;
		for (var i = 0; i < cells.length; i++) {
			var match = !query || haystacks[i].replace(/[\s-]/g, '').indexOf(query) !== -1;
			cells[i].hidden = !match;
			if (match) shown++;
		}
		count.textContent = shown === cells.length
			? cells.length + ' icons'
			: shown + ' of ' + cells.length + ' icons';
		empty.hidden = shown !== 0;
	}

	input.addEventListener('input', apply);
	// Escape clears, matching what a search field is expected to do.
	input.addEventListener('keydown', function (e) {
		if (e.key === 'Escape' && input.value) {
			input.value = '';
			apply();
		}
	});
	apply();
})();
