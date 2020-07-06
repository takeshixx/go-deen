# deen

Decoder/Encoder application for arbitrary input data. Version 3 is a reimplementation of deen in Golang.

## Building

The following command will create a `deen` binary in the `bin/` folder:

```bash
make
```

## Usage

List available plugins:

```bash
$ deen -l
```

Encode data:

```bash
$ deen b64 test
dGVzdA==
```

Decode data:

```bash
$ deen .b64 dGVzdA==
test
```

Encode from stdin:

```bash
echo -n test | deen b64
dGVzdA==
```

Decode from stdin:

```bash
echo -n dGVzdA== | deen .b64
test
```

Plugin help:

```bash
$ deen flate -h 
Usage of flate:

Implements the DEFLATE compressed data format (RFC1951).

  -level int
        compression level
          No compression:       0
          Best speed:           1
          Best compression:     9
          Default compression:  -1
          (default -1)
```

## TODO

* Port additional plugins from previous version
  * X509 certificates
  * X509 certificate cloning
  * JWT parsing/creation
* GUI: Currently there is only an experimental GUI implementation that is not fully functional.
* Add `--web`: Maybe add a frontend that can be used with a browser?

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