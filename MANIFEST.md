# Manifest schema

The build script (`scripts/build-manifest.go`) writes `manifest.json`
at the repo root. The marketplace BFF reads it from R2 and merges it
with the managed-tier manifest before serving
`GET /api/marketplace/templates`. The shape MUST match the
`Manifest` type in
[`workflow-ui/app/lib/r2-catalog.ts`](https://github.com/pigri/arxignis/blob/feature/workflow-template-store/workflow/ui/app/lib/r2-catalog.ts).

## Top-level

```json
{
  "version": 12,
  "generated_at": "2026-06-01T13:45:00Z",
  "templates": [ /* entries */ ]
}
```

- `version` — monotonically increasing integer the BFF uses to
  decide "is my cached manifest stale". Bump on every successful
  publish; the build script does this automatically using the
  commit count on `main`.
- `generated_at` — RFC3339 UTC timestamp the build ran.
- `templates` — array of entries (below).

## Per-template entry

```json
{
  "id": "approval-default",
  "name": "Default approval",
  "description": "Single email-based approval gate with the platform's generic body template.",
  "tier": "system",
  "category": "approvals",
  "version": "1.0.0",
  "definition_path": "templates/system/approval-default.json",
  "definition_sha256": "9f86d081884c…",
  "screenshot_path": null,
  "author": { "kind": "gen0sec", "name": "Gen0Sec" },
  "tags": ["approval", "email"]
}
```

| Field | Source | Notes |
|---|---|---|
| `id` | filename slug | `<basename>.json` minus the `.json`. Must equal the JSON's `name`. |
| `name` | `meta.yaml` `name` | Human label shown on the marketplace card. |
| `description` | `meta.yaml` `description` | One-paragraph blurb. |
| `tier` | parent directory | `system` for `templates/system/`, `custom` for `templates/custom/`. |
| `category` | `meta.yaml` `category` | One of `approvals`, `notify`, `incident`, `data`, `misc`. Used by the UI's filter chips. |
| `version` | `meta.yaml` `version` | SemVer string for this template. Bump on every change. |
| `definition_path` | computed | `templates/<tier>/<id>.json` |
| `definition_sha256` | computed by build script | `sha256(definition_bytes)` hex |
| `screenshot_path` | optional, `meta.yaml` `screenshot` | Path to an optional preview PNG inside the repo. |
| `author` | `meta.yaml` `author` | `{kind: "gen0sec"\|"community", name: "<display>"}` |
| `tags` | `meta.yaml` `tags` | Free-form string array; the UI search matches against these. |

## meta.yaml

Sibling YAML next to each `.json`. The build script reads this and
folds it into the manifest entry. Schema:

```yaml
name: "Default approval"
description: "Single email-based approval gate with the platform's generic body template."
category: "approvals"
version: "1.0.0"
screenshot: null
author:
  kind: gen0sec     # or "community"
  name: "Gen0Sec"
tags: [approval, email]
```

`kind` for the system tier is always `gen0sec`. For custom-tier PRs,
contributors set `kind: community` and `name: "<their github
handle>"`.
