# deen ![Build & Test](https://github.com/takeshixx/go-deen/workflows/Build%20&%20Test/badge.svg?branch=master)

`deen` is a generic data encoding/decoding application. It provides several interfaces including a CLI, a GUI implemented with [fyne](https://github.com/fyne-io/fyne) and a (experimental) web interface built with [Vecty](https://github.com/hexops/vecty) that runes in [WebAssembly](https://de.wikipedia.org/wiki/WebAssembly). Additional interfaces and ways to interact and include `deen` in different workflows include a [Visual Studio Code extension](https://github.com/takeshixx/go-deen/tree/master/extras/vscode-deen). This project aims to be compatible with most common environments, including Linux, Windows and macOS. The resulting binary is static and should therefore run on all supported platforms without any additional dependencies.

`go-deen` is a Go reimplementation of the old `deen` versions that were initially implemented with Python and PyQt5.

It should be noted that this code is still highly experimental. However, the majority of core plugins is already implemented and functional.

## Building & Running

The following command will create a `deen` binary in the `bin/` folder:

```bash
make
```

Running the binary without any arguments spawns the GUI. The CLI is used when plugin names are supplied. By default, `deen` reads input from the first positional argument or stdin, but plugins also support the `-file` flag to read input data directly from files.

Processing is typically implemented by calling plugin names without a prefix. Depending on the plugin, this will encode, compress, hash or otherwise transform input data.

```bash
$ deen b64 test
dGVzdA==
```

Unprocessing is called by calling plugins with a "." (dot) prefix. Depending on the plugin, this will decode, decompress or otherwise transform input data in reverse. Therefore, plugin categories like `hashs` to not implement unprocessing functions.

```bash
$ deen .b64 dGVzdA==
test
```

## TODO

Current and future features and TODOs are tracked in various [repository projects](https://github.com/takeshixx/go-deen/projects).

## WHY?

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