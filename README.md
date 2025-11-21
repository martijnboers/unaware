## Description
`unaware` is a command-line tool for masking sensitive data within XML and JSON files. It processes data from files or `stdin` and anonymizes all property values while mimicking the length and appearance of the original data types.

It is a cross-platform, statically linked binary with no external dependencies. It leverages streaming and concurrency to efficiently process large files entirely offline.

### Installation

Build the binary from the source:
```shell
go build -o unaware main.go
```
Alternatively, check the releases page for pre-built binaries.

### Examples

#### JSON from a file
```shell
./unaware -in source.json -out anonymized.json
```

#### XML from stdin with hashed (deterministic) masking
```shell
cat source.xml | ./unaware -format xml -method hashed > masked.xml
```

#### Use a static salt for consistent masking results
```shell
STATIC_SALT=testing123 ./unaware -in source.json
```

#### JSON from the clipboard into `jq`
```shell
wl-paste | ./unaware | jq '.[0:3]'
```
