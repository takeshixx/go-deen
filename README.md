# deen ![Build & Test](https://github.com/takeshixx/go-deen/workflows/Build%20&%20Test/badge.svg?branch=master)

`deen` is a generic data encoding/decoding application. It provides several interfaces including a CLI, a GUI implemented with [fyne](https://github.com/fyne-io/fyne) and a (experimental) web interface built with [Vecty](https://github.com/hexops/vecty) that runes in [WebAssembly](https://de.wikipedia.org/wiki/WebAssembly). Additional interfaces and ways to interact and include `deen` in different workflows include a [Visual Studio Code extension](https://github.com/takeshixx/go-deen/tree/master/extras/vscode-deen). This project aims to be compatible with most common environments, including Linux, Windows and macOS. The resulting binary is static and should therefore run on all supported platforms without any additional dependencies.

`go-deen` is a Go reimplementation of the old `deen` versions that were initially implemented with Python and PyQt5.

It should be noted that this code is still highly experimental. However, the majority of core plugins is already implemented and functional.

### TODO

Current and future features and TODOs are tracked in various [repository projects](https://github.com/takeshixx/go-deen/projects).

## Building & Running

The following command will create a `deen` binary in the `bin/` folder:

```bash
make
```

Running the binary without any arguments spawns the GUI. The CLI is used when plugin names are supplied. By default, `deen` reads input from the remaining positional arguments or stdin. Input can also be read from a file with the `-file` flag (both globally, before the plugin name, and as a per-plugin flag).

Processing is typically implemented by calling plugin names without a prefix. Depending on the plugin, this will encode, compress, hash or otherwise transform input data.

```bash
$ deen b64 test
dGVzdA==
```

Unprocessing is called by calling plugins with a "." (dot) prefix. Depending on the plugin, this will decode, decompress or otherwise transform input data in reverse. One-way plugin categories like `hashs` do not implement unprocessing and return an error if called with a dot prefix.

```bash
$ deen .b64 dGVzdA==
test
```

### Output and newlines

`deen` writes output **raw** by default — no trailing newline is appended — so that binary output (e.g. compressed data) can be piped without corruption:

```bash
$ echo -n test | deen gzip | deen .gzip
test
```

Pass `-N` (after the plugin name) to append a trailing newline, which is handy when reading output in a terminal:

```bash
$ deen b64 -N test
dGVzdA==
```

Decoders tolerate surrounding whitespace, so a trailing newline from the shell does not break decoding (`echo dGVzdA== | deen .b64` works). Base64 decoding auto-detects the standard, raw, URL and raw-URL alphabets unless an explicit `-strict`, `-url` or `-raw` flag is given.

### Listing plugins

List all available plugins and their aliases, optionally as JSON:

```bash
$ deen -l        # human-readable, grouped by category
$ deen -lj       # JSON (includes aliases)
```

Per-plugin options are shown with `-h` after the plugin name, e.g. `deen base64 -h`.

### GUI

Making a binary with GUI support:

```bash
make gui
```

Running the resulting binary without any CLI arguments will start the GUI.

In order to build the GUI, the following dependencies are required:

**Ubuntu**

```bash
sudo apt update && sudo apt install xorg-dev
```

### WebAssembly

*Note*: WebAssembly code currently resides in the `web` branch.

Build and run the web interface:

```bash
make web
```

This will spawn a local web server on TCP port 9090 (currently requires [http_server.go](https://github.com/takeshixx/tools/blob/master/net/daemons/http_server.go)) that services the web interface.

## Writing plugins

Every plugin is a `*types.DeenPlugin` built by a constructor and registered in [`internal/plugins/plugins.go`](internal/plugins/plugins.go). A plugin implements a single unified contract:

```go
type DeenPlugin struct {
    Name        string
    Aliases     []string
    Category    string
    Description string

    RegisterFlags func(*flag.FlagSet)                          // optional, plugin-specific flags
    Process       func(io.Reader, io.Writer, *flag.FlagSet) error // forward (encode/compress/hash/format)
    Unprocess     func(io.Reader, io.Writer, *flag.FlagSet) error // reverse (decode/decompress); nil = one-way
}
```

`Process` and `Unprocess` read input from an `io.Reader` and write the result to an `io.Writer`, returning on the first error (they must not continue after a failure). Leaving `Unprocess` as `nil` marks the plugin one-way (e.g. hashes). See [`examples/example_plugin.go`](examples/example_plugin.go) for an annotated reference implementation, and the [`pkg/hashs`](pkg/hashs) factory for how families of similar plugins are built with minimal boilerplate.

## Why go-deen?

tl;dr: Because we can.

The original version has several issues/limitations:

* **Dependencies**: it requires various dependencies that are painful to maintain and install on verious operating systems and environments. Thanks to Golang this new version can be compiled and distributed as a single static binary.
* **Performance**: due to the poor implementation in the original version it does not perform properly with large amounts of data. The new version is implemented with stream readers, which also allows to deal with large files. The following shows an example for a 18G VMDK file:
    ```bash
    $ time cat disk001.vmdk | sha256sum
    003b375eeba7e56ae8e1aa03eb2ac6741023478c41fdc077619f144b114e0d02  -
    cat   0.10s user 5.40s system 12% cpu 44.792 total
    sha256sum  41.91s user 2.82s system 99% cpu 44.792 total
    $ time cat disk001.vmdk | deen sha256
    003b375eeba7e56ae8e1aa03eb2ac6741023478c41fdc077619f144b114e0d02
    cat   0.14s user 5.76s system 11% cpu 51.294 total
    bin/deen sha256  48.15s user 3.06s system 99% cpu 51.293 total
    ```
* **GUI**: the original version was implemented with PyQT, which required every system to have QT installed to use the GUI. In Golang there are several projects that allow to create GUIs without any system dependencies. Currently there is only an experimental GUI as the available modules are still evaluated. But the goal is to create deen as a single static binary that also includes a simple GUI without any dependencies.