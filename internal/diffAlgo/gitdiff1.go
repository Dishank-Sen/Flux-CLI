package diffalgo

import (
	"exp1/internal/types"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// ===== File Helper =====

// ReadFileContent reads a file and returns its content as string
func ReadFileContent(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ===== Core Functions =====

// ComputeDelta computes the delta between oldText and newText
func ComputeDelta(oldPath, newPath, oldText, newText string) []types.LineChange {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")
	var changes []types.LineChange

	i, j := 0, 0
	for i < len(oldLines) || j < len(newLines) {
		switch {
		case i < len(oldLines) && j < len(newLines):
			if oldLines[i] == newLines[j] {
				i++
				j++
			} else {
				// Line differs: compute char-level diff
				changes = append(changes, types.LineChange{
					FilePath:   newPath,
					LineNumber: i,
					Type:       "replace",
					Content:    []string{newLines[j]},
					CharDiff:   ComputeCharDiff(oldLines[i], newLines[j]),
				})
				i++
				j++
			}
		case i >= len(oldLines) && j < len(newLines):
			changes = append(changes, types.LineChange{
				FilePath:   newPath,
				LineNumber: i,
				Type:       "add",
				Content:    []string{newLines[j]},
			})
			j++
			i++
		case i < len(oldLines) && j >= len(newLines):
			changes = append(changes, types.LineChange{
				FilePath:   oldPath,
				LineNumber: i,
				Type:       "delete",
			})
			i++
		}
	}

	return changes
}

// ComputeCharDiff computes character-level diffs for a single line
func ComputeCharDiff(oldLine, newLine string) []types.CharDiff {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldLine, newLine, false)
	charDiffs := make([]types.CharDiff, len(diffs))
	for k, d := range diffs {
		var t string
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			t = "Equal"
		case diffmatchpatch.DiffInsert:
			t = "Insert"
		case diffmatchpatch.DiffDelete:
			t = "Delete"
		}
		charDiffs[k] = types.CharDiff{
			Type: t,
			Text: d.Text,
		}
	}
	return charDiffs
}

// ApplyDelta reconstructs newText from oldText + delta
func ApplyDelta(oldText string, changes []types.LineChange) string {
	lines := strings.Split(oldText, "\n")
	offset := 0

	for _, change := range changes {
		idx := change.LineNumber + offset
		switch change.Type {
		case "add":
			if idx > len(lines) {
				idx = len(lines)
			}
			lines = append(lines[:idx], append(change.Content, lines[idx:]...)...)
			offset += len(change.Content)
		case "delete":
			if idx < len(lines) {
				lines = append(lines[:idx], lines[idx+1:]...)
				offset--
			}
		case "replace":
			if idx < len(lines) {
				lines = append(lines[:idx], append(change.Content, lines[idx+1:]...)...)
			} else {
				lines = append(lines, change.Content...)
			}
		}
	}

	return strings.Join(lines, "\n")
}
