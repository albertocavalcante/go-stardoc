package stardoc

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParse_RulesGoCore(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "rules_go_core.md"))
	if err != nil {
		t.Fatal(err)
	}

	m, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// Verify reference links are parsed
	if len(m.RefLinks) == 0 {
		t.Error("expected reference links, got none")
	}
	// Check a known ref link
	found := false
	for _, rl := range m.RefLinks {
		if rl.Label == "GoInfo" {
			found = true
			if rl.URL != "/go/providers.rst#GoInfo" {
				t.Errorf("GoInfo ref link URL = %q, want /go/providers.rst#GoInfo", rl.URL)
			}
			break
		}
	}
	if !found {
		t.Error("expected GoInfo reference link")
	}

	// Verify rules are parsed
	if len(m.Rules) == 0 {
		t.Fatal("expected rules, got none")
	}

	// Check go_binary
	var goBinary *Rule
	for i := range m.Rules {
		if m.Rules[i].Name == "go_binary" {
			goBinary = &m.Rules[i]
			break
		}
	}
	if goBinary == nil {
		t.Fatal("expected go_binary rule")
	}

	// Verify go_binary has documentation
	if goBinary.Doc == "" {
		t.Error("go_binary: expected documentation")
	}
	if !strings.Contains(goBinary.Doc, "executable") {
		t.Errorf("go_binary doc doesn't mention 'executable': %q", goBinary.Doc)
	}

	// Verify signature
	if goBinary.Signature == "" {
		t.Error("go_binary: expected signature")
	}
	if !strings.Contains(goBinary.Signature, "go_binary(") {
		t.Errorf("go_binary signature doesn't contain 'go_binary(': %q", goBinary.Signature)
	}

	// Verify providers
	if len(goBinary.ProviderNames) == 0 {
		t.Error("go_binary: expected providers")
	}
	if !slices.Contains(goBinary.ProviderNames, "GoArchive") {
		t.Errorf("go_binary providers = %v, want GoArchive", goBinary.ProviderNames)
	}

	// Verify attributes
	if len(goBinary.Attributes) == 0 {
		t.Fatal("go_binary: expected attributes")
	}

	// Check 'name' attribute
	nameAttr := findAttr(goBinary.Attributes, "name")
	if nameAttr == nil {
		t.Fatal("go_binary: expected 'name' attribute")
	}
	if !nameAttr.Mandatory {
		t.Error("go_binary.name: expected mandatory")
	}
	if nameAttr.Type != "Name" {
		t.Errorf("go_binary.name: type = %q, want Name", nameAttr.Type)
	}

	// Check 'deps' attribute
	depsAttr := findAttr(goBinary.Attributes, "deps")
	if depsAttr == nil {
		t.Fatal("go_binary: expected 'deps' attribute")
	}
	if depsAttr.Mandatory {
		t.Error("go_binary.deps: expected optional")
	}
	if depsAttr.Type != "List of labels" {
		t.Errorf("go_binary.deps: type = %q, want List of labels", depsAttr.Type)
	}
	if depsAttr.Default != "[]" {
		t.Errorf("go_binary.deps: default = %q, want []", depsAttr.Default)
	}

	// Check 'cgo' attribute (Boolean type)
	cgoAttr := findAttr(goBinary.Attributes, "cgo")
	if cgoAttr == nil {
		t.Fatal("go_binary: expected 'cgo' attribute")
	}
	if cgoAttr.Type != "Boolean" {
		t.Errorf("go_binary.cgo: type = %q, want Boolean", cgoAttr.Type)
	}
	if cgoAttr.Default != "False" {
		t.Errorf("go_binary.cgo: default = %q, want False", cgoAttr.Default)
	}

	// Check 'out' attribute (String type)
	outAttr := findAttr(goBinary.Attributes, "out")
	if outAttr == nil {
		t.Fatal("go_binary: expected 'out' attribute")
	}
	if outAttr.Type != "String" {
		t.Errorf("go_binary.out: type = %q, want String", outAttr.Type)
	}

	// Check 'env' attribute (Dictionary type)
	envAttr := findAttr(goBinary.Attributes, "env")
	if envAttr == nil {
		t.Fatal("go_binary: expected 'env' attribute")
	}
	if envAttr.Type != "Dictionary: String -> String" {
		t.Errorf("go_binary.env: type = %q, want Dictionary: String -> String", envAttr.Type)
	}

	// Verify go_library exists too
	var goLib *Rule
	for i := range m.Rules {
		if m.Rules[i].Name == "go_library" {
			goLib = &m.Rules[i]
			break
		}
	}
	if goLib == nil {
		t.Fatal("expected go_library rule")
	}
	if len(goLib.Attributes) == 0 {
		t.Error("go_library: expected attributes")
	}

	// Verify go_cross_binary
	var goCross *Rule
	for i := range m.Rules {
		if m.Rules[i].Name == "go_cross_binary" {
			goCross = &m.Rules[i]
			break
		}
	}
	if goCross == nil {
		t.Fatal("expected go_cross_binary rule")
	}
	if !strings.Contains(goCross.Doc, "cross compile") {
		t.Errorf("go_cross_binary doc doesn't mention 'cross compile': %q", goCross.Doc)
	}

	t.Logf("Parsed %d rules, %d ref links", len(m.Rules), len(m.RefLinks))
	for _, r := range m.Rules {
		t.Logf("  Rule %s: %d attrs, %d providers", r.Name, len(r.Attributes), len(r.ProviderNames))
	}
}

func TestParse_Minimal(t *testing.T) {
	input := `<!-- Generated with Stardoc -->

<a id="my_rule"></a>

## my_rule

<pre>
my_rule(<a href="#my_rule-name">name</a>, <a href="#my_rule-src">src</a>)
</pre>

A simple rule.

**ATTRIBUTES**

| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="my_rule-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="my_rule-src"></a>src |  The source file.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  ` + "`None`" + `  |
`

	m, err := Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(m.Rules))
	}

	r := m.Rules[0]
	if r.Name != "my_rule" {
		t.Errorf("name = %q, want my_rule", r.Name)
	}
	if r.Signature != "my_rule(name, src)" {
		t.Errorf("signature = %q, want my_rule(name, src)", r.Signature)
	}
	if !strings.Contains(r.Doc, "simple rule") {
		t.Errorf("doc = %q, expected 'simple rule'", r.Doc)
	}

	if len(r.Attributes) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(r.Attributes))
	}

	nameA := r.Attributes[0]
	if nameA.Name != "name" || !nameA.Mandatory || nameA.Type != "Name" {
		t.Errorf("name attr = %+v", nameA)
	}

	srcA := r.Attributes[1]
	if srcA.Name != "src" || srcA.Mandatory || srcA.Type != "Label" || srcA.Default != "None" {
		t.Errorf("src attr = %+v", srcA)
	}
}

func TestParse_MultilineComment(t *testing.T) {
	input := `<!--
  This is a multi-line
  HTML comment
-->

<a id="my_rule"></a>

## my_rule

<pre>
my_rule(<a href="#my_rule-name">name</a>)
</pre>

A rule after a multi-line comment.

**ATTRIBUTES**

| Name | Description | Type | Mandatory | Default |
| :--- | :--- | :--- | :--- | :--- |
| <a id="my_rule-name"></a>name | Target name. | Name | required |  |
`

	m, err := Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(m.Rules))
	}
	if m.Rules[0].Name != "my_rule" {
		t.Errorf("name = %q, want my_rule", m.Rules[0].Name)
	}
	// The comment text should NOT appear in module doc
	if strings.Contains(m.Doc, "multi-line") {
		t.Errorf("comment leaked into doc: %q", m.Doc)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	m, err := Parse([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Rules) != 0 {
		t.Errorf("expected 0 rules from empty input, got %d", len(m.Rules))
	}
}

func TestParseFile(t *testing.T) {
	path := filepath.Join("testdata", "rules_go_core.md")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("testdata not available: %v", err)
	}

	m, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if m.Name != "rules_go_core" {
		t.Errorf("name = %q, want rules_go_core", m.Name)
	}
	if len(m.Rules) == 0 {
		t.Error("expected rules")
	}
}

func TestParseReader(t *testing.T) {
	input := `<a id="my_rule"></a>

## my_rule

<pre>
my_rule(<a href="#my_rule-name">name</a>)
</pre>

A rule.

**ATTRIBUTES**

| Name | Description | Type | Mandatory | Default |
| :--- | :--- | :--- | :--- | :--- |
| <a id="my_rule-name"></a>name | Target name. | Name | required |  |
`
	m, err := ParseReader(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(m.Rules))
	}
}

func TestParseAll_NonRecursive(t *testing.T) {
	dir := t.TempDir()
	// Create a .md file in the directory.
	os.WriteFile(filepath.Join(dir, "rules.md"), []byte(`<a id="r"></a>

## r

<pre>r()</pre>

Doc.
`), 0o644)

	// Create a subdirectory with another .md file.
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "nested.md"), []byte(`<a id="n"></a>

## n

<pre>n()</pre>
`), 0o644)

	modules, err := ParseAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Should only find the top-level file, not the nested one.
	if len(modules) != 1 {
		t.Errorf("expected 1 module (non-recursive), got %d", len(modules))
	}
}

func TestCleanDoc(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{
			"html_list",
			`List: <ul><li>item1</li><li>item2</li></ul>`,
			`List: - item1- item2`,
		},
		{
			"br_tags",
			`Line one.<br>Line two.<br/>Line three.<br />Line four.`,
			"Line one.\nLine two.\nLine three.\nLine four.",
		},
		{
			"anchor_removal",
			`<a id="rule-name"></a>The name attribute.`,
			`The name attribute.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanDoc(tt.input)
			if got != tt.want {
				t.Errorf("cleanDoc(%q) =\n%q\nwant:\n%q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParse_PreservePreambleParagraphs(t *testing.T) {
	input := `<!-- Generated with Stardoc -->

# Title

First paragraph.

Second paragraph.

<a id="r"></a>

## r

<pre>r()</pre>
`
	m, err := Parse([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(m.Doc, "First paragraph.\n\nSecond paragraph.") {
		t.Errorf("preamble lost paragraph break:\n%q", m.Doc)
	}
}

func TestSplitTableRow(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"| a | b | c | d | e |", 5},
		{`| <a id="x"></a>name | desc | <a href="url">Type</a> | required |  |`, 5},
	}
	for _, tt := range tests {
		cells := splitTableRow(tt.input)
		if len(cells) != tt.want {
			t.Errorf("splitTableRow(%q) = %d cells, want %d: %v", tt.input, len(cells), tt.want, cells)
		}
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{`<a href="url">text</a>`, "text"},
		{`plain text`, "plain text"},
		{`<a id="x"></a>name`, "name"},
	}
	for _, tt := range tests {
		got := stripHTML(tt.input)
		if got != tt.want {
			t.Errorf("stripHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// helpers

func findAttr(attrs []Attribute, name string) *Attribute {
	for i := range attrs {
		if attrs[i].Name == name {
			return &attrs[i]
		}
	}
	return nil
}
