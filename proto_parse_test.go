package stardoc

import (
	"os"
	"path/filepath"
	"testing"

	pb "github.com/albertocavalcante/go-stardoc/gen"
	"google.golang.org/protobuf/proto"
)

// testModuleInfo builds a realistic ModuleInfo protobuf for testing.
func testModuleInfo() *pb.ModuleInfo {
	return &pb.ModuleInfo{
		ModuleDocstring: "Core Go rules for Bazel.",
		File:            "@rules_go//go:def.bzl",
		RuleInfo: []*pb.RuleInfo{
			{
				RuleName:  "go_binary",
				DocString: "Builds an executable from Go source files.",
				Attribute: []*pb.AttributeInfo{
					{
						Name:      "name",
						DocString: "A unique name for this target.",
						Type:      pb.AttributeType_NAME,
						Mandatory: true,
					},
					{
						Name:         "deps",
						DocString:    "List of Go libraries this target depends on.",
						Type:         pb.AttributeType_LABEL_LIST,
						DefaultValue: "[]",
					},
					{
						Name:         "srcs",
						DocString:    "Go source files.",
						Type:         pb.AttributeType_LABEL_LIST,
						DefaultValue: "[]",
					},
					{
						Name:         "cgo",
						DocString:    "Enable cgo support.",
						Type:         pb.AttributeType_BOOLEAN,
						DefaultValue: "False",
					},
					{
						Name:            "out",
						DocString:       "Output filename.",
						Type:            pb.AttributeType_STRING,
						DefaultValue:    "",
						Nonconfigurable: true,
					},
				},
				AdvertisedProviders: &pb.ProviderNameGroup{
					ProviderName: []string{"GoInfo", "GoArchive"},
				},
				OriginKey: &pb.OriginKey{
					Name: "go_binary",
					File: "@rules_go//go:def.bzl",
				},
				Test:       false,
				Executable: true,
			},
			{
				RuleName:  "go_test",
				DocString: "Builds and runs Go tests.",
				Attribute: []*pb.AttributeInfo{
					{Name: "name", Type: pb.AttributeType_NAME, Mandatory: true},
					{Name: "srcs", Type: pb.AttributeType_LABEL_LIST, DefaultValue: "[]"},
				},
				Test:       true,
				Executable: true,
			},
		},
		ProviderInfo: []*pb.ProviderInfo{
			{
				ProviderName: "GoInfo",
				DocString:    "Provider for Go compilation information.",
				FieldInfo: []*pb.ProviderFieldInfo{
					{Name: "importpath", DocString: "The import path."},
					{Name: "library", DocString: "The compiled library."},
				},
				OriginKey: &pb.OriginKey{
					Name: "GoInfo",
					File: "@rules_go//go/private:providers.bzl",
				},
			},
		},
		FuncInfo: []*pb.StarlarkFunctionInfo{
			{
				FunctionName: "go_context",
				DocString:    "Returns the Go build context.",
				Parameter: []*pb.FunctionParamInfo{
					{Name: "ctx", DocString: "The rule context.", Mandatory: true, Role: pb.FunctionParamRole_PARAM_ROLE_ORDINARY},
					{Name: "attrs", DocString: "Override attributes.", DefaultValue: "None", Role: pb.FunctionParamRole_PARAM_ROLE_KEYWORD_ONLY},
				},
				Return: &pb.FunctionReturnInfo{DocString: "A GoContext."},
				OriginKey: &pb.OriginKey{
					Name: "go_context",
					File: "@rules_go//go/private:context.bzl",
				},
			},
		},
		AspectInfo: []*pb.AspectInfo{
			{
				AspectName:      "go_archive_aspect",
				DocString:       "Collects Go archive information.",
				AspectAttribute: []string{"deps", "embed"},
				Attribute: []*pb.AttributeInfo{
					{Name: "_go_toolchain", Type: pb.AttributeType_LABEL},
				},
				OriginKey: &pb.OriginKey{
					Name: "go_archive_aspect",
					File: "@rules_go//go/private:rules.bzl",
				},
			},
		},
	}
}

func TestParseProto(t *testing.T) {
	info := testModuleInfo()
	data, err := proto.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	m, err := ParseProto(data)
	if err != nil {
		t.Fatal(err)
	}

	// Module-level fields
	if m.Doc != "Core Go rules for Bazel." {
		t.Errorf("Doc = %q", m.Doc)
	}
	if m.File != "@rules_go//go:def.bzl" {
		t.Errorf("File = %q", m.File)
	}
	if m.Name != "def" {
		t.Errorf("Name = %q, want def (derived from file label)", m.Name)
	}

	// Rules
	if len(m.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(m.Rules))
	}

	goBinary := m.Rules[0]
	if goBinary.Name != "go_binary" {
		t.Errorf("rule[0].Name = %q", goBinary.Name)
	}
	if !goBinary.Executable {
		t.Error("go_binary should be executable")
	}
	if goBinary.Test {
		t.Error("go_binary should not be a test")
	}
	if goBinary.OriginKey == nil || goBinary.OriginKey.Name != "go_binary" {
		t.Errorf("go_binary OriginKey = %+v", goBinary.OriginKey)
	}

	// Advertised providers
	if len(goBinary.ProviderNames) != 2 {
		t.Errorf("go_binary providers = %v", goBinary.ProviderNames)
	}

	// Attributes
	if len(goBinary.Attributes) != 5 {
		t.Fatalf("go_binary: expected 5 attrs, got %d", len(goBinary.Attributes))
	}

	nameAttr := goBinary.Attributes[0]
	if nameAttr.Name != "name" || !nameAttr.Mandatory || nameAttr.AttrType != AttrName {
		t.Errorf("name attr = %+v", nameAttr)
	}
	if nameAttr.Type != "Name" {
		t.Errorf("name attr type string = %q, want Name", nameAttr.Type)
	}

	depsAttr := goBinary.Attributes[1]
	if depsAttr.AttrType != AttrLabelList || depsAttr.Default != "[]" {
		t.Errorf("deps attr = %+v", depsAttr)
	}
	if depsAttr.Type != "List of labels" {
		t.Errorf("deps attr type string = %q, want List of labels", depsAttr.Type)
	}

	outAttr := goBinary.Attributes[4]
	if !outAttr.Nonconfigurable {
		t.Error("out attr should be nonconfigurable")
	}

	// Synthetic signature
	if goBinary.Signature == "" {
		t.Error("go_binary: expected synthetic signature")
	}
	if goBinary.Signature != "go_binary(name, deps, srcs, cgo, out)" {
		t.Errorf("go_binary signature = %q", goBinary.Signature)
	}

	goTest := m.Rules[1]
	if !goTest.Test || !goTest.Executable {
		t.Error("go_test should be test and executable")
	}

	// Providers
	if len(m.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(m.Providers))
	}
	goInfo := m.Providers[0]
	if goInfo.Name != "GoInfo" || len(goInfo.Fields) != 2 {
		t.Errorf("GoInfo = %+v", goInfo)
	}
	if goInfo.OriginKey == nil || goInfo.OriginKey.File != "@rules_go//go/private:providers.bzl" {
		t.Errorf("GoInfo OriginKey = %+v", goInfo.OriginKey)
	}

	// Functions
	if len(m.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(m.Functions))
	}
	goCtx := m.Functions[0]
	if goCtx.Name != "go_context" || goCtx.Returns != "A GoContext." {
		t.Errorf("go_context = %+v", goCtx)
	}
	if len(goCtx.Parameters) != 2 {
		t.Fatalf("go_context params: got %d", len(goCtx.Parameters))
	}
	if goCtx.Parameters[0].Role != ParamRoleOrdinary || !goCtx.Parameters[0].Mandatory {
		t.Errorf("ctx param = %+v", goCtx.Parameters[0])
	}
	if goCtx.Parameters[1].Role != ParamRoleKeywordOnly {
		t.Errorf("attrs param role = %v", goCtx.Parameters[1].Role)
	}
	if goCtx.Signature != "go_context(ctx, attrs)" {
		t.Errorf("go_context signature = %q", goCtx.Signature)
	}

	// Aspects
	if len(m.Aspects) != 1 {
		t.Fatalf("expected 1 aspect, got %d", len(m.Aspects))
	}
	aspect := m.Aspects[0]
	if aspect.Name != "go_archive_aspect" {
		t.Errorf("aspect name = %q", aspect.Name)
	}
	if len(aspect.AspectAttributes) != 2 {
		t.Errorf("aspect propagation attrs = %v", aspect.AspectAttributes)
	}

	t.Logf("Parsed: %d rules, %d providers, %d functions, %d aspects",
		len(m.Rules), len(m.Providers), len(m.Functions), len(m.Aspects))
}

func TestParseProtoFile(t *testing.T) {
	info := testModuleInfo()
	data, err := proto.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "rules.pb")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := ParseProtoFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(m.Rules))
	}
}

func TestParseProto_InvalidData(t *testing.T) {
	_, err := ParseProto([]byte("not a protobuf"))
	if err == nil {
		t.Error("expected error for invalid protobuf data")
	}
}

func TestParseProto_EmptyModule(t *testing.T) {
	info := &pb.ModuleInfo{}
	data, _ := proto.Marshal(info)

	m, err := ParseProto(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Rules) != 0 || len(m.Providers) != 0 {
		t.Error("expected empty module")
	}
}

func TestParseAll_MixedFormats(t *testing.T) {
	dir := t.TempDir()

	// Write a .md file
	os.WriteFile(filepath.Join(dir, "rules.md"), []byte(`<a id="md_rule"></a>

## md_rule

<pre>md_rule()</pre>

A Markdown rule.
`), 0o644)

	// Write a .pb file
	info := &pb.ModuleInfo{
		RuleInfo: []*pb.RuleInfo{
			{RuleName: "proto_rule", DocString: "A proto rule."},
		},
	}
	data, _ := proto.Marshal(info)
	os.WriteFile(filepath.Join(dir, "rules.pb"), data, 0o644)

	modules, err := ParseAll(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(modules) != 2 {
		t.Fatalf("expected 2 modules (1 md + 1 pb), got %d", len(modules))
	}

	// Verify both were parsed
	var foundMD, foundProto bool
	for _, m := range modules {
		for _, r := range m.Rules {
			if r.Name == "md_rule" {
				foundMD = true
			}
			if r.Name == "proto_rule" {
				foundProto = true
			}
		}
	}
	if !foundMD {
		t.Error("missing md_rule from Markdown file")
	}
	if !foundProto {
		t.Error("missing proto_rule from proto file")
	}
}

func TestAttributeType_String(t *testing.T) {
	tests := []struct {
		at   AttributeType
		want string
	}{
		{AttrName, "Name"},
		{AttrLabel, "Label"},
		{AttrLabelList, "List of labels"},
		{AttrBoolean, "Boolean"},
		{AttrStringDict, "Dictionary: String -> String"},
		{AttrUnknown, "unknown"},
		{AttributeType(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.at.String(); got != tt.want {
			t.Errorf("AttributeType(%d).String() = %q, want %q", tt.at, got, tt.want)
		}
	}
}

func TestBuildSignature(t *testing.T) {
	tests := []struct {
		name  string
		attrs []Attribute
		want  string
	}{
		{"empty", nil, "empty()"},
		{"one", []Attribute{{Name: "name"}}, "one(name)"},
		{"two", []Attribute{{Name: "a"}, {Name: "b"}}, "two(a, b)"},
	}
	for _, tt := range tests {
		if got := buildSignature(tt.name, tt.attrs); got != tt.want {
			t.Errorf("buildSignature(%q, %d attrs) = %q, want %q", tt.name, len(tt.attrs), got, tt.want)
		}
	}
}

func TestNameFromFileLabel(t *testing.T) {
	tests := []struct{ input, want string }{
		{"@rules_go//go:def.bzl", "def"},
		{"//foo/bar:baz.bzl", "baz"},
		{"simple.bzl", "simple"},
		{"@repo//pkg:rules.bzl", "rules"},
	}
	for _, tt := range tests {
		if got := nameFromFileLabel(tt.input); got != tt.want {
			t.Errorf("nameFromFileLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestParseProto_WriteFixture generates the test fixture for manual inspection.
// Run with: go test -run TestParseProto_WriteFixture -v
func TestParseProto_WriteFixture(t *testing.T) {
	info := testModuleInfo()
	data, err := proto.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", "rules_go_core.pb")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("Wrote %d bytes to %s", len(data), path)
}
