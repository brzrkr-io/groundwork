// Package content embeds the curriculum: course.json, one pack per exercise
// under packs/<name>/, and the guides under guides/. See AUTHORING.md.
package content

import "embed"

//go:embed course.json packs guides
var Files embed.FS
