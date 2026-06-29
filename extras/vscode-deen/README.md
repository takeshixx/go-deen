# Visual Studio Code Extension for deen

A simple deen frontend for Visual Studio Code. Commands process the current
selection, or the whole document when nothing is selected, through the local
`deen` executable and open the result in a new editor tab.

By default the extension runs `deen` from `PATH`. Set `deen.binaryPath` in VS
Code settings if the executable lives somewhere else.

## Building

```bash
make
```

## Manual Building

Run a Docker container with Node.js:

```bash
docker run --rm -it -v $(pwd):/app node:22-alpine sh
```

Install dependencies and build the extension:

```bash
cd /app
npm install
npm run compile
npm exec -- @vscode/vsce package
```

## Updating Dependencies

To update dependencies, run the following:

```bash
make update
```

Or manually:

```bash
docker run --rm -it -v $(pwd):/app node:22-alpine sh
cd /app
npm audit
npm audit fix
```
