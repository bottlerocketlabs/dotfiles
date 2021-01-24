package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

func splitPath(path string) (org, name string) {
	org = ""
	name = ""
	pathParts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
	if len(pathParts) == 2 {
		org = pathParts[0]
		name = pathParts[1]
	}
	if len(pathParts) == 1 {
		org = pathParts[0]
	}
	return org, name
}

func parseRepo(arg string) (scheme, host, org, name string) {
	u, err := url.Parse(arg)
	if err == nil {
		org, name := splitPath(u.Path)
		return u.Scheme, u.Host, org, name
	}
	dir, file := path.Split(arg)
	if dir == "" {
		return "", "", dir, "dotfiles"
	}
	return "", "", dir, file
}

func buildRepo(arg string) url.URL {
	scheme := "https"
	repoHost := "github.com"
	repoOrg := "stuart-warren"
	repoName := "dotfiles"
	s, h, o, n := parseRepo(arg)
	if s != "" {
		scheme = s
	}
	if h != "" {
		repoHost = h
	}
	if o != "" {
		repoOrg = o
	}
	if n != "" {
		repoName = n
	}
	return url.URL{
		Scheme: scheme,
		Host:   repoHost,
		Path:   fmt.Sprintf("/%s/%s", repoOrg, repoName),
	}
}

// Env is abstracted environment
type Env struct {
	m map[string]string
}

// Get an environment variable by key, or blank string if missing
func (e *Env) Get(key string) string {
	value, ok := e.m[key]
	if !ok {
		return ""
	}
	return value
}

// NewEnv creates a new env from = separated string slice (eg: os.Environ())
func NewEnv(environ []string) Env {
	e := make(map[string]string)
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		e[parts[0]] = parts[1]
	}
	return Env{m: e}
}

func main() {
	err := NewRunner(os.Args, NewEnv(os.Environ()), os.Stdin, os.Stdout, os.Stderr).Run()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

// CloneOrOpenGitRepo clones a repo if directory does not exist already, if it does exist return it.
func CloneOrOpenGitRepo(path string, repoURL url.URL, out io.Writer) (*git.Repository, error) {
	var repo *git.Repository
	if _, err := os.Stat(path); os.IsNotExist(err) {
		repo, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:               repoURL.String(),
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Progress:          out,
		})
		if err != nil {
			return repo, fmt.Errorf("error cloning repository: %w", err)
		}
	} else {
		repo, err = git.PlainOpen(path)
		if err != nil {
			return repo, fmt.Errorf("%s exists but failed to load it as a git repository: %w", path, err)
		}
	}
	return repo, nil
}

// Runner holds startup state
type Runner struct {
	args     []string
	env      Env
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
	readFile func(path string) ([]byte, error)
}

// NewRunner creates a Runner
func NewRunner(args []string, env Env, stdin io.Reader, stdout, stderr io.Writer) Runner {
	return Runner{args: args, env: env, stdin: stdin, stdout: stdout, stderr: stderr, readFile: ioutil.ReadFile}
}

// Run is the main thread but separated out so easier to test
func (r Runner) Run() error {
	flags := flag.NewFlagSet(r.args[0], flag.ExitOnError)
	err := flags.Parse(r.args[1:])
	if err != nil {
		return err
	}
	logger := log.New(r.stderr, "", log.LstdFlags|log.Lshortfile)
	var repoArg string
	if len(r.args) > 1 {
		repoArg = r.args[1]
	}
	u := buildRepo(repoArg)
	logger.Printf("have url: %s", u.String())
	dotfilesEnv := r.env.Get("DOTFILES_DIR")
	dotfilesDir, err := homedir.Expand(dotfilesEnv)
	if err != nil && dotfilesEnv != "" {
		return fmt.Errorf("could not expand DOTFILES_DIR from environment: %w", err)
	}
	homeDir, err := homedir.Dir()
	if err != nil && dotfilesDir == "" {
		return fmt.Errorf("can't get dir for dotfiles: %w", err)
	}
	if dotfilesDir == "" {
		dotfilesDir = path.Join(homeDir, ".dotfiles")
	}
	logger.Printf("have dir: %s", dotfilesDir)
	repo, err := CloneOrOpenGitRepo(dotfilesDir, u, r.stderr)
	if err != nil {
		return fmt.Errorf("failed to get git repo: %w", err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("could not get dotfiles worktree: %w", err)
	}
	err = wt.Pull(&git.PullOptions{Progress: r.stderr})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("could not pull dotfiles repository: %w", err)
	}
	err = r.handleDotfilesConfigFile(dotfilesDir)
	if err != nil {
		logger.Printf("could not find /dotfiles.yaml in the repository")
		logger.Printf("trying standard install scripts instead")
		return r.handleInstallScript(dotfilesDir)
	}
	return nil
}

// Config is structure of configuration file dotfiles.yaml
type Config struct {
	// TODO
}

func (r Runner) handleDotfilesConfigFile(dotfilesDir string) error {
	dotfilesConfigFile := path.Join(dotfilesDir, "dotfiles.yaml")
	_, err := os.Stat(dotfilesConfigFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("dotfiles config file does not exist: %s", dotfilesConfigFile)
	}
	contents, err := r.readFile(dotfilesConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	var cfg Config
	err = yaml.Unmarshal(contents, &cfg)
	if err != nil {
		return fmt.Errorf("could not parse config file: %w", err)
	}
	return nil
}

func (r Runner) handleInstallScript(dotfilesDir string) error {
	dotfilesInstallScript := path.Join(dotfilesDir, "install")
	startScripts := []string{"install", "install.sh", "bootstrap", "bootstrap.sh", "setup", "setup.sh"}
	for _, script := range startScripts {
		fullScriptPath := path.Join(dotfilesDir, script)
		_, err := os.Stat(fullScriptPath)
		if os.IsNotExist(err) {
			continue
		}
		dotfilesInstallScript = fullScriptPath
		break
	}
	if _, err := os.Stat(dotfilesInstallScript); os.IsNotExist(err) {
		return fmt.Errorf("dotfile install script does not exist: %s", dotfilesInstallScript)
	}
	cmd := exec.Command(dotfilesInstallScript)
	cmd.Stdin = r.stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run install script: %w", err)
	}
	return nil
}
