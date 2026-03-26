package stardoc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Parse parses a single Stardoc-generated Markdown file into a [Module].
//
// The input should be the raw Markdown content produced by Stardoc.
// The module Name is left empty; callers should set it based on the filename
// or load path.
func Parse(markdown []byte) (*Module, error) {
	lines := splitLines(string(markdown))
	return parseLines(lines)
}

// ParseFile parses a Stardoc-generated Markdown file from disk.
//
// The module Name is set to the base filename without extension.
func ParseFile(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %w", err)
	}
	m, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %s: %w", path, err)
	}
	m.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return m, nil
}

// ParseAll parses all .md files in a directory as Stardoc output.
// Only reads the immediate directory (non-recursive).
//
// Returns one [Module] per file, sorted by filename.
func ParseAll(dir string) ([]*Module, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("stardoc: ParseAll %s: %w", dir, err)
	}

	modules := make([]*Module, 0)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		m, err := ParseFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})
	return modules, nil
}

// ParseReader parses stardoc Markdown from an [io.Reader].
func ParseReader(r io.Reader) (*Module, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %w", err)
	}
	return Parse(data)
}

// Regex patterns for stardoc format.
var (
	// <a id="go_binary"></a>
	anchorRe = regexp.MustCompile(`^<a id="([^"]+)"></a>\s*$`)

	// ## rule_name
	headingRe = regexp.MustCompile(`^##\s+(.+)$`)

	// [label]: url
	refLinkRe = regexp.MustCompile(`^\s*\[([^\]]+)\]:\s+(.+)$`)

	// | Name | Description | Type | Mandatory | Default |
	tableHeaderRe = regexp.MustCompile(`^\|\s*Name\s*\|`)

	// | :--- | :--- | ... (separator row)
	tableSepRe = regexp.MustCompile(`^\|[\s:-]+\|`)

	// <a id="rule-attr"></a>text  inside table cell
	cellAnchorRe = regexp.MustCompile(`<a id="[^"]*"></a>`)

	// <a href="url">text</a>
	htmlLinkRe = regexp.MustCompile(`<a href="([^"]*)">(.*?)</a>`)

	// **Providers:** or **Providers**
	providersRe = regexp.MustCompile(`^\*\*Providers:?\*\*`)

	// - [ProviderName]
	providerItemRe = regexp.MustCompile(`^-\s+\[([^\]]+)\]`)
)

// docCleaner converts stardoc HTML fragments to Markdown in a single pass.
var docCleaner = strings.NewReplacer(
	"<br>", "\n",
	"<br/>", "\n",
	"<br />", "\n",
	"<ul>", "",
	"</ul>", "",
	"<li>", "- ",
	"</li>", "",
)

func splitLines(s string) []string {
	if strings.ContainsRune(s, '\r') {
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
	}
	return strings.Split(s, "\n")
}

// section represents a range of lines belonging to one stardoc entry.
type section struct {
	name  string // from the anchor id
	start int    // first line index (the anchor line)
	end   int    // one past the last line
}

func parseLines(lines []string) (*Module, error) {
	m := &Module{}

	// Phase 1: find all section boundaries (anchored by <a id="..."></a>).
	var sections []section
	for i, line := range lines {
		if match := anchorRe.FindStringSubmatch(line); match != nil {
			sections = append(sections, section{name: match[1], start: i})
		}
	}
	// Set end boundaries.
	for i := range sections {
		if i+1 < len(sections) {
			sections[i].end = sections[i+1].start
		} else {
			sections[i].end = len(lines)
		}
	}

	// Phase 2: parse preamble (everything before first section).
	preambleEnd := len(lines)
	if len(sections) > 0 {
		preambleEnd = sections[0].start
	}
	parsePreamble(m, lines[:preambleEnd])

	// Phase 3: parse each section, passing name from Phase 1.
	for _, sec := range sections {
		rule := parseRuleSection(sec.name, lines[sec.start+1:sec.end])
		m.Rules = append(m.Rules, rule)
	}

	return m, nil
}

func parsePreamble(m *Module, lines []string) {
	var docLines []string
	inComment := false
	lastNonEmpty := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle multi-line HTML comments.
		if inComment {
			if strings.Contains(trimmed, "-->") {
				inComment = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "<!--") {
			if !strings.Contains(trimmed, "-->") {
				inComment = true
			}
			continue
		}

		// Reference links
		if match := refLinkRe.FindStringSubmatch(line); match != nil {
			m.RefLinks = append(m.RefLinks, RefLink{
				Label: match[1],
				URL:   strings.TrimSpace(match[2]),
			})
			continue
		}

		// Preserve blank lines between non-blank lines for paragraph structure.
		if trimmed == "" {
			if lastNonEmpty >= 0 {
				docLines = append(docLines, "")
			}
		} else {
			docLines = append(docLines, line)
			lastNonEmpty = len(docLines) - 1
		}
	}

	// Trim trailing blank lines.
	for len(docLines) > 0 && docLines[len(docLines)-1] == "" {
		docLines = docLines[:len(docLines)-1]
	}
	m.Doc = strings.Join(docLines, "\n")
}

// parseRuleSection parses a single stardoc section into a Rule.
// anchorName is the id from the <a id="..."></a> anchor.
// lines should start AFTER the anchor line.
func parseRuleSection(anchorName string, lines []string) Rule {
	r := Rule{Name: anchorName}

	var (
		docLines     []string
		inPre        bool
		preLines     []string
		inTable      bool
		seenTableHdr bool
		seenAttrsHdr bool
	)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// The ## heading overrides the anchor name with the display name.
		// Single FindStringSubmatch avoids double regex evaluation.
		if hm := headingRe.FindStringSubmatch(line); hm != nil {
			r.Name = strings.TrimSpace(hm[1])
			continue
		}

		// Handle <pre> blocks (signature).
		if strings.Contains(trimmed, "<pre>") && !inPre {
			inPre = true
			after := trimmed
			if idx := strings.Index(after, "<pre>"); idx >= 0 {
				after = after[idx+5:]
			}
			if strings.Contains(after, "</pre>") {
				after = strings.Replace(after, "</pre>", "", 1)
				preLines = append(preLines, after)
				inPre = false
				r.Signature, r.LoadStatement = parsePreBlock(preLines)
				preLines = nil
			} else if after != "" {
				preLines = append(preLines, after)
			}
			continue
		}
		if inPre {
			if strings.Contains(trimmed, "</pre>") {
				before := strings.Replace(trimmed, "</pre>", "", 1)
				if before != "" {
					preLines = append(preLines, before)
				}
				inPre = false
				r.Signature, r.LoadStatement = parsePreBlock(preLines)
				preLines = nil
			} else {
				preLines = append(preLines, line)
			}
			continue
		}

		// Handle providers.
		if providersRe.MatchString(trimmed) {
			continue
		}
		if m := providerItemRe.FindStringSubmatch(trimmed); m != nil {
			r.ProviderNames = append(r.ProviderNames, m[1])
			continue
		}

		// Handle ATTRIBUTES header.
		if trimmed == "**ATTRIBUTES**" || trimmed == "**Attributes:**" {
			seenAttrsHdr = true
			continue
		}

		// Handle attribute tables.
		if tableHeaderRe.MatchString(trimmed) {
			inTable = true
			seenTableHdr = true
			continue
		}
		if inTable && tableSepRe.MatchString(trimmed) {
			continue
		}
		if inTable && strings.HasPrefix(trimmed, "|") {
			attr := parseTableRow(trimmed)
			if attr.Name != "" {
				r.Attributes = append(r.Attributes, attr)
			}
			continue
		}
		if inTable && !strings.HasPrefix(trimmed, "|") {
			inTable = false
		}

		// Collect documentation (text between signature and attributes).
		if !seenTableHdr && !seenAttrsHdr {
			docLines = append(docLines, line)
		}
	}

	r.Doc = strings.TrimSpace(strings.Join(docLines, "\n"))
	return r
}

// parsePreBlock extracts the call signature and load statement from
// the collected <pre> block lines. Accepts lines directly to avoid
// a redundant join+split cycle.
func parsePreBlock(lines []string) (signature, loadStmt string) {
	var sigLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "load(") {
			loadStmt = stripHTML(trimmed)
			continue
		}
		sigLines = append(sigLines, trimmed)
	}

	signature = stripHTML(strings.Join(sigLines, " "))
	signature = strings.Join(strings.Fields(signature), " ")
	return signature, loadStmt
}

// parseTableRow parses a stardoc attribute table row.
//
// Format: | <a id="..."></a>name | description | type | mandatory | default |
func parseTableRow(line string) Attribute {
	cells := splitTableRow(line)
	if len(cells) < 5 {
		return Attribute{}
	}

	name := stripHTML(cells[0])
	name = strings.TrimSpace(name)

	doc := cleanDoc(cells[1])
	typ, typeURL := parseTypeCell(cells[2])

	mandatory := strings.TrimSpace(cells[3])
	isMandatory := mandatory == "required"

	def := strings.TrimSpace(stripHTML(cells[4]))
	def = strings.Trim(def, "`")
	def = strings.TrimSpace(def)

	return Attribute{
		Name:      name,
		Doc:       doc,
		Type:      typ,
		TypeURL:   typeURL,
		Mandatory: isMandatory,
		Default:   def,
	}
}

// splitTableRow splits a Markdown table row into cells.
// Handles HTML tags spanning cell boundaries (e.g., <a href="...">).
//
// Stardoc output is ASCII-only, so byte iteration is safe and avoids
// the allocation of converting to []rune. The function correctly handles
// multi-byte UTF-8 because '<', '>', '|' are all single-byte ASCII.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|")

	var cells []string
	var current strings.Builder
	inHTML := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '<':
			inHTML = true
			current.WriteByte(ch)
		case ch == '>':
			inHTML = false
			current.WriteByte(ch)
		case ch == '|' && !inHTML:
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
		case ch == '\\' && i+1 < len(line) && line[i+1] == '|':
			current.WriteByte('|')
			i++ // skip escaped pipe
		default:
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		cells = append(cells, strings.TrimSpace(current.String()))
	}

	return cells
}

// parseTypeCell extracts the type name and optional URL from a type cell.
func parseTypeCell(cell string) (typeName, typeURL string) {
	cell = strings.TrimSpace(cell)
	if m := htmlLinkRe.FindStringSubmatch(cell); m != nil {
		return m[2], m[1]
	}
	return stripHTML(cell), ""
}

// cleanDoc cleans up attribute documentation text from a table cell.
// Converts HTML fragments to Markdown in a single pass.
func cleanDoc(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = cellAnchorRe.ReplaceAllString(raw, "")
	raw = docCleaner.Replace(raw)
	return strings.TrimSpace(raw)
}

// stripHTML removes all HTML tags from a string.
func stripHTML(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
