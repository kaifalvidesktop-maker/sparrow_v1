package main

// ---------------------------------------------------
// RIGHT-CLICK CONTEXT MENU (injected JS + Go handlers)
// ---------------------------------------------------

const contextMenuJS = `
(function() {
	if (window.__sparrowContextMenuInstalled) return;
	window.__sparrowContextMenuInstalled = true;

	var menu = null;

	function closeMenu() {
		if (menu) { menu.remove(); menu = null; }
	}

	document.addEventListener('click', closeMenu);

	document.addEventListener('contextmenu', function(e) {
		e.preventDefault();
		closeMenu();

		var target = e.target;
		var linkEl = target.closest('a');
		var imgEl = target.tagName === 'IMG' ? target : null;

		menu = document.createElement('div');
		menu.style.cssText =
			'position:fixed;top:' + e.clientY + 'px;left:' + e.clientX + 'px;' +
			'background:#1c1d20;border:1px solid #2a2a2c;border-radius:10px;' +
			'box-shadow:0 10px 30px rgba(0,0,0,0.5);padding:6px;z-index:99999;' +
			'font-family:Segoe UI,sans-serif;font-size:13px;min-width:180px;';

		function addItem(label, onClick, disabled) {
			var item = document.createElement('div');
			item.innerText = label;
			item.style.cssText = 'padding:9px 14px;color:' + (disabled ? '#666' : '#ccc') + ';cursor:' + (disabled ? 'default' : 'pointer') + ';border-radius:6px;';
			// keep focus + selection inside the text field while clicking the menu
			item.onmousedown = function(ev) { ev.preventDefault(); };
			if (!disabled) {
				item.onmouseenter = function() { item.style.background = '#2a2b2f'; item.style.color = '#fff'; };
				item.onmouseleave = function() { item.style.background = 'transparent'; item.style.color = '#ccc'; };
			}
			item.onclick = function(ev) { ev.stopPropagation(); closeMenu(); if (!disabled) onClick(); };
			menu.appendChild(item);
		}

		function addSeparator() {
			var line = document.createElement('div');
			line.style.cssText = 'height:1px;margin:5px 6px;background:#2a2a2c;';
			menu.appendChild(line);
		}

		function canDo(cmd) {
			try { return document.queryCommandEnabled(cmd); } catch (err) { return false; }
		}

		var editable = target.closest ? target.closest('input, textarea, [contenteditable=""], [contenteditable="true"]') : null;
		if (editable && editable.tagName === 'INPUT') {
			var okTypes = ['text', 'search', 'url', 'tel', 'email', 'password', 'number', ''];
			if (okTypes.indexOf((editable.type || '').toLowerCase()) === -1) editable = null;
		}
		if (editable) {
			editable.focus();
			var hasSel = window.getSelection().toString() !== '' ||
				(typeof editable.selectionStart === 'number' && editable.selectionStart !== editable.selectionEnd);
			addItem('Undo', function() { editable.focus(); document.execCommand('undo'); }, !canDo('undo'));
			addItem('Redo', function() { editable.focus(); document.execCommand('redo'); }, !canDo('redo'));
			addSeparator();
			addItem('Cut', function() { editable.focus(); document.execCommand('cut'); }, !hasSel);
			addItem('Copy', function() { editable.focus(); document.execCommand('copy'); }, !hasSel);
			addItem('Paste', function() {
				editable.focus();
				navigator.clipboard.readText().then(function(text) {
					editable.focus();
					document.execCommand('insertText', false, text);
				}).catch(function() {});
			});
			addItem('Select All', function() {
				editable.focus();
				if (editable.select) editable.select(); else document.execCommand('selectAll');
			});
			addSeparator();
		}

		if (linkEl) {
			addItem('Open Link in New Tab', function() {
				window.openLinkInNewTab(linkEl.href);
			});
			addItem('Copy Link Address', function() {
				navigator.clipboard.writeText(linkEl.href);
			});
			addItem('Download Link', function() {
				window.downloadFile(linkEl.href);
			});
		}

		if (imgEl) {
			addItem('Open Image in New Tab', function() {
				window.openLinkInNewTab(imgEl.src);
			});
			addItem('Save Image As', function() {
				window.downloadFile(imgEl.src);
			});
		}

		addItem('Back', function() { goBack(); });
		addItem('Forward', function() { goForward(); });
		addItem('Reload', function() { reload(); });

		var sel = window.getSelection().toString();
		if (sel && sel.trim() !== '') {
			addItem('Search Google for "' + (sel.length > 20 ? sel.slice(0,20) + '...' : sel) + '"', function() {
				go(sel);
			});
			addItem('Copy', function() {
				navigator.clipboard.writeText(sel);
			});
		}

		addItem('Inspect', function() {
			window.openDevTools();
		});

		document.body.appendChild(menu);

		var rect = menu.getBoundingClientRect();
		if (rect.right > window.innerWidth) {
			menu.style.left = (window.innerWidth - rect.width - 8) + 'px';
		}
		if (rect.bottom > window.innerHeight) {
			menu.style.top = (window.innerHeight - rect.height - 8) + 'px';
		}
	});
})();
`
