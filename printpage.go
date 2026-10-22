package main

// ---------------------------------------------------
// PRINT SUPPORT
// Uses the underlying WebView engine's native print via JS
// window.print() — this just wraps it with a Go-bindable hook
// ---------------------------------------------------

const printPageJS = `
(function() {
	window.print();
})();
`
