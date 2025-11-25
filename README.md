## Description
`unaware` is a command-line tool for masking sensitive data within XML and JSON files. It processes data from files or `stdin` and anonymizes specified property values. Masked values mimick the length and appearance of the original data types.

The program is a cross-platform, statically linked binary with no external dependencies. It leverages streaming and concurrency to efficiently process large files entirely offline.

### Installation

Build the binary from source:
```shell
go build -o unaware main.go
```
Alternatively, check the releases page for pre-built binaries.

### Usage
```
Anonymize data in JSON and XML files by replacing values with realistic-looking fakes.

Use the -method hashed option to preserve relationships by ensuring identical input values get the same masked output value. By default every run uses a random salt, use STATIC_SALT=test123 environment variable for consistent masking.

  -cpu int
    	Numbers of cpu cores used (default 4)
  -exclude value
    	Glob pattern to exclude keys from masking (can be specified multiple times)
  -format string
    	The format of the input data (json or xml) (default "json")
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

#### Use a static salt for consistent masking results
```shell
STATIC_SALT=testing123 ./unaware -in source.json
```

### Advanced Filtering

You can combine `-include` and `-exclude` flags for fine-grained control over what gets masked. The logic follows these simple rules:

1.  **Default (No Flags):** Mask everything.
2.  **Using `-exclude` only:** Mask everything *except* fields matching the exclude patterns (a blacklist).
3.  **Using `-include` only:** Mask *only* the fields matching the include patterns (a whitelist).
4.  **Using Both:** First, select only the fields matching the `-include` patterns, and *then* from that selection, remove any fields that match the `-exclude` pattern. **Exclude always wins.**

This allows for combinations. For example, given `data.json`:
```json
{
  "user": {
    "id": "aa1",
    "personal_info": {
      "subscriber": "uuid-123"
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
./unaware -in data.json -include 'user.personal_info.*' -include 'session.ip_address' -exclude 'user.personal_info.name'
```

**Explanation:**
- `-include 'user.personal_info.*'` selects `user.personal_info.name` and `user.personal_info.email` for masking.
- `-include 'session.ip_address'` adds `session.ip_address` to the list.
- `-exclude 'user.personal_info.subscriber'` then removes the name from that list, even though it was included by the wildcard.

**Result:**
```json
{
  "user": {
    "id": "aa1",
    "personal_info": {
      "subscriber": "uuid-123"
      "name": "Burger Iron",
      "email": "kees@friet.nl",
    }
  },
  "session": {
    "ip_address": "238.108.102.226",
    "timestamp": "2025-11-25T10:00:00Z"
  }
}
```
