# Visual Studio Code Extension for deen

A simple deen frontend for Visual Studio Code.

## Building

```bash
make
```

## Manual Building

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

## Updating Dependencies

To update depdendencies, run the following:

```bash
make update
````

Or manually:

```bash
docker run --rm -it -v $(pwd):/app node bash
cd /app
npm audit
npm audit fix
```