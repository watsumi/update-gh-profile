package repository

import (
	"path/filepath"
	"strings"
)

// Extension to language mapping
var extToLang = map[string]string{
	".go":         "Go",
	".py":         "Python",
	".js":         "JavaScript",
	".ts":         "TypeScript",
	".tsx":        "TypeScript",
	".jsx":        "JavaScript",
	".java":       "Java",
	".kt":         "Kotlin",
	".scala":      "Scala",
	".clj":        "Clojure",
	".cljs":       "Clojure",
	".cpp":        "C++",
	".cc":         "C++",
	".cxx":        "C++",
	".c":          "C",
	".h":          "C",
	".hpp":        "C++",
	".cs":         "C#",
	".php":        "PHP",
	".rb":         "Ruby",
	".swift":      "Swift",
	".dart":       "Dart",
	".rs":         "Rust",
	".cr":         "Crystal",
	".ex":         "Elixir",
	".exs":        "Elixir",
	".erl":        "Erlang",
	".hrl":        "Erlang",
	".lua":        "Lua",
	".r":          "R",
	".sql":        "SQL",
	".sh":         "Shell",
	".bash":       "Shell",
	".zsh":        "Shell",
	".fish":       "Shell",
	".ps1":        "PowerShell",
	".psm1":       "PowerShell",
	".html":       "HTML",
	".htm":        "HTML",
	".css":        "CSS",
	".scss":       "SCSS",
	".sass":       "Sass",
	".less":       "Less",
	".xml":        "XML",
	".json":       "JSON",
	".yaml":       "YAML",
	".yml":        "YAML",
	".toml":       "TOML",
	".ini":        "INI",
	".md":         "Markdown",
	".tex":        "LaTeX",
	".rkt":        "Racket",
	".ml":         "OCaml",
	".fs":         "F#",
	".fsx":        "F#",
	".vb":         "Visual Basic",
	".hs":         "Haskell",
	".lhs":        "Haskell",
	".jl":         "Julia",
	".pl":         "Perl",
	".pm":         "Perl",
	".nim":        "Nim",
	".zig":        "Zig",
	".v":          "V",
	".dockerfile": "Dockerfile",
	".makefile":   "Makefile",
	".make":       "Makefile",
}

// DetectLanguageFromFilename detects language from filename
//
// Preconditions:
// - filename is a valid file path
//
// Postconditions:
// - Returns language name (empty string if cannot determine)
//
// Invariants:
// - Detects language from extension and filename
func DetectLanguageFromFilename(filename string) string {
	if filename == "" {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filename))
	filenameLower := strings.ToLower(filepath.Base(filename))

	// Detect from extension
	if lang, ok := extToLang[ext]; ok {
		return lang
	}

	// Detect from filename (Dockerfile, Makefile, etc.)
	if lang, ok := extToLang[filenameLower]; ok {
		return lang
	}

	// Special filename patterns
	if strings.HasPrefix(filenameLower, "dockerfile") {
		return "Dockerfile"
	}
	if strings.HasPrefix(filenameLower, "makefile") {
		return "Makefile"
	}

	return ""
}
