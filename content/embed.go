// Package content embeds the curriculum: course.json, one pack per exercise
// under exercises/<name>/, and the guides under guides/. See AUTHORING.md.
package content

import "embed"

//go:embed course.json exercises guides
var Files embed.FS
