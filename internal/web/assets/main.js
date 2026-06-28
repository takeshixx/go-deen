// Bootstraps the deen WebAssembly module. Kept in a separate file so the page
// can be served under a strict Content-Security-Policy (no inline scripts).
function showStartupError(message) {
	const fallback = document.getElementById("wasm-fallback");
	if (!fallback) {
		return;
	}
	const detail = document.createElement("p");
	detail.className = "wasm-error";
	detail.textContent = message;
	fallback.appendChild(detail);
}

async function instantiateWasm(go) {
	const response = await fetch("deen.wasm", { cache: "no-cache" });
	if (!response.ok) {
		throw new Error(`failed to download deen.wasm: HTTP ${response.status}`);
	}
	if (WebAssembly.instantiateStreaming) {
		try {
			return await WebAssembly.instantiateStreaming(response, go.importObject);
		} catch (err) {
			const bytes = await (await fetch("deen.wasm", { cache: "no-cache" })).arrayBuffer();
			return await WebAssembly.instantiate(bytes, go.importObject);
		}
	}
	const bytes = await response.arrayBuffer();
	return await WebAssembly.instantiate(bytes, go.importObject);
}

if (!("WebAssembly" in window)) {
	showStartupError("WebAssembly is not available in this browser.");
} else {
	const go = new Go();
	instantiateWasm(go).then((result) => {
		go.run(result.instance);
	}).catch((err) => {
		showStartupError(err instanceof Error ? err.message : String(err));
	});
}
