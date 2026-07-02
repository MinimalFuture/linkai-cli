package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/MinimalFuture/linkai-cli/internal/skill"
)

// embeddedSkillsFS bundles the agent-readable skill docs (SKILL.md +
// references/) so they ship in lockstep with the binary and can be installed
// into agent skill directories without a network fetch. Machine-only resources
// (assets/, scripts/) are intentionally excluded. The embed must live in this
// root package because go:embed cannot reach up out of a package's directory.
//
//go:embed skills/*/SKILL.md skills/*/references
var embeddedSkillsFS embed.FS

// init wires the embedded skill content into the skill package. On assembly
// failure it warns rather than panicking — embedded skills are nice-to-have,
// and the rest of the CLI keeps working.
func init() {
	sub, err := fs.Sub(embeddedSkillsFS, "skills")
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: skill embed assembly failed, 'skill' command disabled:", err)
		return
	}
	skill.SetSource(sub)
}
