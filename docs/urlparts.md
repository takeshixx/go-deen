# URL Parts JSON schema

The `urlparts` formatter parses a URL into editable JSON. Its reverse form,
`.urlparts`, validates that JSON and rebuilds the URL. Both operations are local:
the plugin does not resolve, open, or fetch the supplied URL.

```sh
printf '%s' 'https://example.invalid/a/b?next=https%3A%2F%2Fportal.example.org#continue' \
| deen urlparts
```

The output uses schema version 1:

```json
{
  "version": 1,
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
  }
}
```

## Field behavior

- `version` identifies the schema. `.urlparts` currently accepts version 1.
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

Decoded fields are authoritative when edited. During reconstruction, a raw
spelling is reused only while it still decodes to the corresponding value. A
changed value is encoded again. If only `path_segments` changes, the path is
rebuilt from those segments; if `path` itself changes, it takes precedence.

## jq examples

Extract a nested redirect URL:

```sh
deen urlparts "$URL" \
| deen jq -q '.query[] | select(.key == "redirect") | .value' -no-color
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

## Desktop GUI editor

In the desktop GUI, a forward `urlparts` step adds a **URL Parts** output tab.
It provides editable scheme, hostname, port, user-information, path-segment,
query-parameter, fragment, and advanced fields. Path and query rows can be
added, duplicated, reordered, or removed. Query values without an equals sign
remain distinguishable from explicitly empty values.

The tab continuously shows the locally rebuilt URL and can copy it without
opening it. Structured edits update the JSON in the **Raw** tab and recompute
downstream pipeline steps. Editing valid JSON in the **Raw** tab refreshes the
structured controls in the other direction. The **Show original encoded
values** control exposes the preserved raw path and query spellings.
