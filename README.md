# unaware

A command-line utility to mask sensitive data in JSON and XML files.

It replaces data values with non-reversible, consistently hashed equivalents, preserving the original data structure. This is useful for sharing data schemas or structures without exposing personally identifiable information (PII).

## Features

- Masks JSON and XML files.
- **Consistent Mode**: The same input value always produces the same masked output.
- **Random Mode**: Values are replaced with random, non-consistent data.
- Piped I/O: Reads from `stdin` and writes to `stdout` by default.

## Usage

### Build

```shell
go build -o unaware .
```

### Command-Line Flags

- `-in <file>`: Path to input file (defaults to `stdin`).
- `-out <file>`: Path to output file (defaults to `stdout`).
- `-format <type>`: Data format, `json` or `xml` (defaults to `json`).
- `-consistent=false`: Use random masking instead of consistent hashing.

### Examples

#### JSON from file

```shell
./unaware -in sensitive.json -out masked.json
```

#### XML from stdin to stdout

```shell
cat sensitive.xml | ./unaware -format xml > masked.xml
```

## Development

This project uses `devenv` for a reproducible development environment.

### Running Tests

Activate the environment and run the test script:

```shell
devenv test
```

Or use the standard Go command:

```shell
go test ./...
```
