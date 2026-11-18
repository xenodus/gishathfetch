# Skill: fix-cves

## Purpose

Resolve known vulnerabilities with the smallest safe dependency updates while preserving runtime behavior.

## Inputs

- CVE IDs, advisory links, or Dependabot PRs (if provided)
- Affected manifests and lock/vendor files
- Current test/build commands from repository docs

## Guardrails

1. Prefer patch or minor upgrades before major upgrades.
2. Keep fixes focused on vulnerable dependencies; avoid unrelated refactors.
3. Do not downgrade dependencies to suppress alerts.
4. Apply repository priorities when trade-offs appear:
   - security > correctness > data integrity > performance > clean code

## Procedure

1. Identify actionable vulnerabilities by ecosystem:
   - Use provided advisory context first.
   - If needed, run ecosystem-native security checks.
2. Map each vulnerability to minimum patched versions.
3. Upgrade dependencies with the package manager.
4. Regenerate lock/vendor artifacts with official tooling.
5. Check the Lambda runtime base in `Dockerfile` (final stage):
   - Image: `gcr.io/distroless/static-debian13` with an `@sha256:` digest pin (not a floating tag).
   - Compare the pin to current `:latest` (for example `crane digest gcr.io/distroless/static-debian13:latest` or `docker buildx imagetools inspect gcr.io/distroless/static-debian13:latest`).
   - Keep the **debian13** variant unless deliberately changing runtime OS; refresh the digest when it lags `:latest`.
   - Note old → new digest in the PR summary when updated.
6. Review **GitHub Actions** pins in `.github/workflows/*` (currently `codeql-analysis.yml`, `deploy-on-merge.yml`):
   - Third-party `uses:` refs in this repo include:
     - `actions/checkout`
     - `actions/setup-go` (must stay compatible with `go-version-file: api/go.mod`)
     - `github/codeql-action/init`, `autobuild`, `analyze` (keep all three on the same major, e.g. `@v4`)
     - `oven-sh/setup-bun` (version should match `frontend/package.json` `packageManager` when possible)
     - `docker/setup-buildx-action`
     - `aws-actions/configure-aws-credentials`
   - Compare each action’s pinned major tag to the latest release on GitHub (for example `gh api repos/actions/checkout/tags --jq '.[0].name'`).
   - Bump stale **major** versions when release notes show no breaking input changes for this repo; note old → new in the PR.
   - Leave intentional pins alone (for example `runs-on: ubuntu-24.04` until validated on a newer runner image).
7. Run targeted tests, then broader repository tests.
8. Re-run security checks to confirm remediation.
9. Summarize:
   - advisory/CVE -> dependency -> old/new version
   - validation commands and outcomes
   - risk notes for major upgrades

## Validation Baseline

- `make test`
- Any additional ecosystem checks used for CVE detection/remediation
