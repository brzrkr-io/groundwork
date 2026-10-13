// Package content embeds the curriculum: course.json plus one pack per
// exercise under exercises/<name>/. See AUTHORING.md for the pack format.
package content

import "embed"

//go:embed course.json exercises
var Files embed.FS
