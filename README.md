### Description
`unaware` is a program for anonymizing or masking all XML and JSON property values from stdin or filepath. It tries to mimick the length and appareance of various data types.

### Install
`go build -o unaware cmd/main.go` or see releases

### Flags
- `-format string`: The format of the input data (`json` or `xml`). (default: `json`)
- `-in string`: Input file path. (default: `stdin`)
- `-out string`: Output file path. (default: `stdout`)
- `-random-hash`: Hash data with random salt (default: `false`)

### Examples

#### JSON from file 

```shell
./unaware -in source.json -out anonymized.json
```

#### JSON from clipboard

```shell
wl-paste | ./unaware 
```


#### XML from stdin to stdout with hashed data

```shell
cat source.xml | ./unaware -format xml -random-hash > masked.xml
```

