// Package cad provides lightweight CAD file inspection.
// Currently wraps FreeCAD headless CLI for metadata and BOM extraction.
// FreeCAD: https://github.com/FreeCAD/FreeCAD
//
// Install: apt-get install freecad-python3 OR flatpak install FreeCAD
package cad

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Inspect runs a FreeCAD headless script to extract metadata from a CAD file.
func Inspect(cadPath string) (string, error) {
	script := fmt.Sprintf(`
import FreeCAD
doc = FreeCAD.open("%s")
for obj in doc.Objects:
    print(obj.Label, obj.TypeId)
`, cadPath)
	cmd := exec.Command("freecad", "--console")
	cmd.Stdin = strings.NewReader(script)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cad: inspect %s: %w: %s", cadPath, err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}
