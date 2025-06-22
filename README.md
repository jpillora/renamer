# renamer

`renamer` is a batch file renaming tool

[![GoDev](https://img.shields.io/static/v1?label=godoc&message=reference&color=00add8)](https://pkg.go.dev/github.com/jpillora/renamer)
[![CI](https://github.com/jpillora/renamer/workflows/CI/badge.svg)](https://github.com/jpillora/renamer/actions?workflow=CI)

### Features

* Simple
* Statically compiled binaries
* Regular expression match and replace

### Install

**Binaries**

[![Releases](https://img.shields.io/github/release/jpillora/renamer.svg)](https://github.com/jpillora/renamer/releases)
[![Releases](https://img.shields.io/github/downloads/jpillora/renamer/total.svg)](https://github.com/jpillora/renamer/releases)

Find [the latest pre-compiled binaries here](https://github.com/jpillora/renamer/releases/latest) or download and install it now with `curl https://i.jpillora.com/renamer! | bash`

**Source**

```sh
$ go install -v github.com/jpillora/renamer@latest
```

### Usage

<!--tmpl,code=plain:echo "$ renamer --help" && go run main.go --help | cat -->
``` plain 
$ renamer --help

  Usage: renamer [options] <rule> <target> [target] ...

  renamer is a regular-expression based batch file renaming tool

  <rule> should be in the form:

    <find>:<replace>           for plain text matching, or
    /<find>/<replace>/<flags>  for regular expression matching

    by default, <find> and <replace> only operates on the input
    file's basename (the last part of the file). this ensures
    files are *renamed* and not *moved*. if you'd like to operate
    on the full (absolute) path, you can enable the --fullpath flag.

    plain text matching only finds and replaces the first instance
    of <find>. if you need to replace all instances, use regular
    expression matching with the 'g' flag.

    regular expression matching uses the go regular expression
    engine (https://golang.org/pkg/regexp/). the format above uses
    slash / as the field separator, however any character may be
    used, as long it appears exactly 3 times. regular expression
    groups may also be used, and replaced back into the result
    using $N placeholders ($1 for the first group, $2 for second, etc).

    regular expression flags can be:
      i - enables case-insensitive matching
      g - enables global matching, instead of the
          default single-instance matching

  <target> must refer to an existing file or directory. by default
  a directory target will be renamed, however the --recursive flag
  will change this behaviour to recurse into directories and use the
  directory contents as more targets.

  Options:
  --recursive, -r  recurse into directories instead of renaming them
  --dryrun, -d     read-only mode for testing renames
  --fullpath, -f   find and replace against the targets full path
  --overwrite, -o  allow existing files to be overwritten
  --verbose, -v    enable verbose logs
  --limit, -l      rename limit tries to prevent you from renaming your
                   whole file system (default 1000)
  --version        display version
  --help, -h       display help

  Version:
    0.0.0-src

  Read more:
    github.com/jpillora/renamer

```
<!--/tmpl-->

### Credits

`renamer` in Node https://www.npmjs.com/package/renamer