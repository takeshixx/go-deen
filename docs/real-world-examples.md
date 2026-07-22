# Real-world encoded data examples

This note tracks public, sanitized examples that map well to `deen` chains, plus
formats that look useful but need more support before they can become runnable
built-in examples.

## Added as built-in examples

### Suspicious URL inspection and redirect extraction

Source pattern:
Spam and phishing links often combine a misleading hostname, several path
segments, duplicate tracking parameters, and a percent-encoded redirect URL.
The `urlparts` formatter separates these components locally without resolving
or contacting the destination. Query parameters remain ordered and duplicate
keys are retained. Its derived analysis also lists nested HTTP(S) URLs, common
tracking parameters, evidence-oriented indicators, and a defanged copy.

Runnable chain:

```sh
printf '%s' 'https://login-update.example.invalid/account/verify.php?campaign=Q3&redirect=https%3A%2F%2Fportal.example.org%2Fsignin&campaign=retry#continue' \
| deen urlparts \
| deen jq -q '.analysis.nested_urls[].url' -no-color
```

Edit a decoded value and rebuild the URL:

```sh
printf '%s' 'https://example.invalid/verify?recipient=alice%40corp.example' \
| deen urlparts \
| deen jq -q '(.query[] | select(.key == "recipient").value) = "bob@corp.example"' -no-color \
| deen .urlparts
```

Reference:

- https://www.rfc-editor.org/rfc/rfc3986

### CloudFront signed URL custom policy

Source pattern:
AWS CloudFront signed URLs carry a `Policy` query parameter containing a JSON
policy statement that is Base64-encoded and then made URL-safe with CloudFront's
own substitutions. The policy controls the resource, expiration window, and
optional source IP range.

Runnable chain:

```sh
deen regex -re '(?s).*[\?&]Policy=([^&]+).*' -group 1 \
| deen regex -re '~' -replace '/' \
| deen regex -re '_' -replace '=' \
| deen regex -re '-' -replace '+' \
| deen .base64 \
| deen json \
| deen jq -q '{resource: .Statement[0].Resource, expires: .Statement[0].Condition.DateLessThan["AWS:EpochTime"], source: .Statement[0].Condition.IpAddress["AWS:SourceIp"]}' -no-color
```

References:

- https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/private-content-creating-signed-url-custom-policy.html

### Kubernetes docker config Secret

Source pattern:
Kubernetes Secret manifests store values under `data` as Base64 strings. Docker
config Secret types commonly contain a Base64-encoded serialized Docker config,
which then needs JSON parsing to inspect registry auth material.

Runnable chain:

```sh
deen regex -re '(?m)^\s+\.dockercfg:\s*\|\s*\n\s+([A-Za-z0-9+/=]+)' -group 1 \
| deen .base64 \
| deen json \
| deen jq -q '.auths["https://example/v1/"].auth' -no-color \
| deen regex -re '"([^"]+)"' -group 1
```

References:

- https://kubernetes.io/docs/concepts/configuration/secret/

### Nested protobuf wire message

Source pattern:
The Protocol Buffers encoding guide shows nested messages as length-delimited
wire fields. Without a `.proto`, `deen` can still expose the field numbers,
wire types, lengths, and nested values for triage.

Runnable chain:

```sh
deen .hex | deen protobuf
```

References:

- https://protobuf.dev/programming-guides/encoding/

### PowerShell EncodedCommand from process logs

Source pattern:
Windows process creation telemetry, including Security Event 4688 and Sysmon
process creation events, can include a full command line. PowerShell
`-EncodedCommand` stores a command as Base64 over UTF-16LE, and encoded
PowerShell commands are common enough in intrusion chains to have Sigma
detection rules.

Runnable chain:

```sh
deen regex -re '(?i)-e(?:ncodedcommand|nc|n|c)?\s+([A-Za-z0-9+/=]+)' -group 1 \
| deen .base64 \
| deen .utf16le \
| deen regex -re 'https?://[^'"'"'"\s)]+'
```

References:

- https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_pwsh
- https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688
- https://learn.microsoft.com/en-us/sysinternals/downloads/sysmon
- https://attack.mitre.org/techniques/T1059/001/
- https://raw.githubusercontent.com/SigmaHQ/sigma/master/rules/windows/process_creation/proc_creation_win_powershell_base64_encoded_cmd_patterns.yml

### JWKS x5c signing certificate

Source pattern:
OpenID Connect and OAuth deployments publish signing keys as JWKS. JWK `x5c`
entries contain Base64-encoded DER X.509 certificates, which analysts often
need to inspect for issuer, subject, validity, key material, and thumbprints.

Runnable chain:

```sh
deen jq -q '.keys[0].x5c[0]' -no-color \
| deen regex -re '"([^"]+)"' -group 1 \
| deen .base64 \
| deen pem -cert \
| deen certPrinter
```

References:

- https://www.rfc-editor.org/rfc/rfc7517

## Good roadmap candidates

### PowerShell deobfuscation assistant

The PowerShell EncodedCommand chain works for the first layer, but incident
payloads commonly add more layers: nested Base64, gzip/deflate byte arrays,
string concatenation, backtick escaping, `[char]` arithmetic, XOR, environment
variable reconstruction, and `FromBase64String()` calls inside the decoded
script.

Useful feature shape:

- Detect `-EncodedCommand` in command-line logs and propose
  Base64 -> UTF-16LE automatically.
- Add a bounded PowerShell string/AST deobfuscation helper that extracts
  literals and common decode calls without executing code.
- Detect byte-array and char-code constructions and convert them to text.
- Keep a clear "static analysis only" boundary.

### CloudFront policy codec option

The CloudFront policy example currently needs three regex replacement steps to
undo `+` to `-`, `=` to `_`, and `/` to `~`. A `base64 -cloudfront` decode
option or a dedicated `cloudfront-policy` formatter would make this much easier
to discover and less error-prone.

### JWE compact serialization

RFC 7516 JWE compact serialization uses five Base64URL components: protected
header, encrypted key, initialization vector, ciphertext, and authentication
tag. `deen` can inspect Base64URL parts manually, and the JWT plugin has partial
JWE encoding flags, but JWE decrypt/inspect is stubbed today.

Useful feature shape:

- Parse compact and JSON JWE serializations.
- Decode protected and unprotected headers without a key.
- Decrypt direct, RSA-OAEP, AES key wrap, and common AES-GCM/CBC-HMAC content
  encryption modes when keys are provided.
- Expose a header-only chain for incident triage when keys are unavailable.

References:

- https://www.rfc-editor.org/rfc/rfc7516

### AWS STS encoded authorization messages

AWS authorization failures can include an encoded message that must be sent to
`sts:DecodeAuthorizationMessage`. The message may contain privileged details
such as principal, action, resource, denial reason, and condition keys, and AWS
requires IAM permission to decode it. This is not just local Base64 decoding.

Useful feature shape:

- Keep this out of normal local transforms unless credentials/network access are
  explicitly configured.
- Add a documentation recipe explaining when `deen` can inspect surrounding log
  JSON locally and when AWS STS is required.
- If networked integrations are added later, gate them behind explicit user
  action and avoid automatic credential discovery by default.

References:

- https://docs.aws.amazon.com/STS/latest/APIReference/API_DecodeAuthorizationMessage.html

### Nested JWT/JWE tokens

The JWT decoder currently returns an explicit error for nested JWT content type
(`cty: JWT`). Real identity provider payloads can contain signed-then-encrypted
or encrypted-then-signed nested tokens.

Useful feature shape:

- Recursive JWT/JWE decode with a depth limit.
- Header-only mode for encrypted inner tokens.
- Clear verification/decryption state in structured output.

### CWT, COSE, and WebAuthn/FIDO artifacts

CBOR Web Token uses CBOR claims protected by COSE, and WebAuthn attestation
objects are CBOR structures containing authenticator data, credential public
keys, attestation statements, and signatures. `deen` can decode raw CBOR today,
but it does not understand COSE headers, COSE keys, CWT claim numbers, WebAuthn
authenticator data, or attestation formats.

Useful feature shape:

- COSE_Sign1, COSE_Mac0, and COSE_Encrypt0 structure inspection.
- CWT claim-name rendering, timestamp conversion, and recursive nested CWT
  display.
- WebAuthn attestationObject/clientDataJSON/authenticatorData parser.
- COSE key to JWK/PEM conversion where possible.

References:

- https://www.rfc-editor.org/rfc/rfc8392
- https://www.rfc-editor.org/rfc/rfc9052
- https://www.w3.org/TR/webauthn-2/

### JWKS and x5c workflow polish

The x5c example is possible with a chain, but it is awkward because the user has
to extract a JSON string, strip quotes, decode Base64 DER, wrap as PEM, and then
print the certificate.

Useful feature shape:

- Add `jwk -x5c` or `jwks -certs` to emit PEM certificates from `x5c` chains.
- Add JWK/JWKS key selection by `kid`, `use`, and `alg`.
- Add x5t/x5t#S256 verification against decoded certificates.

### Windows event and endpoint telemetry containers

Security analysts often begin with EVTX, Sysmon XML, EDR JSON, or Windows event
exports. `deen` can format XML/JSON and extract fields with regex or jq, but it
does not parse binary EVTX files, normalize event data, or provide event-aware
field extraction.

Useful feature shape:

- EVTX parser with record iteration and JSON/XML output.
- Sysmon/Security event helpers for common fields such as CommandLine, Image,
  ParentImage, Hashes, TargetObject, DestinationIp, and QueryName.
- Detect-next suggestions for encoded data found inside event fields.
- Hash-list splitter for Sysmon hash fields such as `SHA256=...`.

References:

- https://learn.microsoft.com/en-us/windows/security/threat-protection/auditing/event-4688
- https://learn.microsoft.com/en-us/sysinternals/downloads/sysmon

### CMS, S/MIME, PKCS#7, and signed binaries

CMS/PKCS#7 containers show up in S/MIME, detached signatures, timestamp
responses, Authenticode-related workflows, and certificate bundles. `deen` has
PEM, ASN.1, and certificate inspection, but no CMS-aware parser.

Useful feature shape:

- Parse SignedData, EnvelopedData, DigestedData, and AuthenticatedData.
- Extract embedded certificates, signer infos, digest algorithms, content type,
  signing time, and detached-content requirements.
- Convert PEM/DER CMS containers and chain into certPrinter/asn1 views.

References:

- https://www.rfc-editor.org/rfc/rfc5652

### Email authentication records and MIME phishing artifacts

Email investigations involve layered text encodings: MIME encoded words,
quoted-printable bodies, Base64 attachments, DKIM signatures, SPF/DMARC DNS TXT
records, ARC headers, and sometimes compressed or password-protected
attachments. `deen` has quoted-printable, Base64, DNS-name, regex, and JSON/XML
building blocks, but not header-aware MIME or DKIM canonicalization.

Useful feature shape:

- MIME message parser with part listing and extraction.
- RFC 2047 encoded-word decoder for headers.
- DKIM header parser with body-hash extraction and canonicalization checks.
- SPF/DMARC record parser and DNS TXT helper.

References:

- https://www.rfc-editor.org/rfc/rfc6376

### Archive and executable carrier chains

Real investigations often start with an email attachment, browser cache blob,
or binary artifact containing another encoded payload. `deen` has magic,
entropy, binary inspection, and compression primitives, but does not yet unpack
container formats such as ZIP, TAR, 7z, PDF object streams, or MIME multipart.

Useful feature shape:

- ZIP/TAR/GZIP member listing and extraction by name/index.
- MIME multipart parser with attachment extraction.
- PDF stream/object extraction and FlateDecode helpers.
- Detect-next expansion that can propose container extraction before codecs.
