package rename

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jpillora/ansi"
	"golang.org/x/crypto/ssh/terminal"
)

// Execute runs the renamer with the given config
func Execute(config Config) error {
	//dryrun is read-only + verbose
	if config.Dryrun {
		config.Verbose = true
	}
	//resolve targets into moves
	moves, err := resolveAll(config, config.Targets)
	if err != nil {
		return err
	}
	//validate all moves
	if err := moves.validate(config); err != nil {
		return err
	}
	verbf(config, "resolved %d targets into %d validated moves", len(config.Targets), len(moves))
	//perform all moves
	if err := moves.perform(config); err != nil {
		return err
	}
	if config.Dryrun {
		verbf(config, "dryrun succesfully performed")
	} else {
		verbf(config, "moves succesfully performed")
	}
	return nil
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

func (m move) perform(config Config) error {
	action := "dryrun move"
	if !config.Dryrun {
		action = "moved"
		// Create directories if --fullpath is used
		if config.Fullpath {
			dir := filepath.Dir(m.dst)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
		if err := os.Rename(m.src, m.dst); err != nil {
			return err
		}
	}
	verbf(config, "%s %s", action, m.String())
	return nil
}

type moves []*move

func (ms moves) validate(config Config) error {
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
	// Note: When using --fullpath, we automatically create missing directories
	// during the perform step, so we don't need to validate their existence here
	return nil
}

func (ms moves) perform(config Config) error {
	for _, m := range ms {
		if err := m.perform(config); err != nil {
			return err
		}
	}
	return nil
}

func resolve(config Config, target string) (moves, error) {
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
		return resolveAll(config, targets)
	}
	//skip irregular files
	if !s.Mode().IsRegular() && !s.IsDir() {
		verbf(config, "skip irregular file: %s", target)
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
		verbf(config, "skip no-op target: %s", target)
		return nil, nil
	}
	return moves{m}, nil
}

func resolveAll(config Config, targets []string) (moves, error) {
	combined := moves{}
	for _, target := range targets {
		set, err := resolve(config, target)
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

func verbf(config Config, format string, args ...interface{}) {
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
