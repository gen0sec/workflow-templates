# gen0sec/workflow-templates

The public catalog backing the Gen0Sec workflow marketplace. Every
template in this repo ships as either a free **system** template
(curated by Gen0Sec, anyone can install) or a paid **custom**
template (community-contributed, install requires an active
subscription). The marketplace UI in
[`pigri/arxignis`](https://github.com/pigri/arxignis) reads the
generated manifest + template definitions from a Cloudflare R2
bucket; this repo publishes them.

| Tier | Directory | Who curates | Who can install |
|---|---|---|---|
| `system` | `templates/system/` | Gen0Sec maintainers | Anyone |
| `custom` | `templates/custom/` | Community via PR (Gen0Sec reviews) | Paid subscribers |

For paid **managed** templates, see the private companion repo
[`gen0sec/workflow-templates-managed`](https://github.com/gen0sec/workflow-templates-managed).

## What a template is

A single JSON file matching the engine's
[`workflow.Options`](https://github.com/gen0sec/workflow/blob/main/workflow.go)
shape — exactly the JSON the workflow builder serializes when you
click "View / edit JSON" in the canvas. The CI workflow validates
every changed file against the engine parser before publishing the
manifest.

## Submitting a template (`custom` tier)

Open a PR with a single file at `templates/custom/<slug>.json`. The
slug must match `^[a-z0-9][a-z0-9-_]{0,62}$` and equal the JSON's
top-level `name` field.

A template must:

1. Parse as a valid `workflow.Options` (CI runs the engine validator)
2. Not reference activities outside the public engine catalog (no
   `gen0sec.*` private activities — those are reserved for
   `workflow-templates-managed`)
3. Not hard-code tenant-specific secrets — use `${secrets.NAME}` /
   `${variables.NAME}` placeholders for anything the installer
   needs to supply
4. Carry a sibling metadata file at
   `templates/custom/<slug>.meta.yaml` with the manifest entry
   shape — see [`MANIFEST.md`](./MANIFEST.md) for the schema

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the full guide.

## How the publish works

`.github/workflows/sync-to-r2.yml` runs on push to `main`:

1. Validates every `templates/**/*.json` via the engine parser
2. Computes `sha256` of each definition (content addressing — the
   marketplace BFF re-verifies on every install)
3. Builds `manifest.json` by walking `templates/{system,custom}/*`
   and folding each entry's `.meta.yaml` into the catalog entry
4. PUTs `manifest.json` + each `templates/<tier>/<slug>.json` to
   the R2 bucket `workflow-templates-public` via the S3-compatible
   endpoint

Required repository secrets (configured under
**Settings → Secrets and variables → Actions**):

| Secret | Purpose |
|---|---|
| `R2_ACCOUNT_ID` | Cloudflare account id (subdomain of the R2 endpoint) |
| `R2_ACCESS_KEY_ID` | R2 API token with `workflow-templates-public` write |
| `R2_SECRET_ACCESS_KEY` | the token's secret |
| `R2_BUCKET` | `workflow-templates-public` |

## Layout

```
templates/
  system/
    <slug>.json          # the engine workflow.Options blob
    <slug>.meta.yaml     # name, description, category, tags, screenshot
  custom/
    <slug>.json
    <slug>.meta.yaml
scripts/
  build-manifest.go      # walks templates/, validates, computes sha256, builds manifest.json
.github/workflows/
  sync-to-r2.yml         # CI publish
MANIFEST.md              # manifest schema reference
CONTRIBUTING.md          # template submission guide
```

## Local preview

```sh
# Validate + build manifest locally (writes manifest.json)
go run ./scripts/build-manifest

# Stage the catalog into a local R2 bucket so the marketplace UI
# can read it during pnpm dev:
cd /path/to/arxignis/workflow/ui
npx wrangler r2 object put workflow-templates-public/manifest.json \
  --local --file ../../workflow-templates/manifest.json \
  --content-type application/json
for f in ../../workflow-templates/templates/*/*.json; do
  rel=$(realpath --relative-to=../../workflow-templates "$f")
  npx wrangler r2 object put "workflow-templates-public/$rel" \
    --local --file "$f" --content-type application/json
done
```
