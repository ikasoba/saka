<h1><p align=center><img width=48 height=48 src="https://github.com/user-attachments/assets/d3a2e249-05f5-4857-9d1b-0ff2745f08ca" /> Saka</p></h1>
<p align=center>
A Cute & Easy TUI text editor inspired by GNU Nano.

<a href="https://pkg.go.dev/github.com/ikasoba/saka"><img src="https://pkg.go.dev/badge/github.com/ikasoba/saka.svg" alt="Go Reference"></a>
<a href="https://github.com/ikasoba/saka/actions/workflows/release.yml"><img src="https://github.com/ikasoba/saka/actions/workflows/release.yml/badge.svg?event=release" alt="release"></a>

![Screenshot_20250126_151937](https://github.com/user-attachments/assets/6f18a0ba-885e-433b-a80b-2c265ab4f5d1)

</p>

# Install from go
```sh
go install github.com/ikasoba/saka@latest
```

## If you have problems installing on Windows

Problems often occur if gcc is not installed.

In that case, installing [tdm-gcc](https://jmeubank.github.io/tdm-gcc/) or specifying your preferred compiler in the CC, CXX environment variable may solve the problem.

(In some cases you may need to set CGO_ENABLED)

# Setup
In the initial state, no configuration file exists, so a default configuration file must be obtained.
```sh
saka --fetch-default-config
```

The default configuration file has a tree-sitter parser for golang configured, and this parser must be installed in your editor to use highlighting.
(However, the tree-sitter cli is required).
```sh
saka --install-grammar
```

You can check the location of the configuration file with the following command.
```sh
saka -e
```

# Releases

https://github.com/ikasoba/saka/releases

# Usage

```
Usage: saka <filename>
  Open file

Usage: saka (-e|--show-environment)
  Show environment variables

Usage: saka --install-grammar
  Install grammar from config

Usage: saka --fetch-default-config
  Fetch default config
```
