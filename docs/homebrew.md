# Homebrew publishing and maintenance

Reviewed on 2026-07-21 against the repository's current build and release
workflows and Homebrew's current documentation.

## Goal and package model

Ship two packages from the same stable release:

| User-facing package | Homebrew type | Third-party tap command | Eventual official command |
| --- | --- | --- | --- |
| Cross-platform CLI | Formula, built from source and optionally bottled | `brew install takeshixx/tap/deen` | `brew install deen` |
| macOS desktop GUI | Cask containing an upstream-built, signed, notarized `.app` | `brew install --cask takeshixx/tap/deen` | `brew install --cask deen` |

The formula and cask may both use the token `deen`. The formula owns the `deen`
command. The cask installs only `deen.app` and must not also link a `deen` binary,
so users can install the CLI and GUI together without a file conflict.

The CLI formula should support both macOS and Linux. The initial GUI cask is a
macOS distribution. Linux GUI packaging is a separate project because it needs
a supported Linux artifact layout, desktop integration, runtime dependency
decisions, and its own Homebrew validation.

## Important constraints found during review

- A third-party tap does not make `brew install deen` work globally. Current
  Homebrew tap-trust behavior expects the fully qualified name unless the user
  explicitly trusts the tapped item. The documentation should advertise the
  fully qualified tap commands above.
- The untapped commands require acceptance into `homebrew/core` for the formula
  and `homebrew/cask` for the GUI.
- Official repositories require public interest beyond the author. For a
  repository-owner submission, the normal threshold is 225 GitHub stars, 90
  forks, or 90 watchers. As of this review, the GitHub API reports 3 stars, 1
  fork, and 2 subscribers/watchers for `takeshixx/go-deen`. Unless Homebrew
  grants an exception, official submission is therefore a later milestone; the
  personal tap is the immediate distribution route.
- `deen` is currently absent from both the official formula and cask APIs, but
  names and open pull requests must be checked again before submission.
- All existing tags are prereleases. `v3.4.0-beta` is the newest tag, while
  `homebrew/core` requires a stable, immutable, tagged release.
- The current manual release workflow always marks a release as a prerelease,
  builds only CLI binaries, and derives the new tag by running a binary whose
  version includes the Git branch. That is how a value such as
  `v3.4.0-beta-master` can be produced. A source archive used by Homebrew also
  has no Git checkout from which the current Makefile can reliably run
  `git describe`.
- `core.Version()` falls back to `dev`, but the `-version` flag prints the raw
  linker variable instead. A build without linker flags can therefore print a
  blank version. Fix this before the formula test is written.

## Phase 1: make upstream releases packageable

Complete these once in the `go-deen` repository before publishing a tap.

1. Define the release contract.

   - Use stable semantic-version tags such as `v3.4.0`; reserve suffixes such as
     `-beta.1` for prereleases.
   - Choose and document the GUI bundle identifier, for example
     `com.adversec.deen`, plus the displayed app name, minimum supported macOS
     release, copyright, and icon.
   - Keep the CLI and GUI on the same version and source commit.
   - Make a named maintainer responsible for release failures, Homebrew update
     pull requests, and security releases.

2. Fix version injection.

   - Make the build accept an explicit release version from CI instead of
     deriving it from the nearest tag and current branch.
   - Do not append a branch name to stable release builds.
   - Make `deen -version` call the same fallback-aware version accessor used by
     the rest of the application.
   - Allow the formula to inject the formula version directly with `-ldflags`;
     do not make the formula call the Git-dependent Makefile target.
   - Add tests for an unset development version, an explicit stable version,
     and the exact `deen -version` output.

3. Replace the current manual-only release job with a release pipeline.

   - Run the full CLI and GUI tests before producing artifacts.
   - Build from a specific reviewed commit and validate that the requested
     stable version matches the tag format.
   - Create Linux and macOS CLI artifacts for supported architectures. These are
     useful direct downloads even though the Homebrew formula builds from
     source and Homebrew bottles are produced separately.
   - Produce a draft GitHub release, upload every final artifact and checksum,
     and publish it only after all verification passes.
   - Set GitHub's `prerelease` flag from the version rather than hard-coding it.
   - Generate `SHA256SUMS` for all downloadable assets. Never replace an asset
     attached to a published version; issue a patch release when an artifact is
     wrong.

4. Add a real macOS application build.

   - Add a high-resolution PNG icon and Fyne application metadata to the
     repository.
   - Build the GUI with the `gui` tag on native Apple Silicon and Intel runners,
     or combine verified per-architecture binaries into one universal app.
     Separate arm64 and x86_64 ZIP files are acceptable and map cleanly to a
     cask `arch` stanza.
   - Package `deen.app` with Fyne's packaging tooling. Confirm the bundle has the
     expected identifier, version, executable, icon, and minimum macOS version.
   - Decide whether the release GUI includes `webembed`. If it does, generate
     the WebAssembly assets before building and make this choice part of the
     reproducible release command.

5. Sign and notarize the GUI.

   - Enrol the release owner in the Apple Developer Program and create a
     Developer ID Application certificate.
   - Store the certificate, certificate password, Apple team ID, and notarization
     credentials in a protected GitHub release environment. Restrict their use
     to reviewed release refs and require approval for that environment.
   - Enable the hardened runtime and a secure timestamp. Define only the
     entitlements the app actually needs.
   - Sign nested code first and the app bundle last; do not use `codesign --deep`
     as a substitute for correct signing order.
   - Verify with `codesign --verify --strict --verbose=2` and
     `spctl --assess --type execute --verbose`.
   - Submit a ZIP, DMG, or package with `xcrun notarytool`, require an accepted
     result, staple the ticket to the app, and validate it with
     `xcrun stapler validate`.
   - Archive the stapled `deen.app` as the final versioned ZIP and calculate the
     cask checksum from that exact file.

6. Smoke-test the release exactly as users receive it.

   - On macOS arm64, macOS Intel, and Linux, verify the CLI reports the release
     version and performs a real transform.
   - Download each GUI archive from the draft release into a clean macOS test
     environment, confirm its checksum and architecture, install it in
     `/Applications`, and launch it under Gatekeeper.
   - Confirm the CLI and GUI can be installed together and that neither package
     modifies files owned by the other.

## Phase 2: publish and validate the personal tap

1. Create `takeshixx/homebrew-tap` with `brew tap-new takeshixx/tap` and push it
   as a public GitHub repository. Keep the generated GitHub Actions workflows,
   use pull requests for changes, and protect the default branch.

2. Add `Formula/deen.rb` with the following behavior.

   - Set `desc`, the canonical HTTPS `homepage`, the immutable stable source-tag
     `url`, `sha256`, and `license "Apache-2.0"`.
   - Declare `depends_on "go" => :build`.
   - Build `./cmd/deen` with `CGO_ENABLED=0`, `-mod=readonly`, Homebrew's
     `std_go_args` (which supplies `-trimpath` and the output path), stripped
     linker flags, and an explicit `core.version` matching the formula version.
     Leave the branch linker variable empty for a stable build.
   - Do not install one of the upstream release binaries in the official-style
     formula. Homebrew formulae for open-source CLI tools are expected to build
     from a versioned and checksummed source release.
   - Add a meaningful `test do` block that asserts both the version and a real
     transform, for example piping `test` through `deen base64` and expecting
     `dGVzdA==`. A version-only test is insufficient.
   - Add a stable-only `livecheck` rule if the default GitHub strategy can select
     beta tags.

3. Add `Casks/deen.rb`.

   - Set `version`, per-architecture `sha256` values, versioned GitHub release
     URLs, `name`, `desc`, and `homepage`.
   - Use `arch` and architecture-specific URLs/checksums for separate arm64 and
     Intel archives, or one URL/checksum for a verified universal app.
   - Install only `app "deen.app"` and declare the tested minimum macOS version.
   - Point the cask only at upstream-published, signed, notarized artifacts.
   - Add stable-only `livecheck` behavior so prereleases do not replace the
     default cask.

4. Validate every initial package and update.

   Formula validation, on both macOS and Linux:

   ```sh
   brew update
   brew style --fix --formula takeshixx/tap/deen
   brew audit --strict --new --online --formula takeshixx/tap/deen
   HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source --formula takeshixx/tap/deen
   brew test takeshixx/tap/deen
   deen -version
   printf test | deen base64
   brew uninstall --formula takeshixx/tap/deen
   ```

   Cask validation, on Apple Silicon and Intel macOS:

   ```sh
   brew style --fix --cask takeshixx/tap/deen
   brew audit --new --cask takeshixx/tap/deen
   HOMEBREW_NO_INSTALL_FROM_API=1 brew install --cask takeshixx/tap/deen
   codesign --verify --strict --verbose=2 /Applications/deen.app
   spctl --assess --type execute --verbose /Applications/deen.app
   brew uninstall --cask takeshixx/tap/deen
   ```

   Also run `brew lgtm --online` from the tap checkout. Inspect the installed
   files and launch the GUI manually on the latest macOS; a successful command
   exit alone does not validate a desktop app.

5. Make tap CI a launch requirement.

   - Run formula audit, source install, and tests on macOS and Linux.
   - Run cask audit and install/uninstall checks on macOS for every pull request.
   - Exercise both supported macOS architectures before each release, even if
     one is a scheduled or release-only job.
   - Use the workflows generated by `brew tap-new` and `brew test-bot` to build
     CLI bottles. After reviewing the exact pull-request commit and successful
     checks, publish bottles with
     `brew pr-pull --tap=takeshixx/tap --head-sha=<full-commit-sha> <PR-number>`.
   - Treat bottles as strongly recommended for fast installs, but keep the
     source-build path passing because it is the package definition's source of
     truth.

6. Update installation documentation only after CI and clean-machine installs
   pass. Advertise the fully qualified personal-tap commands. Do not claim the
   untapped commands until the official packages have merged.

## Phase 3: move to official Homebrew repositories

Re-evaluate this phase after deen has a stable release history and meets the
current public-interest threshold or has documented evidence for an exception.

1. Re-read the package acceptance policy, acceptable formula and cask rules,
   and contribution guide; Homebrew policy changes over time.
2. Check `brew search deen`, the formula and cask APIs, and open and closed pull
   requests for name conflicts or prior review feedback.
3. Confirm the stable source tag is immutable, the Apache-2.0 license is clear,
   the homepage explains the project, upstream is active, the formula builds on
   Homebrew's complete current macOS/Linux matrix, and the cask works on the
   latest macOS release.
4. Fork and tap `Homebrew/homebrew-core`, create `Formula/d/deen.rb` from the
   proven tap formula, and rerun the new-formula audit, source install, test,
   style, and `brew lgtm --online` checks from an up-to-date checkout.
5. Submit a focused new-formula pull request. Do not add or modify a `bottle do`
   block; Homebrew's CI and maintainers build and publish official bottles.
6. Fork and tap `Homebrew/homebrew-cask`, create the proven `Casks/d/deen.rb`,
   and rerun the new-cask audit, install/uninstall, style, and `brew lgtm`
   checks.
7. Submit a separate new-cask pull request after the signed/notarized artifacts
   are public and immutable.
8. Because this plan and any derived package definitions used Codex assistance,
   disclose AI/LLM use in the initial Homebrew pull-request descriptions, review
   every generated line, and be prepared to address review comments manually as
   required by Homebrew's contribution policy.
9. Respond to CI and maintainer feedback. If either package is rejected for
   notability or another policy reason, keep the personal tap supported and
   record the condition that must change before resubmission.
10. After merges, verify clean installs using exactly `brew install deen` and
    `brew install --cask deen`, then change the main README to make those the
    primary Homebrew commands. Keep the tap formula/cask temporarily only if a
    migration or fallback is useful; avoid leaving duplicate definitions that
    confuse users.

## Release-by-release maintenance

For every stable patch, minor, or major release:

1. Merge and verify all code, dependency, and security changes; update release
   notes and the supported OS/architecture matrix.
2. Run the release pipeline and verify every CLI artifact, GUI signature,
   notarization ticket, checksum, version string, and smoke test before
   publishing the GitHub release.
3. Update the personal tap through a pull request:

   - Formula: source URL/version and source SHA-256.
   - Cask: version, architecture URLs, and each final GUI archive SHA-256.
   - Tests or package metadata when behavior, dependencies, bundle names, or OS
     support changed.

4. Require tap CI to pass, review the exact commit, merge it, publish bottles,
   and test `brew upgrade` as well as a clean install.
5. When the packages are official, watch for Homebrew autobump pull requests.
   Homebrew currently checks eligible official packages automatically; fix or
   submit an update manually with `brew bump-formula-pr` or `brew bump-cask-pr`
   if detection or CI fails. Do not edit official bottle blocks yourself.
6. Update documentation only after packages are available through the intended
   channel. Announce breaking CLI changes, renamed app bundles, new minimum OS
   versions, or architecture removals before users upgrade.

Ongoing maintenance:

- Run scheduled tap audit/install tests so Homebrew, Go, Xcode, macOS, or Linux
  changes are caught between deen releases.
- Monitor upstream Homebrew issues and pull requests, GitHub release download
  failures, and user reports for both packages.
- Renew the Apple Developer membership and signing/notarization credentials
  before expiry, rotate CI secrets, and keep the signing certificate recoverable
  by the release owner.
- Keep dependencies and the `go` directive buildable with Homebrew's current Go
  formula. Cut prompt patch releases for security or compatibility fixes.
- Never overwrite published source tags or release assets to fix a checksum.
  Publish a new patch version and update the formula/cask.
- If support ends, follow Homebrew's deprecation/disable policy and leave users
  a clear migration path rather than silently breaking an existing package.

## Definition of done

Personal-tap publication is complete when:

- a stable GitHub release contains verified CLI artifacts and signed, notarized
  GUI archives for every declared architecture;
- both fully qualified tap commands install on clean supported systems;
- `deen -version`, a real CLI transform, GUI launch, upgrade, uninstall, and
  simultaneous CLI/GUI installation pass;
- tap CI, bottle publication, release ownership, and the per-release update
  procedure are operating and documented.

Official publication is complete only when both Homebrew pull requests are
merged and the untapped CLI and cask commands pass on clean machines.

## Authoritative references

- [Homebrew package acceptance policy](https://docs.brew.sh/Package-Acceptance-Policy)
- [Acceptable formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Acceptable casks](https://docs.brew.sh/Acceptable-Casks)
- [Adding software to Homebrew](https://docs.brew.sh/Adding-Software-to-Homebrew)
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
- [Creating and maintaining a tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Opening a Homebrew pull request](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request)
- [Homebrew autobump](https://docs.brew.sh/Autobump)
- [Fyne desktop packaging](https://docs.fyne.io/started/packaging/)
- [Apple Developer ID distribution](https://developer.apple.com/developer-id/)
- [Apple notarization](https://developer.apple.com/documentation/security/notarizing-macos-software-before-distribution)
