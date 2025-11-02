package repository

import (
	"path/filepath"
	"strings"
)

// 拡張子と言語のマッピング
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

// DetectLanguageFromFilename ファイル名から言語を判定する
//
// Preconditions:
// - filename が有効なファイルパスであること
//
// Postconditions:
// - 言語名を返す（判定できない場合は空文字列）
//
// Invariants:
// - 拡張子とファイル名から言語を判定する
func DetectLanguageFromFilename(filename string) string {
	if filename == "" {
		return ""
	}

	ext := strings.ToLower(filepath.Ext(filename))
	filenameLower := strings.ToLower(filepath.Base(filename))

	// 拡張子から判定
	if lang, ok := extToLang[ext]; ok {
		return lang
	}

	// ファイル名から判定（Dockerfile, Makefile など）
	if lang, ok := extToLang[filenameLower]; ok {
		return lang
	}

	// 特殊なファイル名パターン
	if strings.HasPrefix(filenameLower, "dockerfile") {
		return "Dockerfile"
	}
	if strings.HasPrefix(filenameLower, "makefile") {
		return "Makefile"
	}

	return ""
}
