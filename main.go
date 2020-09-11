package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jpillora/ansi"
	"github.com/jpillora/opts"
	"golang.org/x/crypto/ssh/terminal"
)

var version = "0.0.0-src"

var config = struct {
	Recursive bool        `help:"recurse into directories instead of renaming them"`
	Dryrun    bool        `help:"read-only mode for testing renames"`
	Fullpath  bool        `help:"find and replace against the targets full path"`
	Overwrite bool        `help:"allow existing files to be overwritten"`
	Verbose   bool        `help:"enable verbose logs"`
	Limit     int         `help:"rename limit tries to prevent you from renaming your whole file system"`
	Rule      findReplace `opts:"mode=arg"`
	Targets   []string    `opts:"mode=arg, min=1"`
}{
	Recursive: false,
	Dryrun:    false,
	Limit:     1000,
}

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
	opts.New(&config).
		Name("renamer").
		Summary("renamer is a regular-expression based batch file renaming tool").
		DocAfter("summary", "arg-help", argHelp).
		SetLineWidth(68).
		Version(version).
		Repo("github.com/jpillora/renamer").
		Parse()
	//dryrun is read-only + verbose
	if config.Dryrun {
		config.Verbose = true
	}
	//resolve targets into moves
	moves, err := resolveAll(config.Targets)
	if err != nil {
		log.Fatal(err)
	}
	//validate all moves
	if err := moves.validate(); err != nil {
		log.Fatal(err)
	}
	verbf("resolved %d targets into %d validated moves", len(config.Targets), len(moves))
	//perform all moves
	if err := moves.perform(); err != nil {
		log.Fatal(err)
	}
	if config.Dryrun {
		verbf("dryrun succesfully performed")
	} else {
		verbf("moves succesfully performed")
	}
}

type move struct {
	src string
	dst string
}

var wd, _ = os.Getwd()

func (m move) String() string {
	s := shorten(m.src)
	d := shorten(m.dst)
	if path, s, d := trimPathPrefix(s, d); path != "" {
		return fmt.Sprintf("%s/{%s -> %s}", blue(path), red(s), green(d))
	}
	return fmt.Sprintf("%s -> %s", red(s), green(d))
}

func (m move) perform() error {
	action := "dryrun move"
	if !config.Dryrun {
		action = "moved"
		if err := os.Rename(m.src, m.dst); err != nil {
			return err
		}
	}
	verbf("%s %s", action, m.String())
	return nil
}

type moves []*move

func (ms moves) validate() error {
	dupes := map[string]string{}
	for _, m := range ms {
		if s1 := dupes[m.dst]; s1 != "" {
			s2 := m.src
			return fmt.Errorf("two targets (%s and %s) rename to the same path", shorten(s1), shorten(s2))
		}
		dupes[m.dst] = m.src
	}
	if !config.Overwrite {
		for _, m := range ms {
			if _, err := os.Stat(m.dst); err == nil {
				return fmt.Errorf("move %s would overwrite an existing file, set the --overwrite flag to force", m.String())
			}
		}
	}
	if config.Fullpath {
		dirs := map[string]bool{}
		for _, m := range ms {
			d := filepath.Dir(m.dst)
			if dirs[d] {
				continue
			}
			if s, err := os.Stat(d); err != nil {
				return fmt.Errorf("move %s cannot be performed, parent directory does not exist", m.String())
			} else if !s.IsDir() {
				return fmt.Errorf("move %s cannot be performed, parent path is not directory", m.String())
			}
			dirs[d] = true
		}
	}
	return nil
}

func (ms moves) perform() error {
	for _, m := range ms {
		if err := m.perform(); err != nil {
			return err
		}
	}
	return nil
}

func resolve(target string) (moves, error) {
	s, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	//recurse into dir?
	if s.IsDir() && config.Recursive {
		ss, err := ioutil.ReadDir(target)
		if err != nil {
			return nil, err
		}
		targets := make([]string, len(ss))
		for i, s := range ss {
			targets[i] = filepath.Join(target, s.Name())
		}
		return resolveAll(targets)
	}
	//skip irregular files
	if !s.Mode().IsRegular() && !s.IsDir() {
		verbf("skip irregular file: %s", target)
		return nil, nil
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}
	m := &move{
		src: abs,
	}
	if config.Fullpath {
		m.dst = config.Rule(abs)
	} else {
		dir, file := filepath.Split(abs)
		m.dst = filepath.Join(
			dir,
			config.Rule(file),
		)
	}
	//noop?
	if m.src == m.dst {
		verbf("skip no-op target: %s", target)
		return nil, nil
	}
	return moves{m}, nil
}

func resolveAll(targets []string) (moves, error) {
	combined := moves{}
	for _, target := range targets {
		set, err := resolve(target)
		if err != nil {
			return nil, err
		}
		if len(set) > 0 {
			combined = append(combined, set...)
		}
	}
	if len(combined) > config.Limit {
		return nil, errors.New("surpassed rename limit")
	}
	return combined, nil
}

type findReplace func(string) string

//Set implements opts.Setter
func (fr *findReplace) Set(s string) error {
	if s == "" {
		return errors.New("empty rename-rule")
	}
	//regex?
	p := strings.Split(s, s[0:1])
	if len(p) == 4 && p[0] == "" && p[1] != "" {
		return fr.setRegExp(p[1], p[2], p[3])
	}
	//plain-text
	i := strings.Index(s, ":")
	if i == -1 || i == 0 || i == len(s)-1 {
		return errors.New("invalid plain-text rename-rule")
	}
	//find and replace first instance
	find := s[:i]
	replace := s[i+1:]
	*fr = func(input string) string {
		return strings.Replace(input, find, replace, 1)
	}
	return nil
}

var groups = regexp.MustCompile(`\$\d+`)

func (fr *findReplace) setRegExp(pattern, replace, flags string) error {
	//regexp flags
	global := false
	ignoreCase := false
	for _, r := range flags {
		switch r {
		case 'g':
			global = true
		case 'i':
			ignoreCase = true
		default:
			return fmt.Errorf("unknown regex flag: %s", string(r))
		}
	}
	//go-specific regexp ignore case
	if ignoreCase {
		pattern = "(i?)" + pattern
	}
	//compile pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex: %s (%s)", pattern, err)
	}
	//replace once
	*fr = func(s string) string {
		first := true
		return re.ReplaceAllStringFunc(s, func(input string) string {
			if !first && !global {
				return input
			}
			first = false
			//get matched groups
			m := re.FindStringSubmatch(input)
			if len(m) == 0 {
				panic("how did we get here")
			}
			//return replace, with groups swapped out
			return groups.ReplaceAllStringFunc(replace, func(group string) string {
				i, _ := strconv.Atoi(group[1:])
				if i >= len(m) {
					return group
				}
				return m[i]
			})
		})
	}
	return nil
}

func verbf(format string, args ...interface{}) {
	if config.Verbose {
		log.Printf(format, args...)
	}
}

var isaTTY = terminal.IsTerminal(int(os.Stdout.Fd()))

func color(attr ansi.Attribute) func(string) string {
	if !isaTTY {
		return func(s string) string {
			return s
		}
	}
	col := string(ansi.Set(attr))
	reset := string(ansi.ResetBytes)
	return func(s string) string {
		return col + s + reset
	}
}

var grey = color(ansi.Black)
var green = color(ansi.Green)
var red = color(ansi.Red)
var blue = color(ansi.Blue)

func shorten(path string) string {
	if len(path)+1 > len(wd) && strings.HasPrefix(path, wd+"/") {
		path = strings.TrimPrefix(path, wd+"/")
	}
	return path
}

func dots(path string) string {
	return strings.Replace(path, " ", "·", -1)
}

func contains(set []string, item string) bool {
	for _, s := range set {
		if s == item {
			return true
		}
	}
	return false
}

const sep = string(filepath.Separator)

func trimPathPrefix(pathA, pathB string) (string, string, string) {
	//TODO: add suffix too for: foo/bar/{bazz->zip}/ping/pong
	partsT := []string{}
	partsA := strings.Split(pathA, sep)
	partsB := strings.Split(pathB, sep)
	for len(partsA) > 0 && len(partsB) > 0 {
		a := partsA[0]
		b := partsB[0]
		if a != b {
			break
		}
		partsT = append(partsT, a)
		partsA = partsA[1:]
		partsB = partsB[1:]
	}
	pathTrimmed := strings.Join(partsT, sep)
	pathA = strings.Join(partsA, sep)
	pathB = strings.Join(partsB, sep)
	return dots(pathTrimmed), dots(pathA), dots(pathB)
}
