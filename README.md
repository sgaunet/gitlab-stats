[![Go Report Card](https://goreportcard.com/badge/github.com/sgaunet/gitlab-stats)](https://goreportcard.com/report/github.com/sgaunet/gitlab-stats)
[![GitHub release](https://img.shields.io/github/release/sgaunet/gitlab-stats.svg)](https://github.com/sgaunet/gitlab-stats/releases/latest)
![GitHub Downloads](https://img.shields.io/github/downloads/sgaunet/gitlab-stats/total)
![coverage](https://raw.githubusercontent.com/wiki/sgaunet/gitlab-stats/coverage-badge.svg)
[![Linter](https://github.com/sgaunet/gitlab-stats/actions/workflows/linter.yml/badge.svg)](https://github.com/sgaunet/gitlab-stats/actions/workflows/linter.yml)
[![Coverage Badge Generation](https://github.com/sgaunet/gitlab-stats/actions/workflows/coverage.yml/badge.svg)](https://github.com/sgaunet/gitlab-stats/actions/workflows/coverage.yml)
[![Snapshot](https://github.com/sgaunet/gitlab-stats/actions/workflows/snapshot.yml/badge.svg)](https://github.com/sgaunet/gitlab-stats/actions/workflows/snapshot.yml)
[![Release](https://github.com/sgaunet/gitlab-stats/actions/workflows/release.yml/badge.svg)](https://github.com/sgaunet/gitlab-stats/actions/workflows/release.yml)
[![License](https://img.shields.io/github/license/sgaunet/gitlab-stats.svg)](LICENSE)

# gitlab-stats

gitlab-stats is a tool to register stats of gitlab projects/groups. Based on the statistics saved, it can generate a graph to visualize the activity on gitlab projects/groups.

Example:

![screenshot](doc/screenshot.png)

Actually, the stats are saved in $HOME/.gitlab-stats/db.json.

To register stats, you need to add cron, for example: 

```
00 00 * * * GITLAB_TOKEN=.... /usr/local/bin/gitlab-stats -g <groupID>   # comment
```

To generate the screenshot, you can also add a cron or execute it in the command line. Example of a cron:

```
00 00 1 * * /usr/local/bin/gitlab-stats -g <groupID> -o stats-`date "+%Y-%m" -d "1 day ago"`.png
```


# Usage

```
$ gitlab-stats -h
Usage of gitlab-stats:
  -d string
        Debug level (info,warn,debug) (default "error")
  -g int
        Group ID to get issues from (not compatible with -p option)
  -o string
        file path to generate statistic graph (do not fulfill DB)
  -p int
        Project ID to get issues from
  -s int
        since (default 6)
  -v    Get version
```

### System Dependencies

This project uses CGO to interface with SQLite, so you need the SQLite development libraries installed.

**Ubuntu/Debian:**
```bash
sudo apt install sqlite3 sqlite3-tools libsqlite3-dev
```

**Red Hat/CentOS/Fedora:**
```bash
sudo yum install sqlite sqlite-devel
# or on newer systems:
sudo dnf install sqlite sqlite-devel
```

# Development

## prerequisites

This project is using :

* golang
* [task for development](https://taskfile.dev/#/)
* [goreleaser](https://goreleaser.com/)
* [pre-commit](https://pre-commit.com/)

There are hooks executed in the precommit stage. Once the project cloned on your disk, please install pre-commit:

```
brew install pre-commit
```

Install tools:

```
task install-prereq
```

And install the hooks:

```
task install-pre-commit
```

If you like to launch manually the pre-commmit hook:

```
task pre-commit
```

## 🕐 Project Status: Low Priority

This project is not under active development. While the project remains functional and available for use, please be aware of the following:

### What this means:
- **Response times will be longer** - Issues and pull requests may take weeks or months to be reviewed
- **Updates will be infrequent** - New features and non-critical bug fixes will be rare
- **Support is limited** - Questions and discussions may not receive timely responses

### We still welcome:
- 🐛 **Bug reports** - Critical issues will eventually be addressed
- 🔧 **Pull requests** - Well-tested contributions are appreciated
- 💡 **Feature requests** - Ideas will be considered for future development cycles
- 📖 **Documentation improvements** - Always helpful for the community

### Before contributing:
1. **Check existing issues** - Your concern may already be documented
2. **Be patient** - Responses may take considerable time
3. **Be self-sufficient** - Be prepared to fork and maintain your own version if needed
4. **Keep it simple** - Small, focused changes are more likely to be merged

### Alternative options:
If you need active support or rapid development:
- Look for actively maintained alternatives
- Reach out to discuss taking over maintenance

We appreciate your understanding and patience. This project remains important to us, but current priorities limit our ability to provide regular updates and support.