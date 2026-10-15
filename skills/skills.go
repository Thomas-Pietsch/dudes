// Package skills ships the framework's level skills inside the dudes binary, so the engine finds
// them from any workspace — an instance repo holds only the company's own (CORE §5). The
// framework folder is not embedded: its skills are DevOps on DUDES, never a worker's.
package skills

import "embed"

// FS holds dev/, spec/, lead/ and shared/, each skill at <folder>/<name>/SKILL.md.
//
//go:embed dev shared spec lead
var FS embed.FS
