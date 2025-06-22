# Copilot Instructions for renamer project

## Project Overview
This is a Go-based batch file renaming tool that supports both plain text and regular expression-based find/replace operations.

## Project Structure
- `main.go` - CLI entry point with opts configuration
- `rename/config.go` - Configuration struct and FindReplace type with parsing logic
- `rename/execute.go` - Core execution logic for file operations
- `rename/config_test.go` - Comprehensive unit tests for FindReplace functionality

## Key Design Patterns
- Value semantics: Functions accept `Config` values, not pointers
- Package separation: CLI logic in main, core logic in rename package
- Comprehensive testing: All documented rule formats are tested

## Important Notes for Future Development

### Binary Handling
- The main binary `renamer` is gitignored - DO NOT commit it
- When creating test files or sample binaries, ALWAYS write them to `/tmp` directory
- Example: `/tmp/test_renamer` or `/tmp/sample_files/`

### Testing Strategy
- Always run tests after making changes: `go test ./rename -v`
- Test both value and pointer scenarios when modifying Config usage
- Verify both dryrun and actual execution modes work

### Rule Format Support
The tool supports two main rule formats (see README for full documentation):
1. Plain text: `<find>:<replace>`
2. Regex: `/<find>/<replace>/<flags>` where flags can be `i` (case-insensitive) and `g` (global)

### Directory Creation
- When `--fullpath` flag is used, missing directories are automatically created with `os.MkdirAll`
- This happens during the perform phase, not validation phase

### Common Commands for Development
```bash
# Build (creates binary in current dir - will be gitignored)
go build

# Run tests
go test ./rename -v

# Test basic functionality (always use /tmp for test files)
echo "test" > /tmp/testfile.txt
./renamer --dryrun "test:demo" /tmp/testfile.txt

# Test fullpath with directory creation
./renamer --dryrun --fullpath "testfile:/tmp/newdir/moved" /tmp/testfile.txt
```

### Code Style Notes
- Use value semantics for Config (not pointers)
- Keep CLI concerns in main.go, business logic in rename package
- Always include comprehensive error handling
- Follow Go naming conventions (DefaultConfig, not NewConfig)
