package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveSVG(t *testing.T) {
	tests := []struct {
		name       string
		svgContent string
		filepath   string
		wantError  bool
	}{
		{
			name:       "Normal case: basic save",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
			filepath:   "test_output.svg",
			wantError:  false,
		},
		{
			name:       "Normal case: save to subdirectory",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg"><text>Test</text></svg>`,
			filepath:   "test_output/subdir/test.svg",
			wantError:  false,
		},
		{
			name:       "Error: empty content",
			svgContent: "",
			filepath:   "test_empty.svg",
			wantError:  true,
		},
		{
			name:       "Error: empty file path",
			svgContent: `<?xml version="1.0" encoding="UTF-8"?><svg></svg>`,
			filepath:   "",
			wantError:  true,
		},
	}

	// Create temporary directory for testing
	testDir := "test_output"
	defer func() {
		// Cleanup after test
		os.RemoveAll(testDir)
		os.Remove("test_output.svg")
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare file path for testing
			var testPath string
			if tt.filepath != "" {
				if filepath.IsAbs(tt.filepath) {
					testPath = tt.filepath
				} else {
					testPath = filepath.Join(testDir, tt.filepath)
				}
			}

			err := SaveSVG(tt.svgContent, testPath)

			if tt.wantError {
				if err == nil {
					t.Errorf("SaveSVG() should have returned an error, but returned nil")
				}
			} else {
				if err != nil {
					t.Errorf("SaveSVG() error = %v, did not expect error", err)
					return
				}

				// Verify file was actually created
				if testPath != "" {
					if _, err := os.Stat(testPath); os.IsNotExist(err) {
						t.Errorf("SaveSVG() file was not created: %s", testPath)
					}
				}
			}
		})
	}
}

func TestSaveMultipleSVGs(t *testing.T) {
	testDir := "test_multiple_output"
	defer func() {
		os.RemoveAll(testDir)
	}()

	tests := []struct {
		name      string
		svgs      map[string]string
		outputDir string
		wantError bool
	}{
		{
			name: "Normal case: save multiple SVGs",
			svgs: map[string]string{
				"chart1":     `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 1</text></svg>`,
				"chart2":     `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 2</text></svg>`,
				"chart3.svg": `<?xml version="1.0" encoding="UTF-8"?><svg><text>Chart 3</text></svg>`,
			},
			outputDir: testDir,
			wantError: false,
		},
		{
			name:      "Error: empty map",
			svgs:      map[string]string{},
			outputDir: testDir,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveMultipleSVGs(tt.svgs, tt.outputDir)

			if tt.wantError {
				if err == nil {
					t.Errorf("SaveMultipleSVGs() should have returned an error, but returned nil")
				}
			} else {
				if err != nil {
					t.Errorf("SaveMultipleSVGs() error = %v, did not expect error", err)
					return
				}

				// Verify all files were created
				for filename := range tt.svgs {
					expectedExt := ".svg"
					if filepath.Ext(filename) != "" {
						expectedExt = ""
					}
					filePath := filepath.Join(tt.outputDir, filename+expectedExt)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("SaveMultipleSVGs() ファイルが作成されませんでした: %s", filePath)
					}
				}
			}
		})
	}
}
