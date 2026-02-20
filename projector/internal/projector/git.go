package projector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type GitClient struct {
	timeout time.Duration
	logger  zerolog.Logger
}

func NewGitClient(timeout time.Duration, logger zerolog.Logger) *GitClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &GitClient{
		timeout: timeout,
		logger:  logger,
	}
}

func (g *GitClient) GetStatus(projectPath string) (GitStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()

	var status GitStatus

	branch, err := g.getBranch(ctx, projectPath)
	if err != nil {
		return status, fmt.Errorf("getting branch: %w", err)
	}
	status.Branch = branch

	remote, err := g.getRemote(ctx, projectPath)
	if err != nil {
		g.logger.Debug().Err(err).Str("path", projectPath).Msg("no remote tracking")
		status.Remote = ""
		status.Status = StatusNoRemote
	} else {
		status.Remote = remote
	}

	uncommitted, err := g.getUncommitted(ctx, projectPath)
	if err != nil {
		g.logger.Debug().Err(err).Str("path", projectPath).Msg("getting uncommitted")
	}
	status.Uncommitted = uncommitted

	if remote != "" {
		unpushed, err := g.getUnpushed(ctx, projectPath)
		if err != nil {
			g.logger.Debug().Err(err).Str("path", projectPath).Msg("getting unpushed")
		}
		status.Unpushed = unpushed

		unpulled, err := g.getUnpulled(ctx, projectPath)
		if err != nil {
			g.logger.Debug().Err(err).Str("path", projectPath).Msg("getting unpulled")
		}
		status.Unpulled = unpulled

		status.Status = g.determineStatus(uncommitted, unpushed, unpulled)
	} else if uncommitted > 0 {
		status.Status = StatusDirty
	}

	msg, author, commitTime, err := g.getLastCommit(ctx, projectPath)
	if err != nil {
		g.logger.Debug().Err(err).Str("path", projectPath).Msg("getting last commit")
	} else {
		status.LastCommitMsg = msg
		status.LastCommitAuthor = author
		status.LastCommitTime = commitTime
	}

	return status, nil
}

func (g *GitClient) getBranch(ctx context.Context, path string) (string, error) {
	out, err := g.runGit(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (g *GitClient) getRemote(ctx context.Context, path string) (string, error) {
	out, err := g.runGit(ctx, path, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (g *GitClient) getUncommitted(ctx context.Context, path string) (int, error) {
	out, err := g.runGit(ctx, path, "status", "--porcelain")
	if err != nil {
		return 0, err
	}
	lines := strings.TrimSpace(out)
	if lines == "" {
		return 0, nil
	}
	return len(strings.Split(lines, "\n")), nil
}

func (g *GitClient) getUnpushed(ctx context.Context, path string) (int, error) {
	out, err := g.runGit(ctx, path, "rev-list", "@{u}..HEAD", "--count")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (g *GitClient) getUnpulled(ctx context.Context, path string) (int, error) {
	out, err := g.runGit(ctx, path, "rev-list", "HEAD..@{u}", "--count")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (g *GitClient) getLastCommit(ctx context.Context, path string) (msg, author string, t time.Time, err error) {
	out, err := g.runGit(ctx, path, "log", "-1", "--format=%s|%an|%ct")
	if err != nil {
		return "", "", time.Time{}, err
	}

	parts := strings.SplitN(strings.TrimSpace(out), "|", 3)
	if len(parts) < 3 {
		return "", "", time.Time{}, fmt.Errorf("unexpected log format: %s", out)
	}

	msg = parts[0]
	author = parts[1]
	timestamp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return msg, author, time.Time{}, fmt.Errorf("parsing timestamp: %w", err)
	}
	t = time.Unix(timestamp, 0)

	return msg, author, t, nil
}

func (g *GitClient) determineStatus(uncommitted, unpushed, unpulled int) StatusType {
	if uncommitted > 0 {
		return StatusDirty
	}
	if unpushed > 0 && unpulled > 0 {
		return StatusDiverged
	}
	if unpushed > 0 {
		return StatusAhead
	}
	if unpulled > 0 {
		return StatusBehind
	}
	return StatusClean
}

func (g *GitClient) runGit(ctx context.Context, path string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", args[0], string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}
