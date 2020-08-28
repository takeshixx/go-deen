import * as vscode from 'vscode';
import * as child from 'child_process';

export function activate(context: vscode.ExtensionContext) {

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base64Encode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('b64', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base64Decode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.b64', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base32Encode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('b32', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base32Decode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.b32', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base85Encode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('b85', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.base85Decode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.b85', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.urlEncode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('url', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.urlDecode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.url', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.htmlEncode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('html', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.htmlDecode', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.html', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.jsonFormat', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('json', content)
		})
	)

	context.subscriptions.push(
		vscode.commands.registerTextEditorCommand('deen.jsonUnformat', (editor: vscode.TextEditor, edit: vscode.TextEditorEdit) => {
			let content = editor.document.getText()
			runDeenPlugin('.json', content)
		})
	)
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
