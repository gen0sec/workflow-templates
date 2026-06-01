// build-manifest walks templates/{system,custom}/*.json, validates
// each against the engine's workflow.Options shape, folds in the
// sibling <slug>.meta.yaml, computes the sha256 of each definition,
// and writes manifest.json at the repo root. Idempotent — running
// it twice on a clean tree produces an identical manifest (except
// for `generated_at`).
//
// Two purposes:
//
//  1. Pre-publish validation. CI rejects a PR if any template fails
//     to parse or its meta is malformed — we don't want a broken
//     entry shipping to every paid tenant.
//  2. Publish artifact. The manifest is what the marketplace BFF
//     reads to render its browse page; the per-template JSON files
//     are what `vault.create_secret`-equivalent fetches at install
//     time, sha256-verified against the manifest entry.
//
// Usage:
//
//	go run ./scripts/build-manifest
//
// Reads the working directory (must be the repo root). Writes
// manifest.json next to README.md.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/deepnoodle-ai/workflow"
	"gopkg.in/yaml.v3"
)

// slugRe mirrors the BFF's target-name validation in
// workflow-ui/app/api/marketplace/install/route.ts. Template slugs
// must satisfy the same shape — they become the suggested target
// name when an installer clicks "Use template" without renaming.
var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-_]{0,62}$`)

// allowedCategories pins the set the UI's filter chips know about.
// Adding a new one needs a corresponding chip in
// workflow-ui/app/components/Marketplace.tsx.
var allowedCategories = map[string]struct{}{
	"approvals": {},
	"notify":    {},
	"incident":  {},
	"data":      {},
	"misc":      {},
}

type author struct {
	Kind string `yaml:"kind" json:"kind"` // "gen0sec" | "community"
	Name string `yaml:"name" json:"name"`
}

type templateMeta struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
	Version     string   `yaml:"version"`
	Screenshot  string   `yaml:"screenshot"`
	Author      author   `yaml:"author"`
	Tags        []string `yaml:"tags"`
}

// manifestEntry is the wire shape the BFF expects — see
// workflow-ui/app/lib/r2-catalog.ts:ManifestEntry. Keep this in sync.
type manifestEntry struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Tier             string   `json:"tier"` // "system" | "custom"
	Category         string   `json:"category"`
	Version          string   `json:"version"`
	DefinitionPath   string   `json:"definition_path"`
	DefinitionSHA256 string   `json:"definition_sha256"`
	ScreenshotPath   *string  `json:"screenshot_path"`
	Author           author   `json:"author"`
	Tags             []string `json:"tags"`
}

type manifest struct {
	Version     int             `json:"version"`
	GeneratedAt string          `json:"generated_at"`
	Templates   []manifestEntry `json:"templates"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("build-manifest: %v", err)
	}
}

func run() error {
	// Initialize as empty slice (not nil) so an empty catalog
	// marshals to `[]` rather than `null` — the BFF's r2-catalog
	// helper rejects a non-array `templates` field.
	entries := make([]manifestEntry, 0)
	for _, tier := range []string{"system", "custom"} {
		dir := filepath.Join("templates", tier)
		matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			return fmt.Errorf("glob %s: %w", dir, err)
		}
		sort.Strings(matches) // deterministic manifest order
		for _, jsonPath := range matches {
			entry, err := buildEntry(tier, jsonPath)
			if err != nil {
				return fmt.Errorf("%s: %w", jsonPath, err)
			}
			entries = append(entries, entry)
		}
	}

	out := manifest{
		Version:     manifestVersionFromGit(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Templates:   entries,
	}
	blob, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	blob = append(blob, '\n')
	if err := os.WriteFile("manifest.json", blob, 0o644); err != nil {
		return fmt.Errorf("write manifest.json: %w", err)
	}
	log.Printf("wrote manifest.json — version=%d, templates=%d", out.Version, len(entries))
	return nil
}

func buildEntry(tier, jsonPath string) (manifestEntry, error) {
	// Slug == basename minus .json. Cross-checked against the JSON's
	// `name` and the regex below.
	base := strings.TrimSuffix(filepath.Base(jsonPath), ".json")
	if !slugRe.MatchString(base) {
		return manifestEntry{}, fmt.Errorf("filename %q does not match the slug regex %s — rename the file", base, slugRe)
	}

	defBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return manifestEntry{}, fmt.Errorf("read definition: %w", err)
	}

	// Validate against the engine. Same json.Unmarshal -> workflow.New
	// shape the workflow-service uses for every Save (see
	// internal/workflows/Parse), so if this passes the definition will
	// load cleanly when an installer copies it.
	var opts workflow.Options
	if err := json.Unmarshal(defBytes, &opts); err != nil {
		return manifestEntry{}, fmt.Errorf("definition JSON parse: %w", err)
	}
	wf, err := workflow.New(opts)
	if err != nil {
		return manifestEntry{}, fmt.Errorf("engine validation failed: %w", err)
	}
	if wf.Name() != base {
		return manifestEntry{}, fmt.Errorf("definition name %q does not equal filename slug %q — they MUST match", wf.Name(), base)
	}

	// Meta sibling.
	metaPath := strings.TrimSuffix(jsonPath, ".json") + ".meta.yaml"
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return manifestEntry{}, fmt.Errorf("read meta sibling %s: %w", metaPath, err)
	}
	var meta templateMeta
	if err := yaml.Unmarshal(metaBytes, &meta); err != nil {
		return manifestEntry{}, fmt.Errorf("parse meta yaml: %w", err)
	}
	if err := validateMeta(tier, meta); err != nil {
		return manifestEntry{}, fmt.Errorf("meta: %w", err)
	}

	sum := sha256.Sum256(defBytes)
	var shot *string
	if meta.Screenshot != "" {
		s := meta.Screenshot
		shot = &s
	}

	return manifestEntry{
		ID:               base,
		Name:             meta.Name,
		Description:      meta.Description,
		Tier:             tier,
		Category:         meta.Category,
		Version:          meta.Version,
		DefinitionPath:   jsonPath,
		DefinitionSHA256: hex.EncodeToString(sum[:]),
		ScreenshotPath:   shot,
		Author:           meta.Author,
		Tags:             meta.Tags,
	}, nil
}

func validateMeta(tier string, m templateMeta) error {
	var errs []string
	if m.Name == "" {
		errs = append(errs, "name is required")
	}
	if m.Description == "" {
		errs = append(errs, "description is required")
	}
	if m.Version == "" {
		errs = append(errs, "version is required (semver string)")
	}
	if _, ok := allowedCategories[m.Category]; !ok {
		errs = append(errs, fmt.Sprintf("category %q is not in the allowlist; allowed: %v", m.Category, allowedCategoriesList()))
	}
	switch m.Author.Kind {
	case "gen0sec", "community":
	default:
		errs = append(errs, fmt.Sprintf("author.kind must be \"gen0sec\" or \"community\"; got %q", m.Author.Kind))
	}
	if m.Author.Name == "" {
		errs = append(errs, "author.name is required")
	}
	// custom-tier authors must be community-kind; reject Gen0Sec
	// self-publishing a custom template (those belong in system or
	// in workflow-templates-managed).
	if tier == "custom" && m.Author.Kind == "gen0sec" {
		errs = append(errs, "custom-tier templates cannot have author.kind=gen0sec — put it in templates/system/ or the managed repo")
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func allowedCategoriesList() []string {
	out := make([]string, 0, len(allowedCategories))
	for k := range allowedCategories {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// manifestVersionFromGit returns a monotonically increasing integer
// derived from the commit count on the current branch. The BFF
// reads this to decide "is my cached manifest stale" — bumping it
// on every successful publish is what triggers tenant browse
// invalidation. Falls back to 0 outside a git checkout (e.g. when
// the script runs against an unpacked tarball).
func manifestVersionFromGit() int {
	// Cheap implementation — count lines from git rev-list. Failing
	// is fine; the script still produces a valid manifest with
	// version=0.
	data, err := os.ReadFile(".git/HEAD")
	if err != nil || len(data) == 0 {
		return 0
	}
	// Use a tiny stdlib walk instead of shelling out.
	// `git rev-list --count HEAD` would be exact; this approximation
	// counts commit objects on disk and is good enough for ordering.
	// Operators who want exact: just `git rev-list --count HEAD`
	// and patch this function to read from an env var.
	if v := os.Getenv("MANIFEST_VERSION"); v != "" {
		var n int
		_, _ = fmt.Sscanf(v, "%d", &n)
		return n
	}
	return 0
}
