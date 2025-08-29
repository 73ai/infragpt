package slack

import (
	"regexp"
	"strings"
	"testing"
)

func TestTransformMarkdownToSlack(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fast path - no markdown",
			input:    "Plain text with no markdown",
			expected: "Plain text with no markdown",
		},
		{
			name:     "headers conversion",
			input:    "## Issue Analysis\nSome content",
			expected: "*Issue Analysis*\nSome content",
		},
		{
			name:     "multiple header levels",
			input:    "# Main\n## Sub\n### Details",
			expected: "*Main*\n*Sub*\n*Details*",
		},
		{
			name:     "numbered list with bold titles",
			input:    "1. **Runaway processes**: There might be issues\n2. **Traffic spike**: High load detected",
			expected: "1. *Runaway processes*: There might be issues\n2. *Traffic spike*: High load detected",
		},
		{
			name:     "numbered list with bold titles no description",
			input:    "1. **Process Check**\n2. **Memory Analysis**",
			expected: "1. *Process Check*\n2. *Memory Analysis*",
		},
		{
			name:     "generic numbered lists",
			input:    "1. Check system logs\n2. Restart service\n3. Monitor metrics",
			expected: "1. Check system logs\n2. Restart service\n3. Monitor metrics",
		},
		{
			name:     "inline bold text",
			input:    "This is **bold** text and **more bold** here",
			expected: "This is *bold* text and *more bold* here",
		},
		{
			name:     "inline code with bold inside - should not transform",
			input:    "Use `**bold**` in code and **actual** bold",
			expected: "Use `**bold**` in code and *actual* bold",
		},
		{
			name:     "code block preservation",
			input:    "```\n**this should not change**\necho \"hello\"\n```",
			expected: "```\n**this should not change**\necho \"hello\"\n```",
		},
		{
			name:     "code block with indentation",
			input:    "    ```bash\n    echo **test**\n    ```",
			expected: "```bash\n    echo **test**\n    ```", // Note: leading indentation before ``` is not preserved
		},
		{
			name:     "mixed content with empty lines",
			input:    "## Header\n\n**Bold** text\n\n```\ncode\n```",
			expected: "*Header*\n\n*Bold* text\n\n```\ncode\n```",
		},
		{
			name:     "nested inline code and bold",
			input:    "Run `docker ps` and check **status** column",
			expected: "Run `docker ps` and check *status* column",
		},
		{
			name: "complex slack message example",
			input: `## Issue Analysis
High CPU usage has been detected on host.name

## Potential Root Causes
1. **Runaway processes**: There might be processes consuming CPU
2. **Traffic spike**: Sudden increase in traffic

## Recommended Next Steps  
1. **Identify Processes**: Check real-time CPU usage
2. Use ` + "`top`" + ` command to monitor`,
			expected: `*Issue Analysis*
High CPU usage has been detected on host.name

*Potential Root Causes*
1. *Runaway processes*: There might be processes consuming CPU
2. *Traffic spike*: Sudden increase in traffic

*Recommended Next Steps*
1. *Identify Processes*: Check real-time CPU usage
2. Use ` + "`top`" + ` command to monitor`,
		},
		{
			name:     "empty lines with whitespace",
			input:    "Line 1\n   \nLine 3",
			expected: "Line 1\n   \nLine 3", // Note: whitespace-only lines preserved intentionally
		},
		{
			name:     "code fence toggle state",
			input:    "Before\n```\n**inside**\n```\n**after**",
			expected: "Before\n```\n**inside**\n```\n*after*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformMarkdownToSlack(tt.input)
			if result != tt.expected {
				t.Errorf("transformMarkdownToSlack() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTransformBoldOutsideCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold outside code",
			input:    "**bold** text",
			expected: "*bold* text",
		},
		{
			name:     "bold inside code - no change",
			input:    "`**bold**` text",
			expected: "`**bold**` text",
		},
		{
			name:     "mixed bold and code",
			input:    "**bold** and `**code**` and **more**",
			expected: "*bold* and `**code**` and *more*",
		},
		{
			name:     "multiple code spans",
			input:    "`**a**` normal **bold** `**b**` more **bold**",
			expected: "`**a**` normal *bold* `**b**` more *bold*",
		},
	}

	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformBoldOutsideCode(tt.input, boldRegex)
			if result != tt.expected {
				t.Errorf("transformBoldOutsideCode() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestTransformMarkdownToSlack_Performance tests the fast-path optimization
func TestTransformMarkdownToSlack_Performance(t *testing.T) {
	plainText := strings.Repeat("Plain text without any markdown symbols", 100)
	result := transformMarkdownToSlack(plainText)
	expected := strings.TrimSpace(plainText)

	if result != expected {
		t.Errorf("Fast-path failed: got %q, want %q", result, expected)
	}
}
