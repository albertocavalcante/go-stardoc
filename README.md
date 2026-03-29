# go-stardoc

Go library for parsing [Stardoc](https://github.com/bazelbuild/stardoc)-generated documentation into structured types.

Supports both **Markdown** output (from stardoc Velocity templates) and **binary protobuf** output (from `starlark_doc_extract`).

## Install

```bash
go get github.com/albertocavalcante/go-stardoc@latest
```

## Usage

### Parse Markdown

```go
m, err := stardoc.ParseFile("docs/rules.md")
for _, r := range m.Rules {
    fmt.Printf("%s: %d attributes\n", r.Name, len(r.Attributes))
}
```

### Parse Protobuf

```go
m, err := stardoc.ParseProtoFile("docs/rules.pb")
for _, r := range m.Rules {
    fmt.Printf("%s (test=%v, exec=%v)\n", r.Name, r.Test, r.Executable)
}
```

### Auto-detect format

```go
// ParseAll reads .md and .pb files from a directory
modules, err := stardoc.ParseAll("docs/")
```

## Types

| Type | Description |
|------|-------------|
| `Module` | One .bzl file's documentation (rules, providers, functions, aspects) |
| `Rule` | Starlark rule with attributes, signature, providers |
| `Provider` | Provider with fields |
| `Function` | Function/macro with parameters and return type |
| `Aspect` | Aspect with propagation attributes |
| `Attribute` | Rule attribute with type, default, mandatory flag |
| `OriginKey` | Cross-reference key (proto only) |
| `AttributeType` | Strongly-typed attribute enum (proto only) |

## Proto support

Binary protobuf parsing uses the vendored `stardoc_output.proto` from [bazelbuild/bazel](https://github.com/bazelbuild/bazel) (Apache 2.0). Proto types are generated with [buf](https://buf.build/) into the `gen/` package.

Proto parsing provides richer data than Markdown: `OriginKey` for cross-references, `AttributeType` enum, `Test`/`Executable` flags, `Nonconfigurable` attribute marking, and `MacroInfo`/`RepositoryRuleInfo` support.

## License

Apache 2.0
