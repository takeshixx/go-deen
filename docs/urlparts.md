# URL Parts JSON schema

The `urlparts` formatter parses a URL into editable JSON. Its reverse form,
`.urlparts`, validates that JSON and rebuilds the URL. Both operations are local:
the plugin does not resolve, open, or fetch the supplied URL.

```sh
printf '%s' 'https://example.invalid/a/b?next=https%3A%2F%2Fportal.example.org#continue' \
| deen urlparts
```

New output uses schema version 2. The decoder continues to accept version 1
documents and upgrades them the next time they are encoded:

```json
{
  "version": 2,
  "scheme": "https",
  "userinfo": null,
  "hostname": "example.invalid",
  "port": "",
  "opaque": "",
  "omit_host": false,
  "path": "/a/b",
  "path_segments": ["a", "b"],
  "query": [
    {
      "key": "next",
      "value": "https://portal.example.org",
      "has_value": true,
      "raw_key": "next",
      "raw_value": "https%3A%2F%2Fportal.example.org"
    }
  ],
  "force_query": false,
  "fragment": "continue",
  "raw": {
    "host": "example.invalid",
    "path": "/a/b",
    "fragment": "continue"
  },
  "analysis": {
    "indicators": [
      {
        "code": "nested_url",
        "severity": "info",
        "message": "Found 1 nested HTTP(S) URL(s) in query values."
      }
    ],
    "nested_urls": [
      {
        "parameter_index": 1,
        "key": "next",
        "url": "https://portal.example.org",
        "scheme": "https",
        "hostname": "portal.example.org"
      }
    ],
    "tracking_parameters": [],
    "defanged_url": "hxxps://example[.]invalid/a/b?next=https%3A%2F%2Fportal.example.org#continue"
  }
}
```

## Field behavior

- `version` identifies the schema. `.urlparts` emits version 2 and accepts
  versions 1 and 2.
- `scheme`, `userinfo`, `hostname`, and `port` describe the authority. A null
  `userinfo` means the URL contains no username. `password_set` inside a
  non-null `userinfo` distinguishes no password from an explicitly empty one.
- `opaque` carries the encoded opaque portion of URLs such as `mailto:` URLs.
- `omit_host` preserves the distinction between an omitted and an explicitly
  empty host.
- `path` is the decoded complete path. `path_segments` exposes its individual
  encoded-path segments as decoded strings, so an encoded slash can remain
  inside one segment.
- `query` is an ordered array. It intentionally does not use a JSON object,
  because URLs may contain duplicate keys. `has_value` distinguishes `flag`
  from `flag=`.
- `force_query` preserves a trailing `?` when there are no parameters.
- `fragment` is the decoded fragment.
- `raw` and the `raw_key`/`raw_value` fields retain the original encoded
  spelling. They preserve evidence such as `%2F` versus `/`, `%20` versus `+`,
  and hexadecimal escape casing.
- `analysis` is derived locally from the editable fields. It contains
  evidence-oriented indicators, nested HTTP(S) URLs extracted from decoded
  query values, common analytics/campaign parameter occurrences, and a
  defanged form for pasting into tickets or chat. It is informational, is not a
  malicious/benign verdict, and is ignored during URL reconstruction.

Decoded fields are authoritative when edited. During reconstruction, a raw
spelling is reused only while it still decodes to the corresponding value. A
changed value is encoded again. If only `path_segments` changes, the path is
rebuilt from those segments; if `path` itself changes, it takes precedence.

## jq examples

Extract a nested redirect URL:

```sh
deen urlparts "$URL" \
| deen jq -q '.analysis.nested_urls[].url' -no-color
```

List common tracking parameter occurrences:

```sh
deen urlparts "$URL" \
| deen jq -q '.analysis.tracking_parameters[] | {parameter_index, key}' -no-color
```

Produce a defanged copy without opening or resolving the URL:

```sh
deen urlparts "$URL" \
| deen jq -q '.analysis.defanged_url' -no-color
```

Change a query value and rebuild the URL:

```sh
deen urlparts "$URL" \
| deen jq -q '(.query[] | select(.key == "recipient").value) = "bob@example.org"' -no-color \
| deen .urlparts
```

URLs can contain credentials, session tokens, email addresses, and other
sensitive values. Treat the JSON output and any saved pipeline source with the
same care as the original URL.

## Graphical editors

In the desktop and WebAssembly GUIs, a forward `urlparts` step adds a **URL
Parts** output tab. It provides editable scheme, hostname, port,
user-information, path-segment, query-parameter, fragment, and advanced fields.
Path and query rows can be added, duplicated, reordered, or removed. Query
values without an equals sign remain distinguishable from explicitly empty
values.

The tab continuously shows the locally rebuilt URL and can copy it without
opening it. Structured edits update the JSON in the **Raw** tab and recompute
downstream pipeline steps. Editing valid JSON in the **Raw** tab refreshes the
structured controls in the other direction. The **Show original encoded
values** control exposes the preserved raw path and query spellings. The
**Local analysis** section lists derived indicators, nested URLs, tracking
parameters, and a copyable defanged URL; it never performs network access.
