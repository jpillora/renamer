package rename

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Config holds all configuration options for the renamer tool
type Config struct {
	Recursive bool        `help:"recurse into directories instead of renaming them"`
	Dryrun    bool        `help:"read-only mode for testing renames"`
	Fullpath  bool        `help:"find and replace against the targets full path"`
	Overwrite bool        `help:"allow existing files to be overwritten"`
	Verbose   bool        `help:"enable verbose logs"`
	Limit     int         `help:"rename limit tries to prevent you from renaming your whole file system"`
	Rule      FindReplace `opts:"mode=arg"`
	Targets   []string    `opts:"mode=arg, min=1"`
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() Config {
	return Config{
		Recursive: false,
		Dryrun:    false,
		Limit:     1000,
	}
}

// FindReplace is a function type that transforms input strings
type FindReplace func(string) string

// Set implements opts.Setter for the FindReplace type
func (fr *FindReplace) Set(s string) error {
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

func (fr *FindReplace) setRegExp(pattern, replace, flags string) error {
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
		pattern = "(?i)" + pattern
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
