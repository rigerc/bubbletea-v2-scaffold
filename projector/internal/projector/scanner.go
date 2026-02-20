package projector

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Scanner struct {
	rootDir     string
	concurrency int
	gitClient   *GitClient
	logger      zerolog.Logger
}

func NewScanner(rootDir string, concurrency int, gitTimeout time.Duration, logger zerolog.Logger) *Scanner {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &Scanner{
		rootDir:     rootDir,
		concurrency: concurrency,
		gitClient:   NewGitClient(gitTimeout, logger),
		logger:      logger,
	}
}

func (s *Scanner) Scan() ([]Project, error) {
	entries, err := os.ReadDir(s.rootDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(s.rootDir, entry.Name()))
		}
	}

	jobs := make(chan string, len(dirs))
	results := make(chan Project, len(dirs))
	errors := make(chan error, len(dirs))

	var wg sync.WaitGroup
	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go s.worker(jobs, results, errors, &wg)
	}

	for _, dir := range dirs {
		jobs <- dir
	}
	close(jobs)

	wg.Wait()
	close(results)
	close(errors)

	var projects []Project
	for project := range results {
		projects = append(projects, project)
	}

	for err := range errors {
		s.logger.Debug().Err(err).Msg("scan error")
	}

	return projects, nil
}

func (s *Scanner) worker(jobs <-chan string, results chan<- Project, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for dir := range jobs {
		project, err := s.scanProject(dir)
		if err != nil {
			errors <- err
			continue
		}
		if project != nil {
			results <- *project
		}
	}
}

func (s *Scanner) scanProject(dir string) (*Project, error) {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	project := &Project{
		Name: filepath.Base(dir),
		Path: dir,
	}

	gitStatus, err := s.gitClient.GetStatus(dir)
	if err != nil {
		s.logger.Debug().Err(err).Str("path", dir).Msg("git status error")
	}
	project.Git = gitStatus

	return project, nil
}
