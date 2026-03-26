// Package stardoc parses Stardoc-generated Markdown into structured Go types.
//
// Stardoc (https://github.com/bazelbuild/stardoc) is Bazel's official
// documentation generator for Starlark rules. It outputs Markdown files
// with a specific format: HTML anchors, pre-formatted signatures, and
// GFM-style attribute tables.
//
// This package parses that output into [Module], [Rule], [Provider],
// [Function], and [Aspect] types that can be used for documentation
// generation, API diffing, or other tooling.
package stardoc

// Module represents a single Stardoc-generated documentation file.
// Each .bzl file produces one Module.
type Module struct {
	// Name is the module name, derived from the filename or load path.
	Name string

	// Doc is the file-level documentation (HTML comment preamble or
	// text before the first rule/function heading).
	Doc string

	// RefLinks are Markdown reference-style link definitions found at
	// the top of the file (e.g., [GoInfo]: /go/providers.rst#GoInfo).
	RefLinks []RefLink

	// Rules are the rule definitions documented in this module.
	Rules []Rule

	// Providers are the provider definitions.
	Providers []Provider

	// Functions are the function/macro definitions.
	Functions []Function

	// Aspects are the aspect definitions.
	Aspects []Aspect
}

// Rule represents a Starlark rule definition.
type Rule struct {
	// Name is the rule name (e.g., "go_binary").
	Name string

	// Doc is the rule's documentation text.
	Doc string

	// Signature is the formatted call signature
	// (e.g., "go_binary(name, deps, srcs, ...)").
	Signature string

	// LoadStatement is the load() path shown in the signature block
	// (e.g., `load("@rules_go//go:def.bzl", "go_binary")`).
	LoadStatement string

	// ProviderNames lists the providers this rule returns
	// (e.g., ["GoInfo", "GoArchive"]).
	ProviderNames []string

	// Attributes are the rule's attributes.
	Attributes []Attribute
}

// Provider represents a Starlark provider definition.
type Provider struct {
	Name   string
	Doc    string
	Fields []Field
}

// Function represents a Starlark function or macro.
type Function struct {
	Name       string
	Doc        string
	Signature  string
	Parameters []Parameter
	Returns    string
}

// Aspect represents a Starlark aspect definition.
type Aspect struct {
	Name       string
	Doc        string
	Attributes []Attribute
}

// Attribute represents a rule or aspect attribute.
type Attribute struct {
	// Name is the attribute name (e.g., "deps").
	Name string

	// Doc is the attribute documentation.
	Doc string

	// Type is the attribute type as displayed by Stardoc
	// (e.g., "Label", "List of labels", "String", "Boolean",
	// "Dictionary: String -> String").
	Type string

	// TypeURL is the link target for the type, if present
	// (e.g., "https://bazel.build/concepts/labels").
	TypeURL string

	// Mandatory indicates whether the attribute is required.
	Mandatory bool

	// Default is the default value as a string (e.g., "[]", `""`, "False").
	// Empty string means no default was specified (typically mandatory attrs).
	Default string
}

// Field represents a provider field.
type Field struct {
	Name string
	Doc  string
}

// Parameter represents a function parameter.
type Parameter struct {
	Name    string
	Doc     string
	Default string
}

// RefLink is a Markdown reference-style link definition.
type RefLink struct {
	Label string // e.g., "GoInfo"
	URL   string // e.g., "/go/providers.rst#GoInfo"
}
