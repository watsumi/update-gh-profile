package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveSVG saves SVG content to a file
//
// Preconditions:
// - svgContent is a valid SVG string
// - filePath is a valid file path
//
// Postconditions:
// - SVG file is created at the specified path
// - File encoding is UTF-8
//
// Invariants:
// - Directories are automatically created if they don't exist
// - Existing files are overwritten
func SaveSVG(svgContent, filePath string) error {
	if svgContent == "" {
		return fmt.Errorf("SVG content is empty")
	}

	if filePath == "" {
		return fmt.Errorf("file path is empty")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Save SVG file with UTF-8 encoding
	// os.WriteFile writes in UTF-8, so explicit encoding specification is not needed
	err := os.WriteFile(filePath, []byte(svgContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to save SVG file: %w", err)
	}

	return nil
}

// SaveMultipleSVGs saves multiple SVG files at once
//
// Preconditions:
// - svgs is in the format map[filename]SVG content
// - outputDir is a valid directory path
//
// Postconditions:
// - Each SVG file is saved to outputDir
// - All files are saved successfully, or an error is returned
//
// Invariants:
// - All files are saved to the same directory
func SaveMultipleSVGs(svgs map[string]string, outputDir string) error {
	if len(svgs) == 0 {
		return fmt.Errorf("no SVG files to save")
	}

	if outputDir == "" {
		outputDir = "."
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Save each SVG
	for filename, content := range svgs {
		// Add .svg extension if not present
		if filepath.Ext(filename) != ".svg" {
			filename = filename + ".svg"
		}

		filepath := filepath.Join(outputDir, filename)

		if err := SaveSVG(content, filepath); err != nil {
			return fmt.Errorf("failed to save file %s: %w", filename, err)
		}
	}

	return nil
}
