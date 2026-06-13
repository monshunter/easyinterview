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
    <version>.schema.json                    (language-independent output schema)
    <version>.yaml                          (language: multi default)
    <version>.md                            (template body for multi)
    <version>.<language>.yaml               (optional semantic language override)
    <version>.<language>.md                 (template body for that override)
```

- `<feature_key>` is one of the 13 baseline keys frozen in
  `docs/spec/prompt-rubric-registry/spec.md` §3.1.1.
- `<version>` is a SemVer string; baseline starts at `v0.1.0`.
- `<language>` is either `multi` (omitted from the filename) or an ISO-639
  lowercase code such as `en` or `zh`. Baseline feature keys ship only the
  canonical `multi` coordinate. Add a concrete language override only when
  the task semantics genuinely differ by language or region; do not duplicate
  `multi` just to create an English or UI-language copy.
- The Markdown file at the same path holds the prompt template body. Every
  YAML must have a sibling Markdown file with the matching basename.
- `<version>.schema.json` is language-independent. There is exactly one output
  schema per `(feature_key, version)` for JSON-producing chat feature keys.
  It is shared by `multi` and every optional language override and never has
  a language suffix.

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

## 4 Output schema truth source

`config/prompts/<feature_key>/<version>.schema.json` is the only field-level
truth source for the model output shape. Prompt bodies may explain task context
and business rules, but they must not maintain an independent hand-written
field list that can drift from the schema.

Rules:

1. **Language-independent**: JSON keys and structure do not vary by prompt
   language. The same schema applies to `multi` and any later language
   override for the same `(feature_key, version)`.
2. **Allowed validation subset**: schemas may use only `type`, `required`,
   `properties`, `items`, and `enum`. `description` is allowed as a
   non-validation annotation for rendered prompt text and reviewer context.
   Do not use provider-specific structured-output keys, `$ref`, `oneOf`,
   `anyOf`, `additionalProperties`, `format`, `minimum`, `maximum`, or SDK
   private fields.
3. **Top-level shape**: chat feature keys use top-level `type: object`.
   (The retired jd_match feature keys were the only top-level `array`
   schemas; they were removed with the module per product-scope v2.1 D-17.)
   Voice / STT / TTS feature keys do not produce JSON content and must not
   have output schema files.
4. **Required fields**: `required` contains only fields the backend parser or
   struct actually depends on. A prompt may not require fields the backend does
   not consume unless those fields are optional in schema and their
   `description` states the evaluation value.
5. **Struct / parser alignment**: schema field names must match the backend
   consumer's json tags or explicitly documented parser-required keys. For
   `json.RawMessage` consumers, alignment is against the keys the parser checks
   before persisting the raw payload.
6. **Alias policy**: parser aliases kept for historical outputs are
   compatibility behavior, not new prompt contract fields. For example, a
   parser may accept legacy aliases while the schema-rendered prompt block uses
   only the canonical key.
7. **Template hash boundary**: output schema bytes do not participate in
   `template_hash`. Prompt body edits still require YAML hash refresh; schema
   edits are validated by schema/prompt/struct lint gates instead.
8. **Provider-neutral boundary**: schemas describe business output only. They
   must not include model names, provider names, endpoints, `response_format`,
   `json_schema`, tool-call SDK shapes, or secret material.

## 5 Markdown template body conventions

- Encode as UTF-8 with `\n` line endings.
- Use `{{variable_name}}` interpolation markers. Variables share
  vocabulary with the existing targetjob bridge (`{{jd_text}}`, `{{language}}`,
  `{{rubric_dimensions}}`, etc.) so no new templating syntax is introduced.
- Keep the body aligned to the language coordinate. The `multi` body must be
  language-neutral and must include an explicit output-language instruction
  through `{{language}}` or an equivalent runtime variable. English-anchored
  instruction prose is acceptable when the model is told to answer in the
  requested language; do not create an `en` variant only to replace that
  instruction with "Respond in English."
- Do **not** include the literal stub markers reserved for the linter to
  reject. The exact list of marker tokens lives in
  `scripts/lint/prompt_lint.py` so the README can sit under the same grep
  scan as the bodies it governs. Every baseline body must be runnable text
  that a real provider could execute, even if the 005-grayscale-and-quality-feedback
  plan later replaces the wording.
- Do **not** embed model identifiers, provider names, or secret material in
  the body. The Resolve output supplies the model profile name; the body is
  provider-neutral.

## 6 Prompt output contract block

Every Markdown body for a JSON-producing chat feature key must contain an
output contract block that can be rendered from the schema, or is byte-for-byte
verified by lint against the schema renderer. Humans edit the schema and task
context, not a second copy of the field table inside each language body.

The rendered block contract is:

- Field order is stable: required fields first in schema order, then optional
  properties in schema order.
- Field descriptions come from schema `description` values. Required fields
  and important optional fields must have descriptions so the rendered prompt
  remains readable.
- Examples are generated from schema into complete representative JSON output:
  every schema-declared required and optional property is included, values are
  business-shaped examples rather than `string` / `1` placeholders, and the
  block explicitly says the model must return a JSON value rather than JSON
  Schema or an OpenAPI schema. The example must parse as JSON and pass the same
  schema subset validator used by lint.
- `multi` and optional language overrides render the same JSON keys and
  structure. Only surrounding task prose may differ, and only when that
  difference carries real task semantics.
- Manual edits that add, remove, rename, or reorder output keys in the prompt
  contract block must make `make lint-prompts` fail until the schema or prompt
  block is corrected.

## 7 Language coordinate convention

- `<feature_key>/<version>.yaml` is the canonical baseline (`language:
  multi`) and is paired with `<feature_key>/<version>.md`.
- Optional overrides are `<feature_key>/<version>.<language>.yaml` paired with
  `<feature_key>/<version>.<language>.md`. The filename language tag must
  match the YAML `language` field.
- Baseline feature keys must ship the `multi` coordinate. They must not ship
  duplicate `en`, `zh`, or other language copies unless a spec/plan records
  the semantic reason for that override. Examples of valid reasons include
  region-specific legal language, localized interview norms, or an explicit
  English-expression training mode. UI localization and normal output-language
  selection are not valid reasons; those are handled by `{{language}}`,
  frontend i18n, and caller-provided runtime language.
- Resolution at runtime prefers exact-language match, falling back to
  `multi`. If neither the exact language nor `multi` is present, F3
  RegistryClient returns `ErrLanguageUnsupported`.

## 8 Status / lifecycle

- `draft`: file exists in the truth source but is not seeded into the
  database; the lint gate accepts it but Phase 4 seed migration excludes it.
- `active`: file is seeded into the database with `is_active = true`. Only
  one `active` YAML may exist per `(feature_key, language)` pair.
- `deprecated`: file is kept for history but not seeded as `active`. The
  lint gate accepts it; Phase 4 seed migration writes the row with
  `is_active = false` (or excludes it, depending on Phase 4.4 design).

DB-level `is_active` reflects staging/prod runtime state and does not write
back into this YAML. Edits to YAML `status` are pull-request reviewable.

## 9 Forbidden values

The lint gate rejects:

- Retired-module names from earlier product iterations. The exact list lives
  in `scripts/lint/prompt_lint.py` so this README can stay under the same
  recursive grep scan as the bodies it governs.
- Bare hardcoded prompt assignments inside Go business packages — these
  belong in this directory, not in code. The hardcode lint script
  (`scripts/lint/prompt_hardcode_lint.py`) enforces this separately.

## 10 References

- Spec: `docs/spec/prompt-rubric-registry/spec.md` v2.9
- Plans: `docs/spec/prompt-rubric-registry/plans/001-baseline/plan.md`,
  `docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/plan.md`,
  `docs/spec/prompt-rubric-registry/plans/003-language-coordinate-simplification/plan.md`,
  `docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/plan.md`
- DB schema: `migrations/000001_create_baseline.up.sql` (`prompt_versions` table)
- Lint: `scripts/lint/prompt_lint.py`
- Go loader: `backend/internal/ai/registry/loader.go`
- Sibling rubric README: `config/rubrics/README.md`
