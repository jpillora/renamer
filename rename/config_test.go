package rename

import (
	"testing"
)

func TestFindReplace_PlainText(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple replace",
			rule:     "old:new",
			input:    "oldfile.txt",
			expected: "newfile.txt",
			wantErr:  false,
		},
		{
			name:     "no match",
			rule:     "old:new",
			input:    "somefile.txt",
			expected: "somefile.txt",
			wantErr:  false,
		},
		{
			name:     "multiple occurrences - only first replaced",
			rule:     "old:new",
			input:    "oldoldfile.txt",
			expected: "newoldfile.txt",
			wantErr:  false,
		},
		{
			name:     "empty rule",
			rule:     "",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid rule - no colon",
			rule:     "oldnew",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid rule - starts with colon",
			rule:     ":new",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid rule - ends with colon",
			rule:     "old:",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "replace with empty string",
			rule:     "old:",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "complex filename",
			rule:     "_v1:_v2",
			input:    "document_v1.pdf",
			expected: "document_v2.pdf",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fr FindReplace
			err := fr.Set(tt.rule)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindReplace.Set() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FindReplace.Set() unexpected error: %v", err)
				return
			}

			result := fr(tt.input)
			if result != tt.expected {
				t.Errorf("FindReplace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindReplace_RegularExpression(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "basic regex",
			rule:     "/old/new/",
			input:    "oldfile.txt",
			expected: "newfile.txt",
			wantErr:  false,
		},
		{
			name:     "regex with global flag",
			rule:     "/old/new/g",
			input:    "oldoldfile.txt",
			expected: "newnewfile.txt",
			wantErr:  false,
		},
		{
			name:     "regex without global flag - only first match",
			rule:     "/old/new/",
			input:    "oldoldfile.txt",
			expected: "newoldfile.txt",
			wantErr:  false,
		},
		{
			name:     "case insensitive flag",
			rule:     "/OLD/new/i",
			input:    "oldfile.txt",
			expected: "newfile.txt",
			wantErr:  false,
		},
		{
			name:     "case insensitive and global flags",
			rule:     "/OLD/new/ig",
			input:    "OldOLDfile.txt",
			expected: "newnewfile.txt",
			wantErr:  false,
		},
		{
			name:     "regex groups",
			rule:     "/([a-z]+)_([0-9]+)/prefix_$2_$1/",
			input:    "file_123.txt",
			expected: "prefix_123_file.txt",
			wantErr:  false,
		},
		{
			name:     "regex groups with global",
			rule:     "/([a-z])([0-9])/letter$1num$2/g",
			input:    "a1b2c3",
			expected: "letteranum1letterbnum2lettercnum3",
			wantErr:  false,
		},
		{
			name:     "alternative separator",
			rule:     "#old#new#",
			input:    "oldfile.txt",
			expected: "newfile.txt",
			wantErr:  false,
		},
		{
			name:     "alternative separator with flags",
			rule:     "|old|new|gi",
			input:    "OldOLDfile.txt",
			expected: "newnewfile.txt",
			wantErr:  false,
		},
		{
			name:     "invalid regex - missing parts",
			rule:     "/old/",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid regex pattern",
			rule:     "/[/new/",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "unknown flag",
			rule:     "/old/new/x",
			input:    "file.txt",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "digit replacement",
			rule:     "/([0-9]+)/v$1/",
			input:    "file123.txt",
			expected: "filev123.txt",
			wantErr:  false,
		},
		{
			name:     "file extension change",
			rule:     "/\\.txt$/.bak/",
			input:    "document.txt",
			expected: "document.bak",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fr FindReplace
			err := fr.Set(tt.rule)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindReplace.Set() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FindReplace.Set() unexpected error: %v", err)
				return
			}

			result := fr(tt.input)
			if result != tt.expected {
				t.Errorf("FindReplace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindReplace_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		rule     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "regex with no matches",
			rule:     "/xyz/abc/",
			input:    "hello.txt",
			expected: "hello.txt",
			wantErr:  false,
		},
		{
			name:     "regex with empty replacement",
			rule:     "/old//",
			input:    "oldfile.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "plain text with colon in replacement",
			rule:     "old:new:value",
			input:    "oldfile.txt",
			expected: "new:valuefile.txt",
			wantErr:  false,
		},
		{
			name:     "regex group reference beyond available groups",
			rule:     "/([a-z]+)/$1$2/",
			input:    "file.txt",
			expected: "file$2.txt",
			wantErr:  false,
		},
		{
			name:     "empty input",
			rule:     "old:new",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "regex with special characters",
			rule:     "/\\./underscore/",
			input:    "file.txt",
			expected: "fileunderscoretxt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fr FindReplace
			err := fr.Set(tt.rule)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindReplace.Set() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FindReplace.Set() unexpected error: %v", err)
				return
			}

			result := fr(tt.input)
			if result != tt.expected {
				t.Errorf("FindReplace() = %v, want %v", result, tt.expected)
			}
		})
	}
}
