// Package doc_parser provides document ingestion (PDF, DOCX, Markdown).
// Uses poppler-utils (pdftotext), pandoc, and native Go markdown.
//
// apt: apt-get install poppler-utils pandoc
package doc_parser

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Parse returns the plain-text content of a document.
func Parse(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return pdfToText(path)
	case ".docx", ".odt", ".rtf":
		return pandoc(path)
	case ".md", ".txt":
		out, err := exec.Command("cat", path).Output()
		return string(out), err
	default:
		return "", fmt.Errorf("doc_parser: unsupported type %s", ext)
	}
}

func pdfToText(path string) (string, error) {
	out, err := exec.Command("pdftotext", path, "-").Output()
	if err != nil {
		return "", fmt.Errorf("doc_parser: pdftotext %s: %w", path, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func pandoc(path string) (string, error) {
	out, err := exec.Command("pandoc", path, "-t", "plain").Output()
	if err != nil {
		return "", fmt.Errorf("doc_parser: pandoc %s: %w", path, err)
	}
	return strings.TrimSpace(string(out)), nil
}
