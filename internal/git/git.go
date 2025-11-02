package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectChanges 変更されたファイルを検出する
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
//
// Postconditions:
// - 変更されたファイルのリストが返される
// - エラーが発生した場合はエラーが返される
//
// Invariants:
// - 返されるファイルパスは相対パスである
func DetectChanges(repoPath string) ([]string, error) {
	// git status --porcelain を実行して変更ファイルを取得
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git status の実行に失敗しました: %w\nstderr: %s", err, stderr.String())
	}

	// 出力を解析してファイルリストを作成
	output := stdout.String()
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string

	for _, line := range lines {
		// git status --porcelain の出力形式: " M file.txt" または "MM file.txt"
		// ステータスコードの後の部分がファイルパス
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// 2番目のフィールドがファイルパス
			file := strings.Join(fields[1:], " ")
			// 絶対パスを相対パスに変換
			if filepath.IsAbs(file) {
				relPath, err := filepath.Rel(repoPath, file)
				if err == nil {
					file = relPath
				}
			}
			files = append(files, file)
		}
	}

	return files, nil
}

// HasChanges リポジトリに変更があるかどうかを確認する
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
//
// Postconditions:
// - 変更がある場合は true、ない場合は false が返される
// - エラーが発生した場合はエラーが返される
//
// Invariants:
// - Git リポジトリが初期化されていることが前提
func HasChanges(repoPath string) (bool, error) {
	files, err := DetectChanges(repoPath)
	if err != nil {
		return false, err
	}

	return len(files) > 0, nil
}

// Commit 変更をコミットする
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
// - message が有効なコミットメッセージであること
// - files がコミットするファイルのリストであること（空の場合はすべての変更をコミット）
//
// Postconditions:
// - 指定されたファイルがコミットされる
// - コミットメッセージが適用される
//
// Invariants:
// - Git リポジトリが初期化されていることが前提
func Commit(repoPath, message string, files []string) error {
	if message == "" {
		return fmt.Errorf("コミットメッセージが空です")
	}

	// ファイルが指定されていない場合は、すべての変更をステージング
	if len(files) == 0 {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = repoPath

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("git add の実行に失敗しました: %w\nstderr: %s", err, stderr.String())
		}
	} else {
		// 指定されたファイルをステージング
		args := append([]string{"add"}, files...)
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("git add の実行に失敗しました: %w\nstderr: %s", err, stderr.String())
		}
	}

	// コミットを実行
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		// コミットする変更がない場合はエラーとしない（既にコミット済みの場合）
		if strings.Contains(stderrStr, "nothing to commit") || strings.Contains(stderrStr, "nothing added to commit") {
			return nil
		}
		return fmt.Errorf("git commit の実行に失敗しました: %w\nstderr: %s", err, stderrStr)
	}

	return nil
}

// Push リモートリポジトリにプッシュする
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
// - remote がリモート名であること（省略可能、デフォルトは "origin"）
// - branch がブランチ名であること（省略可能、デフォルトは現在のブランチ）
//
// Postconditions:
// - 変更がリモートリポジトリにプッシュされる
//
// Invariants:
// - リモートリポジトリが設定されていることが前提
func Push(repoPath, remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}

	// ブランチが指定されていない場合は現在のブランチを取得
	if branch == "" {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("現在のブランチの取得に失敗しました: %w\nstderr: %s", err, stderr.String())
		}

		branch = strings.TrimSpace(stdout.String())
	}

	// プッシュを実行
	cmd := exec.Command("git", "push", remote, branch)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git push の実行に失敗しました: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// CommitAndPush 変更をコミットしてプッシュする（一括処理）
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
// - message が有効なコミットメッセージであること
// - files がコミットするファイルのリストであること（空の場合はすべての変更をコミット）
// - remote がリモート名であること（省略可能）
// - branch がブランチ名であること（省略可能）
//
// Postconditions:
// - 変更がコミットされ、リモートリポジトリにプッシュされる
//
// Invariants:
// - コミットとプッシュが順次実行される
func CommitAndPush(repoPath, message string, files []string, remote, branch string) error {
	// 変更があるか確認
	hasChanges, err := HasChanges(repoPath)
	if err != nil {
		return fmt.Errorf("変更の確認に失敗しました: %w", err)
	}

	if !hasChanges {
		return nil // 変更がない場合は何もしない
	}

	// コミット
	err = Commit(repoPath, message, files)
	if err != nil {
		return fmt.Errorf("コミットに失敗しました: %w", err)
	}

	// プッシュ
	err = Push(repoPath, remote, branch)
	if err != nil {
		return fmt.Errorf("プッシュに失敗しました: %w", err)
	}

	return nil
}

// IsGitRepository 指定されたパスが Git リポジトリかどうかを確認する
//
// Preconditions:
// - repoPath が有効なディレクトリパスであること
//
// Postconditions:
// - Git リポジトリの場合は true、そうでない場合は false が返される
//
// Invariants:
// - .git ディレクトリの存在を確認する（シンボリックリンクも含む）
func IsGitRepository(repoPath string) bool {
	// 絶対パスに変換
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return false
	}

	gitDir := filepath.Join(absPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	// ディレクトリまたはシンボリックリンクの場合は Git リポジトリとみなす
	return info.IsDir() || (info.Mode()&os.ModeSymlink != 0)
}

// GetCurrentBranch 現在のブランチ名を取得する
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
//
// Postconditions:
// - 現在のブランチ名が返される
//
// Invariants:
// - Git リポジトリが初期化されていることが前提
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("現在のブランチの取得に失敗しました: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
