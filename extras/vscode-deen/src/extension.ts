import * as child from "child_process";
import * as vscode from "vscode";

type DeenCommand = {
  command: string;
  title: string;
  plugin: string;
  language?: string;
};

const commands: DeenCommand[] = [
  { command: "base32Encode", title: "Base32 Encode", plugin: "base32" },
  { command: "base32Decode", title: "Base32 Decode", plugin: ".base32" },
  { command: "base64Encode", title: "Base64 Encode", plugin: "base64" },
  { command: "base64Decode", title: "Base64 Decode", plugin: ".base64" },
  { command: "base85Encode", title: "Base85 Encode", plugin: "base85" },
  { command: "base85Decode", title: "Base85 Decode", plugin: ".base85" },
  { command: "urlEncode", title: "URL Encode", plugin: "url" },
  { command: "urlDecode", title: "URL Decode", plugin: ".url" },
  { command: "htmlEncode", title: "HTML Encode", plugin: "html", language: "html" },
  { command: "htmlDecode", title: "HTML Decode", plugin: ".html" },
  { command: "asciiConvert", title: "ASCII Convert", plugin: "ascii" },
  { command: "hexEncode", title: "Hex Encode", plugin: "hex" },
  { command: "hexDecode", title: "Hex Decode", plugin: ".hex" },
  { command: "pemEncode", title: "PEM Encode", plugin: "pem" },
  { command: "pemDecode", title: "PEM Decode", plugin: ".pem" },
  { command: "quotedPrintableEncode", title: "Quoted-Printable Encode", plugin: "quoted-printable" },
  { command: "quotedPrintableDecode", title: "Quoted-Printable Decode", plugin: ".quoted-printable" },
  { command: "rot13Transform", title: "ROT13 Transform", plugin: "rot13" },
  { command: "jsonFormat", title: "JSON Format", plugin: "json", language: "json" },
  { command: "jsonUnformat", title: "JSON Unformat", plugin: ".json", language: "json" },
  { command: "xmlFormat", title: "XML Format", plugin: "xml", language: "xml" },
  { command: "xmlUnformat", title: "XML Unformat", plugin: ".xml", language: "xml" },
  { command: "yamlFormat", title: "YAML Format", plugin: "yaml", language: "yaml" },
  { command: "yamlUnformat", title: "YAML Unformat", plugin: ".yaml", language: "yaml" },
  { command: "tomlFormat", title: "TOML Format", plugin: "toml" },
  { command: "protobufInspect", title: "Protocol Buffers Inspect", plugin: "protobuf" },
  { command: "msgpackDecode", title: "MessagePack Decode", plugin: ".msgpack", language: "json" },
  { command: "cborDecode", title: "CBOR Decode", plugin: ".cbor", language: "json" },
  { command: "samlDecode", title: "SAML Decode", plugin: ".saml", language: "xml" },
  { command: "flateCompress", title: "flate Compress", plugin: "flate" },
  { command: "flateUncompress", title: "flate Uncompress", plugin: ".flate" },
  { command: "lzmaCompress", title: "lzma Compress", plugin: "lzma" },
  { command: "lzmaUncompress", title: "lzma Uncompress", plugin: ".lzma" },
  { command: "lzma2Compress", title: "lzma2 Compress", plugin: "lzma2" },
  { command: "lzma2Uncompress", title: "lzma2 Uncompress", plugin: ".lzma2" },
  { command: "lzwCompress", title: "lzw Compress", plugin: "lzw" },
  { command: "lzwUncompress", title: "lzw Uncompress", plugin: ".lzw" },
  { command: "gzipCompress", title: "gzip Compress", plugin: "gzip" },
  { command: "gzipUncompress", title: "gzip Uncompress", plugin: ".gzip" },
  { command: "zlibCompress", title: "zlib Compress", plugin: "zlib" },
  { command: "zlibUncompress", title: "zlib Uncompress", plugin: ".zlib" },
  { command: "bzip2Compress", title: "bzip2 Compress", plugin: "bzip2" },
  { command: "bzip2Uncompress", title: "bzip2 Uncompress", plugin: ".bzip2" },
  { command: "brotliCompress", title: "brotli Compress", plugin: "brotli" },
  { command: "brotliUncompress", title: "brotli Uncompress", plugin: ".brotli" },
  { command: "zstdCompress", title: "zstd Compress", plugin: "zstd" },
  { command: "zstdUncompress", title: "zstd Uncompress", plugin: ".zstd" },
  { command: "sha1Hash", title: "SHA1 Hash", plugin: "sha1" },
  { command: "sha224Hash", title: "SHA224 Hash", plugin: "sha224" },
  { command: "sha256Hash", title: "SHA256 Hash", plugin: "sha256" },
  { command: "sha384Hash", title: "SHA384 Hash", plugin: "sha384" },
  { command: "sha512Hash", title: "SHA512 Hash", plugin: "sha512" },
  { command: "sha512-224Hash", title: "SHA512-224 Hash", plugin: "sha512-224" },
  { command: "sha512-256Hash", title: "SHA512-256 Hash", plugin: "sha512-256" },
  { command: "sha3-224Hash", title: "SHA3-224 Hash", plugin: "sha3-224" },
  { command: "sha3-256Hash", title: "SHA3-256 Hash", plugin: "sha3-256" },
  { command: "sha3-384Hash", title: "SHA3-384 Hash", plugin: "sha3-384" },
  { command: "sha3-512Hash", title: "SHA3-512 Hash", plugin: "sha3-512" },
  { command: "md4Hash", title: "MD4 Hash", plugin: "md4" },
  { command: "md5Hash", title: "MD5 Hash", plugin: "md5" },
  { command: "ripemd160Hash", title: "RIPEMD160 Hash", plugin: "ripemd160" },
  { command: "blake2sHash", title: "BLAKE2s Hash", plugin: "blake2s" },
  { command: "blake2bHash", title: "BLAKE2b Hash", plugin: "blake2b" },
  { command: "blake2xHash", title: "BLAKE2x Hash", plugin: "blake2x" },
  { command: "blake3Hash", title: "BLAKE3 Hash", plugin: "blake3" },
  { command: "bcryptHash", title: "bcrypt Hash", plugin: "bcrypt" },
  { command: "scryptHash", title: "scrypt Hash", plugin: "scrypt" },
  { command: "adler32Hash", title: "Adler-32 Hash", plugin: "adler32" },
  { command: "crc32Hash", title: "CRC-32 Hash", plugin: "crc32" },
  { command: "crc32cHash", title: "CRC-32C Hash", plugin: "crc32c" },
  { command: "crc32kHash", title: "CRC-32K Hash", plugin: "crc32k" },
  { command: "crc64Hash", title: "CRC-64 Hash", plugin: "crc64" },
  { command: "crc64EcmaHash", title: "CRC-64 ECMA Hash", plugin: "crc64-ecma" },
  { command: "fnv32Hash", title: "FNV-32 Hash", plugin: "fnv32" },
  { command: "fnv32aHash", title: "FNV-32a Hash", plugin: "fnv32a" },
  { command: "fnv64Hash", title: "FNV-64 Hash", plugin: "fnv64" },
  { command: "fnv64aHash", title: "FNV-64a Hash", plugin: "fnv64a" },
  { command: "fnv128Hash", title: "FNV-128 Hash", plugin: "fnv128" },
  { command: "fnv128aHash", title: "FNV-128a Hash", plugin: "fnv128a" },
  { command: "entropyAnalyze", title: "Entropy Analyze", plugin: "entropy" },
  { command: "magicDetect", title: "Magic Detect", plugin: "magic" },
];

let outputChannel: vscode.OutputChannel;

function selectedOrDocumentText(editor: vscode.TextEditor): string {
  const selection = editor.selection;
  if (!selection.isEmpty) {
    return editor.document.getText(selection);
  }
  return editor.document.getText();
}

function deenBinaryPath(): string {
  return vscode.workspace.getConfiguration("deen").get("binaryPath", "deen");
}

function runDeenPlugin(plugin: string, content: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const proc = child.spawn(deenBinaryPath(), [plugin], {
      stdio: ["pipe", "pipe", "pipe"],
    });
    const stdout: Buffer[] = [];
    const stderr: Buffer[] = [];

    proc.stdout.on("data", (chunk: Buffer) => stdout.push(chunk));
    proc.stderr.on("data", (chunk: Buffer) => stderr.push(chunk));
    proc.on("error", reject);
    proc.on("close", (code) => {
      const output = Buffer.concat(stdout).toString("utf8");
      const errorOutput = Buffer.concat(stderr).toString("utf8");
      if (code === 0) {
        resolve(output);
        return;
      }
      reject(new Error(errorOutput.trim() || `deen exited with code ${code ?? "unknown"}`));
    });

    proc.stdin.end(content);
  });
}

async function showResult(command: DeenCommand, result: string): Promise<void> {
  const document = await vscode.workspace.openTextDocument({
    content: result,
    language: command.language,
  });
  await vscode.window.showTextDocument(document, vscode.ViewColumn.Beside);
}

async function runCommand(command: DeenCommand, editor: vscode.TextEditor): Promise<void> {
  const input = selectedOrDocumentText(editor);
  try {
    const result = await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: `deen: ${command.title}`,
        cancellable: false,
      },
      () => runDeenPlugin(command.plugin, input),
    );
    await showResult(command, result);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    outputChannel.appendLine(`deen ${command.plugin} failed:`);
    outputChannel.appendLine(message);
    outputChannel.show(true);
    void vscode.window.showErrorMessage(`deen ${command.plugin} failed: ${message}`);
  }
}

function registerCommand(context: vscode.ExtensionContext, command: DeenCommand): void {
  context.subscriptions.push(
    vscode.commands.registerTextEditorCommand(`deen.${command.command}`, (editor) => {
      void runCommand(command, editor);
    }),
  );
}

export function activate(context: vscode.ExtensionContext): void {
  outputChannel = vscode.window.createOutputChannel("deen");
  context.subscriptions.push(outputChannel);
  for (const command of commands) {
    registerCommand(context, command);
  }
}

export function deactivate(): void {}
