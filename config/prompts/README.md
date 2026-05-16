# Prompt Truth Source

This directory is the F3 prompt registry truth source. Every baseline prompt
shipped to staging or production is a YAML meta file paired with a Markdown
template body under `<feature_key>/<version>[.<language>].{yaml,md}`.

The `prompt-rubric-registry` Go package, the `scripts/lint/prompt_lint.py`
linter, and the seed migration in `migrations/` all derive their behavior from
the rules in this README. Any change to the field set, hash algorithm, or
language convention must update this README first; the Go loader and Python
lint share this description as their canonical specification.

## 1 File layout

```
config/prompts/
  README.md                                 (this file — algorithm description)
  <feature_key>/
    <version>.yaml                          (language: multi default)
    <version>.md                            (template body for multi)
    <version>.<language>.yaml               (language variant; ISO-639 lowercase)
    <version>.<language>.md                 (template body for that variant)
```

- `<feature_key>` is one of the 11 baseline keys frozen in
  `docs/spec/prompt-rubric-registry/spec.md` §3.1.1.
- `<version>` is a SemVer string; baseline starts at `v0.1.0`.
- `<language>` is either `multi` (omitted from the filename) or an ISO-639
  lowercase code such as `en` or `zh`. Every feature_key ships at least two
  language coordinates including `multi`, so the F3 RegistryClient can
  exercise the D-6 fallback path (`exact language → multi`).
- The Markdown file at the same path holds the prompt template body. Every
  YAML must have a sibling Markdown file with the matching basename.

## 2 YAML meta schema (fixed field order)

```yaml
feature_key: <string>
version: <semver-string>
language: <multi | iso-639>
template_hash: <sha256-hex>
status: <draft | active | deprecated>
created_at: <RFC 3339 timestamp>
```

Rules:

1. **Field order is fixed.** Linters reject any reordering, additions, or
   missing fields. The order matches the
   `prompt_versions` table column order in `migrations/000001_create_baseline.up.sql`.
2. **`feature_key`** must equal the parent directory name and the YAML's
   feature key. Cross-feature reuse is not allowed.
3. **`version`** is a SemVer string of the literal form `vMAJOR.MINOR.PATCH`
   (the leading `v` is part of the version literal, mirroring the filename
   prefix). Baseline ships as `v0.1.0`. The lint gate enforces the regex
   `^v\d+\.\d+\.\d+(-[A-Za-z0-9\.-]+)?(\+[A-Za-z0-9\.-]+)?$`. The same
   `(feature_key, version, language)` triple cannot appear twice — DB-level
   `UNIQUE(feature_key, version, language)` and the lint gate both enforce.
4. **`language`** is `multi` or a lowercase ISO-639 code. The filename
   suffix (`.<language>` segment between version and extension) must match
   this field exactly when present.
5. **`template_hash`** is the result of the canonical hash algorithm in §3.
6. **`status`** is one of `draft`, `active`, `deprecated`. Only `active` rows
   are seeded into the database with `is_active = true`; toggles between
   `is_active` rows are runtime-derived and do not write back into this YAML.
7. **`created_at`** is RFC 3339 (`2026-05-09T12:00:00Z`). Humans write this
   when the file lands.

## 3 Canonical `template_hash` algorithm

The `template_hash` field in YAML must equal the runtime computation:

```
template_hash = sha256_hex(
    template_body_bytes
    || canonical_json_bytes(meta_for_hash)
)
```

Where:

- `template_body_bytes` is the entire Markdown sibling file's contents,
  read as UTF-8, with `\n` line endings.
- `meta_for_hash` is the parsed YAML map with the `template_hash` field
  **removed** (so the hash never references its own value).
- `canonical_json_bytes(meta_for_hash)` is `json.dumps(meta, sort_keys=True,
  ensure_ascii=False, separators=(",", ":"))` encoded as UTF-8 plus a single
  trailing `\n`. The Go loader uses `encoding/json` with the same key sort
  and the same trailing newline; both implementations must agree byte-for-byte.

Concretely, the linter and loader both:

1. Parse the YAML to a map.
2. Drop `template_hash` from the map.
3. Sort keys lexicographically and serialize with the rules above.
4. Read the Markdown sibling file's bytes.
5. Hash the concatenation `body || canonical_meta_json`.
6. Compare against the stored `template_hash` field.

Any drift — body edit without hash bump, hash bump without body edit, key
reordering without canonicalization, or substituting an empty/old/new hash
into the YAML before re-hashing — fails the lint gate.

## 4 Markdown template body conventions

- Encode as UTF-8 with `\n` line endings.
- Use `{{variable_name}}` interpolation markers. Variables share
  vocabulary with the existing targetjob bridge (`{{jd_text}}`, `{{language}}`,
  `{{rubric_dimensions}}`, etc.) so no new templating syntax is introduced.
- Keep the body in the language declared by `language`. The `multi` body
  must be language-neutral: an English-anchored prompt is acceptable when the
  variables themselves carry user content, but the prose surrounding the
  variables must not assume a specific user language.
- Do **not** include the literal stub markers reserved for the linter to
  reject. The exact list of marker tokens lives in
  `scripts/lint/prompt_lint.py` so the README can sit under the same grep
  scan as the bodies it governs. Every baseline body must be runnable text
  that a real provider could execute, even if the 003-grayscale plan later
  replaces the wording.
- Do **not** embed model identifiers, provider names, or secret material in
  the body. The Resolve output supplies the model profile name; the body is
  provider-neutral.

## 5 Multi-language convention

- `<feature_key>/<version>.yaml` is the default (`language: multi`) and is
  paired with `<feature_key>/<version>.md`.
- Variants are `<feature_key>/<version>.<language>.yaml` paired with
  `<feature_key>/<version>.<language>.md`. The filename language tag must
  match the YAML `language` field.
- A feature_key must ship at least two language coordinates including
  `multi`. Phase 5 verifies that every baseline directory contains both
  the `multi` baseline and at least one language variant.
- Resolution at runtime prefers exact-language match, falling back to
  `multi`. If neither the exact language nor `multi` is present, F3
  RegistryClient returns `ErrLanguageUnsupported`.

## 6 Status / lifecycle

- `draft`: file exists in the truth source but is not seeded into the
  database; the lint gate accepts it but Phase 4 seed migration excludes it.
- `active`: file is seeded into the database with `is_active = true`. Only
  one `active` YAML may exist per `(feature_key, language)` pair.
- `deprecated`: file is kept for history but not seeded as `active`. The
  lint gate accepts it; Phase 4 seed migration writes the row with
  `is_active = false` (or excludes it, depending on Phase 4.4 design).

DB-level `is_active` reflects staging/prod runtime state and does not write
back into this YAML. Edits to YAML `status` are pull-request reviewable.

## 7 Forbidden values

The lint gate rejects:

- Retired-module names from earlier product iterations. The exact list lives
  in `scripts/lint/prompt_lint.py` so this README can stay under the same
  recursive grep scan as the bodies it governs.
- Bare hardcoded prompt assignments inside Go business packages — these
  belong in this directory, not in code. The hardcode lint script
  (`scripts/lint/prompt_hardcode_lint.py`) enforces this separately.

## 8 References

- Spec: `docs/spec/prompt-rubric-registry/spec.md` v2.1
- Plan: `docs/spec/prompt-rubric-registry/plans/001-baseline/plan.md`
- DB schema: `migrations/000001_create_baseline.up.sql` (`prompt_versions` table)
- Lint: `scripts/lint/prompt_lint.py`
- Go loader: `backend/internal/ai/registry/loader.go`
- Sibling rubric README: `config/rubrics/README.md`
