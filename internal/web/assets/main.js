// Bootstraps the deen WebAssembly module. Kept in a separate file so the page
// can be served under a strict Content-Security-Policy (no inline scripts).
const go = new Go();
WebAssembly.instantiateStreaming(fetch("deen.wasm"), go.importObject).then((result) => {
	go.run(result.instance);
});
