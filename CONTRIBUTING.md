# Contributing a template

This repo accepts PRs for the `custom` tier of the workflow
marketplace. Each accepted PR ships your template to every paid
Gen0Sec tenant that has subscribed.

## What ships, what doesn't

| Allowed | Rejected |
|---|---|
| Workflows that use only public engine activities (`approval.gate`, `notify.webhook`, `http.request`, `print`, etc.) | Workflows that reference `gen0sec.*` private activities — those live in the managed (paid) catalog |
| `${secrets.NAME}` / `${variables.NAME}` placeholders the installer fills in | Hard-coded credentials, customer-specific URLs, signing secrets |
| Workflows < 32 KiB of JSON | Larger blobs — split into multiple templates |
| English `name` + `description` | Marketing copy, vendor pitches |

## PR checklist

- [ ] One template per PR.
- [ ] Filename slug at `templates/custom/<slug>.json` matches
      `^[a-z0-9][a-z0-9-_]{0,62}$`.
- [ ] The JSON's top-level `name` equals the slug.
- [ ] Sibling `templates/custom/<slug>.meta.yaml` describes the
      template — see [MANIFEST.md](./MANIFEST.md) for the schema.
- [ ] `author.kind: community` + `author.name: <your-github-handle>`.
- [ ] Optional: PNG screenshot at `templates/custom/<slug>.png` and
      `screenshot: templates/custom/<slug>.png` in meta.
- [ ] CI passes — the engine validator + manifest build run on every PR.

## Local preview

```sh
go run ./scripts/build-manifest    # validates + writes manifest.json
```

If validation fails the error message points at the offending
template + line number.

## Review SLO

We aim to review community submissions within 5 business days.
Templates that look spammy, ship secrets, or reference private
activities will be closed without comment. Repeat offenders get
blocked.

## Maintainers

PRs are reviewed by the Gen0Sec workflow team. The two reviewers
required to land:

- One Gen0Sec maintainer
- One automated check (CI green)

## License

By submitting a PR you agree to license your template under the same
terms as this repo — see [LICENSE](./LICENSE).
