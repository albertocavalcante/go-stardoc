// Package stardoc parses Stardoc-generated output into structured Go types.
//
// Stardoc (https://github.com/bazelbuild/stardoc) is Bazel's official
// documentation generator for Starlark rules. It outputs either Markdown
// files or binary protobuf (ModuleInfo).
//
// This package parses both formats into [Module], [Rule], [Provider],
// [Function], and [Aspect] types that can be used for documentation
// generation, API diffing, or other tooling.
//
// Prefer [ParseProto] / [ParseProtoFile] when protobuf output is available;
// use [Parse] / [ParseFile] for Markdown output.
package stardoc

// Module represents a single Stardoc-generated documentation file.
// Each .bzl file produces one Module.
type Module struct {
	// Name is the module name, derived from the filename or load path.
	Name string

	// File is the display label of the .bzl file (e.g., "@rules_go//go:def.bzl").
	// Only populated when parsing from protobuf.
	File string

	// Doc is the file-level documentation.
	Doc string

	// RefLinks are Markdown reference-style link definitions found at
	// the top of the file (e.g., [GoInfo]: /go/providers.rst#GoInfo).
	// Only populated when parsing from Markdown.
	RefLinks []RefLink

	Rules     []Rule
	Providers []Provider
	Functions []Function
	Aspects   []Aspect
}

// Rule represents a Starlark rule definition.
type Rule struct {
	// Name is the rule name (e.g., "go_binary").
	Name string

	// Doc is the rule's documentation text.
	Doc string

	// Signature is the formatted call signature
	// (e.g., "go_binary(name, deps, srcs, ...)").
	// Only populated when parsing from Markdown.
	Signature string

	// LoadStatement is the load() path shown in the signature block.
	// Only populated when parsing from Markdown.
	LoadStatement string

	// ProviderNames lists the providers this rule advertises.
	ProviderNames []string

	// Attributes are the rule's attributes.
	Attributes []Attribute

	// OriginKey uniquely identifies this rule across modules.
	// Only populated when parsing from protobuf.
	OriginKey *OriginKey

	// Test is true if this is a test rule.
	// Only populated when parsing from protobuf.
	Test bool

	// Executable is true if this is an executable rule.
	// Only populated when parsing from protobuf.
	Executable bool
}

// Provider represents a Starlark provider definition.
type Provider struct {
	Name      string
	Doc       string
	Fields    []Field
	OriginKey *OriginKey
}

// Function represents a Starlark function or macro.
type Function struct {
	Name       string
	Doc        string
	Signature  string
	Parameters []Parameter
	Returns    string
	Deprecated string
	OriginKey  *OriginKey
}

// Aspect represents a Starlark aspect definition.
type Aspect struct {
	Name string
	Doc  string

	// AspectAttributes are the rule attributes along which the aspect propagates.
	AspectAttributes []string

	Attributes []Attribute
	OriginKey  *OriginKey
}

// Attribute represents a rule or aspect attribute.
type Attribute struct {
	// Name is the attribute name (e.g., "deps").
	Name string

	// Doc is the attribute documentation.
	Doc string

	// Type is the attribute type as a human-readable string
	// (e.g., "Label", "List of labels", "String", "Boolean").
	Type string

	// TypeURL is the link target for the type, if present.
	// Only populated when parsing from Markdown.
	TypeURL string

	// AttrType is the strongly-typed attribute type enum.
	// Only populated when parsing from protobuf; zero (AttrUnknown) for Markdown.
	AttrType AttributeType

	// Mandatory indicates whether the attribute is required.
	Mandatory bool

	// Default is the default value as a string (e.g., "[]", `""`, "False").
	// Empty string means no default was specified (typically mandatory attrs).
	Default string

	// Nonconfigurable is true if the attribute cannot be configured via select().
	// Only populated when parsing from protobuf.
	Nonconfigurable bool
}

// Field represents a provider field.
type Field struct {
	Name string
	Doc  string
}

// Parameter represents a function parameter.
type Parameter struct {
	Name      string
	Doc       string
	Default   string
	Mandatory bool
	Role      ParamRole
}

// ParamRole describes the syntactic role of a function parameter.
type ParamRole int32

const (
	ParamRoleUnspecified  ParamRole = 0
	ParamRoleOrdinary    ParamRole = 1
	ParamRolePositional  ParamRole = 2
	ParamRoleKeywordOnly ParamRole = 3
	ParamRoleVarargs     ParamRole = 4
	ParamRoleKwargs      ParamRole = 5
)

// AttributeType classifies a rule attribute's type.
type AttributeType int32

const (
	AttrUnknown        AttributeType = 0
	AttrName           AttributeType = 1
	AttrInt            AttributeType = 2
	AttrLabel          AttributeType = 3
	AttrString         AttributeType = 4
	AttrStringList     AttributeType = 5
	AttrIntList        AttributeType = 6
	AttrLabelList      AttributeType = 7
	AttrBoolean        AttributeType = 8
	AttrLabelStringDict AttributeType = 9
	AttrStringDict     AttributeType = 10
	AttrStringListDict AttributeType = 11
	AttrOutput         AttributeType = 12
	AttrOutputList     AttributeType = 13
	AttrLabelDictUnary AttributeType = 14
	AttrLabelListDict  AttributeType = 15
)

// String returns the human-readable name of an [AttributeType].
func (t AttributeType) String() string {
	switch t {
	case AttrName:
		return "Name"
	case AttrInt:
		return "Integer"
	case AttrLabel:
		return "Label"
	case AttrString:
		return "String"
	case AttrStringList:
		return "List of strings"
	case AttrIntList:
		return "List of integers"
	case AttrLabelList:
		return "List of labels"
	case AttrBoolean:
		return "Boolean"
	case AttrLabelStringDict:
		return "Dictionary: Label -> String"
	case AttrStringDict:
		return "Dictionary: String -> String"
	case AttrStringListDict:
		return "Dictionary: String -> List of strings"
	case AttrOutput:
		return "Output"
	case AttrOutputList:
		return "List of outputs"
	case AttrLabelDictUnary:
		return "Dictionary: Label -> String"
	case AttrLabelListDict:
		return "Dictionary: String -> List of labels"
	default:
		return "unknown"
	}
}

// OriginKey uniquely identifies a rule, provider, aspect, or function
// across modules. Useful for building unambiguous cross-references.
type OriginKey struct {
	// Name is the name under which the entity was originally exported.
	Name string
	// File is the display label of the .bzl file where the entity was declared.
	File string
}

// RefLink is a Markdown reference-style link definition.
type RefLink struct {
	Label string // e.g., "GoInfo"
	URL   string // e.g., "/go/providers.rst#GoInfo"
}
