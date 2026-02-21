package analyzer

import (
	"strings"
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

// TestFormat_NoStrayBraceRuleset checks that various nested CSS
// patterns do not produce a phantom "} {}" in the output.
func TestFormat_NoStrayBraceRuleset(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "media with nested ruleset",
			src: `@media (max-width: 768px) {
  .list {
    .item {
      display: flex;
    }
  }
}`,
		},
		{
			name: "triple nested rulesets",
			src: `.portfolio-list {
  .item {
    &:hover {
      opacity: 0.8;
    }
  }
}`,
		},
		{
			name: "media with double nested rulesets",
			src: `@media (max-width: 768px) {
  .portfolio-list {
    .item {
      &:hover {
        opacity: 0.8;
      }
    }
  }
}`,
		},
		{
			name: "nested at-rule inside nested ruleset",
			src: `.portfolio-list {
  display: grid;

  .item {
    padding: 1rem;

    @media (max-width: 768px) {
      padding: 0.5rem;
    }
  }
}`,
		},
		{
			name: "multiple nested rulesets",
			src: `.portfolio-list {
  display: grid;

  .item {
    color: red;
  }

  .item:hover {
    color: blue;
  }
}`,
		},
		{
			name: "empty nested ruleset",
			src: `.parent {
  .child {}
}`,
		},
		{
			name: "deeply nested 4 levels",
			src: `@layer components {
  @media (max-width: 768px) {
    .portfolio-list {
      .item {
        display: flex;
      }
    }
  }
}`,
		},
		{
			name: "compact nested with no newlines",
			src:  `.a{.b{.c{color:red;}}}`,
		},
		{
			name: "single-line nested",
			src:  `.a { .b { color: red; } }`,
		},
		{
			name: "nested with trailing whitespace",
			src:  ".parent {\n  .child {\n    color: red;\n  }\n}\n",
		},
		{
			name: "nested with no trailing newline",
			src:  ".parent {\n  .child {\n    color: red;\n  }\n}",
		},
	}

	modes := []struct {
		name string
		opts FormatOptions
	}{
		{"expanded", FormatOptions{TabSize: 2, InsertSpaces: true}},
		{
			"compact",
			FormatOptions{
				TabSize:      2,
				InsertSpaces: true,
				Mode:         FormatCompact,
				PrintWidth:   80,
			},
		},
		{
			"preserve",
			FormatOptions{
				TabSize:      2,
				InsertSpaces: true,
				Mode:         FormatPreserve,
				PrintWidth:   80,
			},
		},
		{
			"detect",
			FormatOptions{
				TabSize:      2,
				InsertSpaces: true,
				Mode:         FormatDetect,
				PrintWidth:   80,
			},
		},
	}

	for _, tc := range cases {
		for _, m := range modes {
			t.Run(tc.name+"/"+m.name, func(t *testing.T) {
				src := []byte(tc.src)
				ss, _ := parser.Parse(src)
				result := Format(ss, src, m.opts)
				if strings.Contains(result, "} {}") {
					t.Errorf("stray '} {}' in output:\n%s", result)
				}
				if strings.HasSuffix(result, "} {\n}\n") {
					t.Errorf("stray expanded empty ruleset at end:\n%s", result)
				}
			})
		}
	}
}
