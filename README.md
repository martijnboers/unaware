## Description
`unaware` is a command-line tool for masking sensitive data within JSON, XML, and CSV files. It processes data from files or `stdin` and anonymizes specified property values. Masked values mimick the length and appearance of the original data types.

The program is a cross-platform, statically linked binary with no external dependencies. It leverages streaming and concurrency to efficiently process large files entirely offline.

### Installation

Build the program from source:
```shell
go build -o unaware main.go
```
Alternatively, check the releases page for pre-built binaries.

### Usage
```
Anonymize data in JSON, XML, and CSV files by replacing values with realistic-looking alternatives.

Use the -method hashed option to preserve relationships by ensuring identical
input values get the same masked output value. By default every run uses a
random salt, use STATIC_SALT=test123 environment variable for consistent
masking.

  -cpu int
    	Numbers of cpu cores used (default 4)
  -exclude value
    	Glob pattern to exclude keys from masking (can be specified multiple times)
  -format string
    	The format of the input data (json, xml, csv or text) (default "json")
  -in string
    	Input file path (default: stdin)
  -include value
    	Glob pattern to include keys for masking (can be specified multiple times)
  -method string
    	Method of masking (random or hashed) (default "random")
  -out string
    	Output file path (default: stdout)
```

### Examples

#### JSON from a file
```shell
./unaware -in source.json -out anonymized.json
```

#### XML from stdin with hashed (deterministic) masking
```shell
cat source.xml | ./unaware -format xml -method hashed > masked.xml
```

### Filtering

You can control which fields are masked using the `-include` and `-exclude` flags, which both accept glob patterns (e.g., `user.*`, `session.ip_*`).

This allows for precise control. For example, given `data.json`:
```json
{
  "user": {
    "id": "aa1",
    "personal_info": {
      "subscriber": "uuid-123",
      "name": "Jane Doe",
      "email": "jane.doe@example.com"
    }
  },
  "session": {
    "ip_address": "198.51.100.22",
    "timestamp": "2025-11-25T10:00:00Z"
  }
}
```

**Goal:** Mask all sensitive user details and the session IP address, but leave the subscriber field untouched for reference.

**Command:**
```shell
./unaware -in data.json -include 'user.personal_info.*' -include 'session.ip_address' -exclude 'user.personal_info.subscriber'
```

**Result:**
```json
{
  "user": {
    "id": "aa1",
    "personal_info": {
      "subscriber": "uuid-123",
      "name": "Burger Iron",
      "email": "kees@friet.nl"
    }
  },
  "session": {
    "ip_address": "238.108.102.226",
    "timestamp": "2025-11-25T10:00:00Z"
  }
}
```
