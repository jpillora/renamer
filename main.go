package main

import (
	"log"

	"github.com/jpillora/opts"
	"github.com/jpillora/renamer/rename"
)

var version = "0.0.0-src"

const argHelp = `
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
`

func main() {
	c := rename.DefaultConfig()
	opts.New(&c).
		Name("renamer").
		Summary("renamer is a regular-expression based batch file renaming tool").
		DocAfter("summary", "arg-help", argHelp).
		SetLineWidth(68).
		Version(version).
		Repo("github.com/jpillora/renamer").
		Parse()

	if err := rename.Execute(c); err != nil {
		log.Fatal(err)
	}
}
