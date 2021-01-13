# dotfiles

A tool to manage dotfiles with zero dependencies

Uses https://github.com/go-git/go-git instead of git

## usage

```sh
# download a release onto your path

# output help
$ dotfiles --help

# clone repo https://github.com/stuart-warren/.dotfiles.git to $HOME/.dotfiles
# if repository contains an executable script `/install`, run that
# if repository contains `/dotfiles.yaml` process and execute that config (see below)
$ dotfiles stuart-warren
# same for https://github.com/stuart-warren/my-dot-files.git to $HOME/.dotfiles
$ dotfiles stuart-warren/my-dot-files
# same for explicit repo url to $HOME/.dotfiles
$ dotfiles https://gitlab.com/stuart-warren/somerepo.git

# if $HOME/.dotfiles directory already exists and all files are unmodified, pull latest version, clean up if possible and rerun install process
$ dotfiles

```

# config

Handles differences in OS darwin/linux
```yaml

```