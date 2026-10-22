package main

// ---------------------------------------------------
// FIND-IN-PAGE INJECTED JS
// Highlights matches inside the iframe's document
// ---------------------------------------------------

const findInPageJS = `
(function() {
	window.__sparrowFind = window.__sparrowFind || {
		matches: [],
		current: -1,

		clear: function() {
			document.querySelectorAll('mark.sparrow-find-mark').forEach(function(el) {
				var parent = el.parentNode;
				parent.replaceChild(document.createTextNode(el.textContent), el);
				parent.normalize();
			});
			this.matches = [];
			this.current = -1;
		},

		search: function(query) {
			this.clear();
			if (!query) return 0;

			var walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT, null, false);
			var nodes = [];
			var node;
			var chromeRoot = document.getElementById('__sparrow_chrome');
			while (node = walker.nextNode()) {
				if (node.parentNode.tagName === 'SCRIPT' || node.parentNode.tagName === 'STYLE') continue;
				if (chromeRoot && chromeRoot.contains(node)) continue;
				nodes.push(node);
			}

			var lowerQuery = query.toLowerCase();
			var self = this;

			nodes.forEach(function(textNode) {
				var text = textNode.textContent;
				var lowerText = text.toLowerCase();
				var idx = lowerText.indexOf(lowerQuery);
				if (idx === -1) return;

				var frag = document.createDocumentFragment();
				var lastIndex = 0;
				var searchIdx = 0;

				while ((searchIdx = lowerText.indexOf(lowerQuery, lastIndex)) !== -1) {
					frag.appendChild(document.createTextNode(text.slice(lastIndex, searchIdx)));
					var mark = document.createElement('mark');
					mark.className = 'sparrow-find-mark';
					mark.style.background = '#ffd93d';
					mark.style.color = '#000';
					mark.textContent = text.slice(searchIdx, searchIdx + query.length);
					frag.appendChild(mark);
					self.matches.push(mark);
					lastIndex = searchIdx + query.length;
				}
				frag.appendChild(document.createTextNode(text.slice(lastIndex)));
				textNode.parentNode.replaceChild(frag, textNode);
			});

			if (this.matches.length > 0) {
				this.current = 0;
				this.jumpTo(0);
			}
			return this.matches.length;
		},

		jumpTo: function(index) {
			this.matches.forEach(function(m) { m.style.background = '#ffd93d'; });
			if (this.matches[index]) {
				this.matches[index].style.background = '#ff9d4d';
				this.matches[index].scrollIntoView({ behavior: 'smooth', block: 'center' });
			}
			this.current = index;
		},

		next: function() {
			if (this.matches.length === 0) return -1;
			this.current = (this.current + 1) % this.matches.length;
			this.jumpTo(this.current);
			return this.current;
		},

		prev: function() {
			if (this.matches.length === 0) return -1;
			this.current = (this.current - 1 + this.matches.length) % this.matches.length;
			this.jumpTo(this.current);
			return this.current;
		}
	};
})();
`
