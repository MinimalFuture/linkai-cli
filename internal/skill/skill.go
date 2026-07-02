// Package skill serves the agent-readable skill content that ships embedded in
// the CLI binary (SKILL.md + references/*.md), and installs it into the agent
// homes that tools like Claude Code / Cursor / Codex scan for skills.
//
// The embed.FS itself lives in the root package (go:embed cannot reach up out
// of a package's directory); it is injected here via SetSource at init time.
// When no source is set (e.g. a build without the embed), the exported helpers
// return an error so callers can degrade gracefully.
package skill

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// source is the embedded skills tree, rooted so that each top-level entry is a
// skill directory (e.g. "linkai-cli/SKILL.md"). Injected by the root package.
var source fs.FS

// SetSource wires the binary-embedded skill content into this package. It is
// called once from the root package's init.
func SetSource(fsys fs.FS) { source = fsys }

// ErrNoContent is returned when no skill content was embedded in the build.
var ErrNoContent = errors.New("skill content not embedded in this build")

// agentSkillDirs are the per-agent skill directories (relative to a base dir
// such as $HOME) that agents scan for installed skills. This mirrors the list
// baked into install.sh so both installers behave identically.
var agentSkillDirs = []string{
	".agents/skills", // generic fallback, always attempted
	"cow/skills",
	".claude/skills",
	".cursor/skills",
	".codex/skills",
	".gemini/skills",
	".windsurf/skills",
	".qoder/skills",
}

// SkillInfo describes a skill available for listing.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// List returns the skills embedded in the binary.
func List() ([]SkillInfo, error) {
	if source == nil {
		return nil, ErrNoContent
	}
	entries, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, err
	}
	var skills []SkillInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info := SkillInfo{Name: e.Name()}
		if md, err := fs.ReadFile(source, e.Name()+"/SKILL.md"); err == nil {
			info.Description = frontmatterDescription(md)
		}
		skills = append(skills, info)
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

// Read returns the raw bytes of a file within a skill. relpath "" (or
// "SKILL.md") reads the skill's SKILL.md. The path is confined to the skill
// dir; any attempt to escape it (via "..") is rejected.
func Read(name, relpath string) ([]byte, string, error) {
	if source == nil {
		return nil, "", ErrNoContent
	}
	if relpath == "" {
		relpath = "SKILL.md"
	}
	clean := filepath.ToSlash(filepath.Clean(relpath))
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return nil, "", errors.New("invalid path: must stay within the skill directory")
	}
	full := name + "/" + clean
	data, err := fs.ReadFile(source, full)
	if err != nil {
		return nil, "", err
	}
	return data, clean, nil
}

// InstallResult reports where a skill was installed.
type InstallResult struct {
	Dest    string `json:"dest"`
	Skipped bool   `json:"skipped"`
}

// Install copies every embedded skill into the given base directory's agent
// skill dirs. When a specific dir is provided it copies the skills directly
// under that dir (one <dir>/<skill-name> per skill) regardless of agent
// detection. Returns the list of destinations written.
//
// When base is empty it defaults to $HOME. Agent dirs whose parent does not
// exist are skipped (except the generic .agents fallback, always attempted),
// matching install.sh's behavior.
func Install(base string) ([]InstallResult, error) {
	if source == nil {
		return nil, ErrNoContent
	}
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		base = home
	}

	skills, err := List()
	if err != nil {
		return nil, err
	}

	var results []InstallResult
	for _, dir := range agentSkillDirs {
		parent := filepath.Join(base, filepath.Dir(dir))
		// Only install into agent homes that already exist, so we don't
		// litter directories for agents the user doesn't use. The generic
		// .agents fallback is always attempted.
		if dir != ".agents/skills" {
			if _, statErr := os.Stat(parent); statErr != nil {
				continue
			}
		}
		for _, s := range skills {
			dest := filepath.Join(base, dir, s.Name)
			if err := copySkill(s.Name, dest); err != nil {
				return results, err
			}
			results = append(results, InstallResult{Dest: dest})
		}
	}
	return results, nil
}

// InstallToDir copies every embedded skill directly under dir (one
// <dir>/<skill-name> per skill), without any agent-home detection. Used for an
// explicit --dir target.
func InstallToDir(dir string) ([]InstallResult, error) {
	if source == nil {
		return nil, ErrNoContent
	}
	skills, err := List()
	if err != nil {
		return nil, err
	}
	var results []InstallResult
	for _, s := range skills {
		dest := filepath.Join(dir, s.Name)
		if err := copySkill(s.Name, dest); err != nil {
			return results, err
		}
		results = append(results, InstallResult{Dest: dest})
	}
	return results, nil
}

// copySkill writes the embedded skill tree rooted at name into dest, replacing
// any existing content there so a stale prior install can't linger.
func copySkill(name, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return fs.WalkDir(source, name, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(name, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(source, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// frontmatterDescription extracts the `description:` value from a SKILL.md YAML
// frontmatter block. Returns "" when absent.
func frontmatterDescription(md []byte) string {
	lines := strings.Split(string(md), "\n")
	inFront := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFront {
				inFront = true
				continue
			}
			break
		}
		if inFront && strings.HasPrefix(trimmed, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
		}
	}
	return ""
}
