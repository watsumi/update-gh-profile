package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureGitConfig Git の user.name と user.email を設定する
// GitHub Actions環境では GITHUB_ACTOR 環境変数を使用
func ensureGitConfig(repoPath string) error {
	// user.name を確認
	cmd := exec.Command("git", "config", "--local", "user.name")
	cmd.Dir = repoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	userNameSet := err == nil && strings.TrimSpace(stdout.String()) != ""

	if !userNameSet {
		// user.name が設定されていない場合、設定する
		var userName string
		if actor := os.Getenv("GITHUB_ACTOR"); actor != "" {
			userName = actor
		} else {
			// フォールバック: GitHub Actions デフォルト
			userName = "github-actions[bot]"
		}

		cmd := exec.Command("git", "config", "--local", "user.name", userName)
		cmd.Dir = repoPath
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("user.name の設定に失敗しました: %w\nstderr: %s", err, stderr.String())
		}
	}

	// user.email を確認
	cmd = exec.Command("git", "config", "--local", "user.email")
	cmd.Dir = repoPath
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	userEmailSet := err == nil && strings.TrimSpace(stdout.String()) != ""

	if !userEmailSet {
		// user.email が設定されていない場合、設定する
		var userEmail string
		if actor := os.Getenv("GITHUB_ACTOR"); actor != "" {
			// GitHub Actions の no-reply メールアドレス形式
			userEmail = fmt.Sprintf("%s@users.noreply.github.com", actor)
		} else {
			// フォールバック: GitHub Actions デフォルト
			userEmail = "github-actions[bot]@users.noreply.github.com"
		}

		cmd := exec.Command("git", "config", "--local", "user.email", userEmail)
		cmd.Dir = repoPath
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("user.email の設定に失敗しました: %w\nstderr: %s", err, stderr.String())
		}
	}

	return nil
}

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

	// Git の user.name と user.email が設定されているか確認し、設定されていない場合は設定する
	// GitHub Actions環境では GITHUB_ACTOR を使用
	err := ensureGitConfig(repoPath)
	if err != nil {
		return fmt.Errorf("Git 設定の確認に失敗しました: %w", err)
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

// SetRemoteURLWithToken リモートURLにトークンを設定する
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
// - remote がリモート名であること（省略可能、デフォルトは "origin"）
// - token がGitHub Personal Access Tokenであること
//
// Postconditions:
// - リモートURLがトークンを含む形式に更新される
//
// Invariants:
// - リモートリポジトリが設定されていることが前提
func SetRemoteURLWithToken(repoPath, remote, token string) error {
	if remote == "" {
		remote = "origin"
	}

	if token == "" {
		return fmt.Errorf("トークンが設定されていません")
	}

	// 現在のリモートURLを取得
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("リモートURLの取得に失敗しました: %w\nstderr: %s", err, stderr.String())
	}

	currentURL := strings.TrimSpace(stdout.String())

	// HTTPS形式のURLを構築
	// https://github.com/owner/repo.git を https://TOKEN@github.com/owner/repo.git に変換
	var newURL string
	if strings.HasPrefix(currentURL, "https://") {
		// 既に https:// で始まる場合
		if strings.Contains(currentURL, "@") {
			// 既にトークンが含まれている場合は、置き換える
			// https://oldtoken@github.com/owner/repo.git -> https://newtoken@github.com/owner/repo.git
			parts := strings.SplitN(currentURL, "@", 2)
			if len(parts) == 2 {
				// parts[1] は github.com/owner/repo.git の部分
				newURL = fmt.Sprintf("https://%s@%s", token, parts[1])
			} else {
				return fmt.Errorf("リモートURLの解析に失敗しました: %s", currentURL)
			}
		} else {
			// トークンが含まれていない場合は追加
			// https://github.com/owner/repo.git -> https://TOKEN@github.com/owner/repo.git
			newURL = strings.Replace(currentURL, "https://", fmt.Sprintf("https://%s@", token), 1)
		}
	} else if strings.HasPrefix(currentURL, "git@") {
		// SSH形式の場合はHTTPS形式に変換してトークンを追加
		// git@github.com:owner/repo.git -> https://TOKEN@github.com/owner/repo.git
		newURL = strings.Replace(currentURL, "git@github.com:", fmt.Sprintf("https://%s@github.com/", token), 1)
	} else {
		return fmt.Errorf("サポートされていないリモートURL形式です: %s", currentURL)
	}

	// リモートURLを設定
	cmd = exec.Command("git", "remote", "set-url", remote, newURL)
	cmd.Dir = repoPath

	stderr.Reset()
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("リモートURLの設定に失敗しました: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// Push リモートリポジトリにプッシュする
//
// Preconditions:
// - repoPath が有効な Git リポジトリのパスであること
// - remote がリモート名であること（省略可能、デフォルトは "origin"）
// - branch がブランチ名であること（省略可能、デフォルトは現在のブランチ）
// - token がGitHub Personal Access Tokenであること（省略可能、設定されていない場合はURLにトークンを含めない）
//
// Postconditions:
// - 変更がリモートリポジトリにプッシュされる
//
// Invariants:
// - リモートリポジトリが設定されていることが前提
func Push(repoPath, remote, branch, token string) error {
	if remote == "" {
		remote = "origin"
	}

	// トークンが設定されている場合はリモートURLに設定
	if token != "" {
		err := SetRemoteURLWithToken(repoPath, remote, token)
		if err != nil {
			return fmt.Errorf("リモートURLへのトークン設定に失敗しました: %w", err)
		}
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
// - token がGitHub Personal Access Tokenであること（省略可能）
//
// Postconditions:
// - 変更がコミットされ、リモートリポジトリにプッシュされる
//
// Invariants:
// - コミットとプッシュが順次実行される
func CommitAndPush(repoPath, message string, files []string, remote, branch, token string) error {
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
	err = Push(repoPath, remote, branch, token)
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
// - git rev-parse コマンドを使用して確認（より確実）
// - フォールバックとして .git ディレクトリ/ファイルの存在も確認
func IsGitRepository(repoPath string) bool {
	// まず git rev-parse コマンドで確認（より確実な方法）
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		// git rev-parse が成功した場合は Git リポジトリ
		return true
	}

	// git rev-parse が失敗した場合、フォールバックとして .git の存在を確認
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return false
	}

	gitDir := filepath.Join(absPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	// ディレクトリ、ファイル、またはシンボリックリンクの場合は Git リポジトリとみなす
	// (shallow clone の場合は .git がファイルになることがある)
	return info.IsDir() || (info.Mode()&os.ModeSymlink != 0) || (info.Mode().IsRegular())
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
