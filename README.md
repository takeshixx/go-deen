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