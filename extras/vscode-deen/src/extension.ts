import * as vscode from 'vscode';
import * as child from 'child_process';
import { allowedNodeEnvironmentFlags } from 'process';

export function activate(context: vscode.ExtensionContext) {
	
	addDeenPlugin('b64', 'base64Encode', context)
	addDeenPlugin('.b64', 'base64Decode', context)
	addDeenPlugin('b32', 'base32Encode', context)
	addDeenPlugin('.b32', 'base32Decode', context)
	addDeenPlugin('b85', 'base85Encode', context)
	addDeenPlugin('.b85', 'base85Decode', context)
	addDeenPlugin('url', 'urlEncode', context)
	addDeenPlugin('.url', 'urlDecode', context)
	addDeenPlugin('html', 'htmlEncode', context)
	addDeenPlugin('.html', 'htmlDecode', context)
	addDeenPlugin('hex', 'hexEncode', context)
	addDeenPlugin('.hex', 'hexDecode', context)
	addDeenPlugin('json', 'jsonFormat', context)
	addDeenPlugin('.json', 'jsonUnformat', context)
	addDeenPlugin('flate', 'flateCompress', context)
	addDeenPlugin('.flate', 'flateUncompress', context)
	addDeenPlugin('lzma', 'lzmaCompress', context)
	addDeenPlugin('.lzma', 'lzmaUncompress', context)
	addDeenPlugin('lzma2', 'lzma2Compress', context)
	addDeenPlugin('.lzma2', 'lzma2Uncompress', context)
	addDeenPlugin('lzw', 'lzwCompress', context)
	addDeenPlugin('.lzw', 'lzwUncompress', context)
	addDeenPlugin('gzip', 'gzipCompress', context)
	addDeenPlugin('.gzip', 'gzipUncompress', context)
	addDeenPlugin('zlib', 'zlibCompress', context)
	addDeenPlugin('.zlib', 'zlibUncompress', context)
	addDeenPlugin('bzip2', 'bzip2Compress', context)
	addDeenPlugin('.bzip2', 'bzip2Uncompress', context)
	addDeenPlugin('brotli', 'brotliCompress', context)
	addDeenPlugin('.brotli', 'brotliUncompress', context)
	addDeenPlugin('sha1', 'sha1Hash', context)
	addDeenPlugin('sha224', 'sha224Hash', context)
	addDeenPlugin('sha256', 'sha256Hash', context)
	addDeenPlugin('sha384', 'sha384Hash', context)
	addDeenPlugin('sha512', 'sha512Hash', context)
	addDeenPlugin('sha3-224', 'sha3-244Hash', context)
	addDeenPlugin('sha3-256', 'sha3-256Hash', context)
	addDeenPlugin('sha3-384', 'sha3-384Hash', context)
	addDeenPlugin('sha3-512', 'sha3-512Hash', context)
	addDeenPlugin('md4', 'md4Hash', context)
	addDeenPlugin('md5', 'md5Hash', context)
	addDeenPlugin('ripemd160', 'ripemd160Hash', context)
	addDeenPlugin('blake2s', 'blake2sHash', context)
	addDeenPlugin('blake2b', 'blake2bHash', context)
	addDeenPlugin('blake2x', 'blake2xHash', context)
	addDeenPlugin('blake3', 'blake3Hash', context)
	addDeenPlugin('bcrypt', 'bcryptHash', context)
	addDeenPlugin('scrypt', 'scryptHash', context)
}


function getWebviewContent(content: string) {
	return `<!DOCTYPE html>
  <html lang="en">
  <head>
	  <meta charset="UTF-8">
	  <meta name="viewport" content="width=device-width, initial-scale=1.0">
	  <title>Cat Coding</title>
  </head>
  <body>
	  <pre>${content}</pre>
  </body>
  </html>`;
}

function addDeenPlugin(plugin: string, vscodeCommand: string, context: vscode.ExtensionContext){
	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.'+vscodeCommand, (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin(plugin, content)
		})
	)
}

function runDeenPlugin(plugin: string, content: string) {
	const panel = vscode.window.createWebviewPanel(
		'deen',
		'deen ' + plugin,
		vscode.ViewColumn.One,
		{}
	);
	const cmd = child.execFile('/home/takeshix/bin/godeen', [plugin, content], (err, stdout, stderr) => {
		if (err) {
			throw err
		}
		panel.webview.html = getWebviewContent(stdout)
	});
}

export function deactivate() {}
