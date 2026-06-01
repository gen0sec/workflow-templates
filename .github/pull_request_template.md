<!--
Thanks for contributing a template!

Community submissions go in `templates/custom/`. The `templates/system/`
directory is reserved for Gen0Sec-curated templates that ship pre-installed
to every tenant; PRs that modify `templates/system/` from outside the
@gen0sec/developers team will be blocked by CI.

If you want your template considered for promotion to `system`, mention
that in the description and the maintainers will review separately.
-->

## What this PR adds

<!-- One-line summary -->

## Tier

- [ ] **Community (custom)** — adds files under `templates/custom/`
- [ ] **System (Gen0Sec only)** — adds/edits files under `templates/system/`
      (maintainers only)

## Checklist

- [ ] Template definition is at `templates/<tier>/<slug>.json`
- [ ] Sidecar metadata is at `templates/<tier>/<slug>.meta.yaml`
- [ ] `meta.yaml` `author.kind` matches the tier
      (`gen0sec` for system, anything else for custom)
- [ ] Template validates locally with `go run ./scripts/build-manifest .`
- [ ] No secrets / org-specific identifiers baked into the JSON
- [ ] Description in `meta.yaml` explains what the template does and
      when an operator would install it

## Notes

<!-- anything reviewers should know -->
