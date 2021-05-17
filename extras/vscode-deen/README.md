# Visual Studio Code Extension for deen

A simple deen frontend for Visual Studio Code.

## Building

Run Docker container with NodeJS:

```bash
docker run --rm -it -v $(pwd):/app node bash
```

Install building dependencies:

```bash
npm install -g typescript vsce
```

Build the extension:

```bash
cd /app
vsce package
```