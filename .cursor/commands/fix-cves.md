---
name: fix-cves
description: Triage and patch open Dependabot CVEs with tested dependency upgrades.
---

Patch open Dependabot CVEs for this repository end-to-end.

## Goal

Resolve security alerts by upgrading vulnerable dependencies to patched versions, preserving behavior, and preparing a clean PR.

## Execution Rules

1. Prefer the smallest safe upgrade (patch/minor first).
2. Avoid unrelated refactors.
3. Keep each logical ecosystem change in a separate commit when practical.
4. Never downgrade dependencies to silence advisories.
5. If a major upgrade is required, call it out clearly with risk notes.

## Steps

1. Discover current vulnerabilities:
   - Use any CVE IDs, Dependabot alert links, or Dependabot PRs already provided in chat.
   - If none were provided, inspect GitHub data available in this environment (for example open Dependabot PRs and available alert metadata).
   - If GitHub alert metadata is unavailable, run ecosystem-native security tooling (for example `npm audit`, `pip-audit`, `govulncheck`, etc.) to identify actionable vulnerabilities in this repo.

2. Build a remediation plan:
   - Group vulnerabilities by ecosystem and manifest/lockfile.
   - Determine minimum patched versions.
   - Identify whether each fix is patch/minor/major and note expected impact.

3. Implement fixes:
   - Upgrade dependencies via the native package manager.
   - Regenerate lockfiles with the same package manager.
   - Keep formatting/style changes out unless required by tooling.

4. Validate:
   - Run targeted tests first, then broader project tests/build checks as appropriate.
   - Re-run security scans to confirm vulnerabilities are no longer reported.

5. Prepare delivery:
   - Summarize each fixed CVE/advisory, old->new version, and affected files.
   - Include validation commands and outcomes.
   - If operating with git write access, commit changes with clear messages and push the branch.
   - If PR tooling is available, open/update a PR with a concise security-focused description.

## Final Response Format

Return:

1. `Fixed`:
   - CVE/advisory -> dependency -> version change.
2. `Validation`:
   - Commands run and pass/fail outcomes.
3. `Risk Notes`:
   - Any major upgrades, deferred items, or remaining vulnerabilities.
4. `PR`:
   - Branch/commit/PR details if created.
