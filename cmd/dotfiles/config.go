package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
)

const (
	unsupported = "Notice: This feature is unsupported on the current system: %s"
	noWork      = "Notice: No work to do for feature: %s"
)

// Config is structure of configuration file dotfiles.yaml
type Config struct {
	AptRepositories  []AptRepo `yaml:"apt_repositories"`
	ApkRepositories  []ApkRepo `yaml:"apk_repositories"`
	AptPackages      []string  `yaml:"apt_packages"`
	ApkPackages      []string  `yaml:"apk_packages"`
	HomebrewFormula  []string
	HomebrewCask     []string
	LinuxbrewFormula []string
	LinuxbrewCask    []string
	GitRepositories  []GitRepo
	InstallerURLs    []string
	AsdfPlugins      []string
	Links            []Link
	Commands         []string
}

type Link struct {
	Link   string
	Target string
	Force  bool
}

type AptRepo struct {
	KeyURL       string   `yaml:"key_url"`
	Repositories []string `yaml:"repositories"`
}

type ApkRepo struct {
	KeyURL       string
	Repositories []string
}

type GitRepo struct {
	URL      string
	Branch   string
	DestPath string
}

func (r Runner) logUnsupported(feature string) {
	r.logger.Printf(unsupported, feature)
}

func (r Runner) logNoWork(feature string) {
	r.logger.Printf(noWork, feature)
}

func (r *Runner) RunCommand(name string, args ...string) error {
	u, _ := user.Current()
	if u.Uid == "0" && name == "sudo" {
		name = args[0]
		args = args[1:]
	}
	cmd := exec.Command(name, args...)
	cmd.Env = []string{
		"CI=true",
		"DEBIAN_FRONTEND=noninteractive",
	}
	cmd.Stdin = r.stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	err := cmd.Run()
	if err != nil {
		r.logger.Printf("command failed: %s", cmd.String())
		return err
	}
	return nil
}

func (r *Runner) AptPrerequisites() error {
	_, err := exec.LookPath("apt-get")
	if err != nil {
		return err
	}
	err = r.RunCommand("sudo", "apt-get", "update")
	if err != nil {
		return err
	}
	r.RunCommand("sudo", "apt-get", "install", "-y", "--no-install-recommends", "bash", "curl", "wget", "ca-certificates", "apt-transport-https", "gpg", "gpg-agent", "sudo")
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) Apt(cfg Config) error {
	feature := "apt"
	_, err := exec.LookPath("apt-get")
	if err != nil {
		r.logUnsupported(feature)
		return nil
	}
	_, err = os.Stat("/etc/apt/sources.list")
	if os.IsNotExist(err) {
		r.logUnsupported(feature)
		return nil
	}
	if len(cfg.AptRepositories) == 0 && len(cfg.AptPackages) == 0 {
		r.logNoWork(feature)
	}
	// TODO: keep in mind https://www.linuxuprising.com/2021/01/apt-key-is-deprecated-how-to-add.html
	r.RunCommand("sudo", "rm", "/etc/apt/sources.list.d/dotfiles.list")
	dotfilesListContents, err := os.Create("/tmp/dotfiles.list")
	if err != nil {
		return err
	}
	fmt.Fprintf(dotfilesListContents, "# created by dotfiles\n")
	for _, aptRepository := range cfg.AptRepositories {
		r.RunCommand("wget", "-O", "/tmp/repo.key", aptRepository.KeyURL)
		r.RunCommand("sudo", "apt-key", "add", "/tmp/repo.key")
		r.RunCommand("rm", "/tmp/repo.key")
		for _, repository := range aptRepository.Repositories {
			fmt.Fprintf(dotfilesListContents, "%s\n", repository)
		}
	}
	dotfilesListContents.Close()
	r.RunCommand("sudo", "mv", "/tmp/dotfiles.list", "/etc/apt/sources.list.d/dotfiles.list")
	r.RunCommand("sudo", "apt-get", "update")
	for _, pkg := range cfg.AptPackages {
		err := r.RunCommand("sudo", "apt-get", "install", "-y", "--no-install-recommends", pkg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Apk(cfg Config) error {
	feature := "apk"
	_, err := exec.LookPath("apk")
	if err != nil {
		r.logUnsupported(feature)
		return nil
	}
	_, err = os.Stat("/etc/apk/repositories")
	if os.IsNotExist(err) {
		r.logUnsupported(feature)
		return nil
	}
	if len(cfg.ApkRepositories) == 0 && len(cfg.ApkPackages) == 0 {
		r.logNoWork(feature)
	}
	// TODO
	return nil
}

func (r *Runner) Homebrew(cfg Config) error {
	feature := "homebrew"
	if runtime.GOOS != "darwin" {
		r.logUnsupported(feature)
		return nil
	}
	// TODO
	return nil
}

func (r *Runner) Linuxbrew(cfg Config) error {
	feature := "linuxbrew"
	if runtime.GOOS != "linux" {
		r.logUnsupported(feature)
		return nil
	}
	// TODO
	return nil
}
