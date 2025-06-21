// Copyright 2025 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grh

import (
	"strings"
	"testing"
)

func TestHugoProcessor_FindShortcodes(t *testing.T) {
	processor := NewHugoProcessor()

	tests := []struct {
		name     string
		input    string
		expected []HugoShortcode
	}{
		{
			name:  "paired shortcode with angle brackets",
			input: "{{< highlight javascript >}}console.log('test');{{< /highlight >}}",
			expected: []HugoShortcode{
				{
					Type:    "paired",
					Name:    "highlight",
					Content: "console.log('test');",
				},
			},
		},
		{
			name:  "paired shortcode with percent signs",
			input: "{{% note %}}This is a note{{% /note %}}",
			expected: []HugoShortcode{
				{
					Type:    "paired",
					Name:    "note",
					Content: "This is a note",
				},
			},
		},
		{
			name:  "self-closing shortcode",
			input: "{{< figure src=\"image.jpg\" alt=\"Test\" >}}",
			expected: []HugoShortcode{
				{
					Type:    "self-closing",
					Name:    "figure",
					Content: "",
				},
			},
		},
		{
			name:  "multiple shortcodes",
			input: "{{< highlight go >}}func main() {}{{< /highlight >}} and {{< figure src=\"test.jpg\" >}}",
			expected: []HugoShortcode{
				{
					Type:    "paired",
					Name:    "highlight",
					Content: "func main() {}",
				},
				{
					Type:    "self-closing",
					Name:    "figure",
					Content: "",
				},
			},
		},
		{
			name:     "no shortcodes",
			input:    "This is regular markdown text with no shortcodes.",
			expected: []HugoShortcode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortcodes := processor.FindShortcodes(tt.input)

			if len(shortcodes) != len(tt.expected) {
				t.Errorf("Expected %d shortcodes, got %d", len(tt.expected), len(shortcodes))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(shortcodes) {
					t.Errorf("Missing shortcode at index %d", i)
					continue
				}

				sc := shortcodes[i]
				if sc.Type != expected.Type {
					t.Errorf("Shortcode %d: expected type %q, got %q", i, expected.Type, sc.Type)
				}
				if sc.Name != expected.Name {
					t.Errorf("Shortcode %d: expected name %q, got %q", i, expected.Name, sc.Name)
				}
				if sc.Content != expected.Content {
					t.Errorf("Shortcode %d: expected content %q, got %q", i, expected.Content, sc.Content)
				}
			}
		})
	}
}

func TestHugoProcessor_PreserveAndRestoreShortcodes(t *testing.T) {
	processor := NewHugoProcessor()

	input := `# Test Document

This is a test with {{< highlight javascript >}}
console.log('test');
{{< /highlight >}} and {{% note %}}
This is a note
{{% /note %}} shortcodes.

Also {{< figure src="test.jpg" >}} here.`

	// ショートコードを保護
	preserved, placeholders := processor.PreserveShortcodes(input)

	// プレースホルダーが作成されているかチェック
	if len(placeholders) == 0 {
		t.Error("Expected placeholders to be created")
	}

	// 元のショートコードがプレースホルダーに置換されているかチェック
	if strings.Contains(preserved, "{{<") || strings.Contains(preserved, "{{% ") {
		t.Error("Original shortcodes should be replaced with placeholders")
	}

	// プレースホルダーが含まれているかチェック
	hasPlaceholder := false
	for placeholder := range placeholders {
		if strings.Contains(preserved, placeholder) {
			hasPlaceholder = true
			break
		}
	}
	if !hasPlaceholder {
		t.Error("Preserved text should contain placeholders")
	}

	// ショートコードを復元
	restored := processor.RestoreShortcodes(preserved, placeholders)

	// 復元されたテキストが元のテキストと一致するかチェック
	if restored != input {
		t.Errorf("Restored text does not match original.\nOriginal:\n%s\nRestored:\n%s", input, restored)
	}
}

func TestHugoProcessor_ValidateHugoMarkdown(t *testing.T) {
	processor := NewHugoProcessor()

	tests := []struct {
		name          string
		input         string
		expectIssues  bool
		expectedCount int
	}{
		{
			name: "valid markdown with shortcodes",
			input: `# Title

{{< highlight go >}}
func main() {
    fmt.Println("Hello")
}
{{< /highlight >}}

{{% note %}}
This is a note
{{% /note %}}`,
			expectIssues:  false,
			expectedCount: 0,
		},
		{
			name: "empty paired shortcode",
			input: `{{< highlight >}}{{< /highlight >}}`,
			expectIssues:  true,
			expectedCount: 1,
		},
		{
			name: "unclosed code block",
			input: `# Title

` + "```go" + `
func main() {}
// missing closing fence`,
			expectIssues:  true,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := processor.ValidateHugoMarkdown(tt.input)

			if tt.expectIssues && len(issues) == 0 {
				t.Error("Expected validation issues, but got none")
			}

			if !tt.expectIssues && len(issues) > 0 {
				t.Errorf("Expected no validation issues, but got %d: %v", len(issues), issues)
			}

			if tt.expectedCount > 0 && len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d: %v", tt.expectedCount, len(issues), issues)
			}
		})
	}
}

func TestHugoProcessor_isValidShortcodeName(t *testing.T) {
	processor := NewHugoProcessor()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid name", "highlight", true},
		{"valid with dash", "code-block", true},
		{"valid with underscore", "my_shortcode", true},
		{"valid with numbers", "shortcode123", true},
		{"empty name", "", false},
		{"invalid with @", "invalid@name", false},
		{"invalid with space", "invalid name", false},
		{"invalid with dot", "invalid.name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.isValidShortcodeName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidShortcodeName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
