package stardoc

import (
	"strings"

	pb "github.com/albertocavalcante/go-stardoc/gen"
)

// moduleFromProto converts a protobuf ModuleInfo to a Module.
func moduleFromProto(info *pb.ModuleInfo) *Module {
	m := &Module{
		File: info.GetFile(),
		Doc:  info.GetModuleDocstring(),
	}

	for _, ri := range info.GetRuleInfo() {
		m.Rules = append(m.Rules, ruleFromProto(ri))
	}
	for _, pi := range info.GetProviderInfo() {
		m.Providers = append(m.Providers, providerFromProto(pi))
	}
	for _, fi := range info.GetFuncInfo() {
		m.Functions = append(m.Functions, functionFromProto(fi))
	}
	for _, ai := range info.GetAspectInfo() {
		m.Aspects = append(m.Aspects, aspectFromProto(ai))
	}
	// Macros are treated as rules (they have name, doc, attributes).
	for _, mi := range info.GetMacroInfo() {
		m.Rules = append(m.Rules, macroFromProto(mi))
	}
	// Repository rules are treated as rules.
	for _, rri := range info.GetRepositoryRuleInfo() {
		m.Rules = append(m.Rules, repoRuleFromProto(rri))
	}

	// Derive Name from File label if available.
	if m.File != "" {
		m.Name = nameFromFileLabel(m.File)
	}

	return m
}

func ruleFromProto(ri *pb.RuleInfo) Rule {
	r := Rule{
		Name:       ri.GetRuleName(),
		Doc:        ri.GetDocString(),
		Test:       ri.GetTest(),
		Executable: ri.GetExecutable(),
		OriginKey:  originKeyFromProto(ri.GetOriginKey()),
	}

	// Extract advertised provider names.
	if ap := ri.GetAdvertisedProviders(); ap != nil {
		r.ProviderNames = ap.GetProviderName()
	}

	for _, ai := range ri.GetAttribute() {
		r.Attributes = append(r.Attributes, attrFromProto(ai))
	}

	// Build a synthetic signature from rule name + attribute names.
	r.Signature = buildSignature(r.Name, r.Attributes)

	return r
}

func providerFromProto(pi *pb.ProviderInfo) Provider {
	p := Provider{
		Name:      pi.GetProviderName(),
		Doc:       pi.GetDocString(),
		OriginKey: originKeyFromProto(pi.GetOriginKey()),
	}
	for _, fi := range pi.GetFieldInfo() {
		p.Fields = append(p.Fields, Field{
			Name: fi.GetName(),
			Doc:  fi.GetDocString(),
		})
	}
	return p
}

func functionFromProto(fi *pb.StarlarkFunctionInfo) Function {
	f := Function{
		Name:      fi.GetFunctionName(),
		Doc:       fi.GetDocString(),
		OriginKey: originKeyFromProto(fi.GetOriginKey()),
	}
	if ret := fi.GetReturn(); ret != nil {
		f.Returns = ret.GetDocString()
	}
	if dep := fi.GetDeprecated(); dep != nil {
		f.Deprecated = dep.GetDocString()
	}

	var paramNames []string
	for _, pi := range fi.GetParameter() {
		p := Parameter{
			Name:      pi.GetName(),
			Doc:       pi.GetDocString(),
			Default:   pi.GetDefaultValue(),
			Mandatory: pi.GetMandatory(),
			Role:      ParamRole(pi.GetRole()),
		}
		f.Parameters = append(f.Parameters, p)

		switch p.Role {
		case ParamRoleVarargs:
			paramNames = append(paramNames, "*"+p.Name)
		case ParamRoleKwargs:
			paramNames = append(paramNames, "**"+p.Name)
		default:
			paramNames = append(paramNames, p.Name)
		}
	}

	f.Signature = f.Name + "(" + strings.Join(paramNames, ", ") + ")"
	return f
}

func macroFromProto(mi *pb.MacroInfo) Rule {
	r := Rule{
		Name:      mi.GetMacroName(),
		Doc:       mi.GetDocString(),
		OriginKey: originKeyFromProto(mi.GetOriginKey()),
	}
	for _, ai := range mi.GetAttribute() {
		r.Attributes = append(r.Attributes, attrFromProto(ai))
	}
	r.Signature = buildSignature(r.Name, r.Attributes)
	return r
}

func repoRuleFromProto(rri *pb.RepositoryRuleInfo) Rule {
	r := Rule{
		Name:      rri.GetRuleName(),
		Doc:       rri.GetDocString(),
		OriginKey: originKeyFromProto(rri.GetOriginKey()),
	}
	for _, ai := range rri.GetAttribute() {
		r.Attributes = append(r.Attributes, attrFromProto(ai))
	}
	r.Signature = buildSignature(r.Name, r.Attributes)
	return r
}

func aspectFromProto(ai *pb.AspectInfo) Aspect {
	a := Aspect{
		Name:             ai.GetAspectName(),
		Doc:              ai.GetDocString(),
		AspectAttributes: ai.GetAspectAttribute(),
		OriginKey:        originKeyFromProto(ai.GetOriginKey()),
	}
	for _, attr := range ai.GetAttribute() {
		a.Attributes = append(a.Attributes, attrFromProto(attr))
	}
	return a
}

func attrFromProto(ai *pb.AttributeInfo) Attribute {
	attrType := AttributeType(ai.GetType())
	return Attribute{
		Name:            ai.GetName(),
		Doc:             ai.GetDocString(),
		Type:            attrType.String(),
		AttrType:        attrType,
		Mandatory:       ai.GetMandatory(),
		Default:         ai.GetDefaultValue(),
		Nonconfigurable: ai.GetNonconfigurable(),
	}
}

func originKeyFromProto(ok *pb.OriginKey) *OriginKey {
	if ok == nil || (ok.GetName() == "" && ok.GetFile() == "") {
		return nil
	}
	return &OriginKey{
		Name: ok.GetName(),
		File: ok.GetFile(),
	}
}

// buildSignature creates a synthetic call signature from a name and attributes.
// e.g., "go_binary(name, deps, srcs, ...)" for rules with many attributes.
func buildSignature(name string, attrs []Attribute) string {
	if len(attrs) == 0 {
		return name + "()"
	}

	var names []string
	for _, a := range attrs {
		names = append(names, a.Name)
	}

	sig := name + "(" + strings.Join(names, ", ") + ")"

	// If the signature is very long, truncate to first few attrs.
	if len(sig) > 120 && len(attrs) > 5 {
		short := make([]string, 5)
		for i := range 5 {
			short[i] = attrs[i].Name
		}
		sig = name + "(" + strings.Join(short, ", ") + ", ...)"
	}

	return sig
}

// nameFromFileLabel extracts a short name from a Bazel file label.
// e.g., "@rules_go//go:def.bzl" -> "def"
// e.g., "//foo/bar:baz.bzl" -> "baz"
func nameFromFileLabel(label string) string {
	// Take everything after the last ':'
	if idx := strings.LastIndexByte(label, ':'); idx >= 0 {
		label = label[idx+1:]
	} else if idx := strings.LastIndexByte(label, '/'); idx >= 0 {
		label = label[idx+1:]
	}
	// Strip .bzl extension
	label = strings.TrimSuffix(label, ".bzl")
	return label
}
