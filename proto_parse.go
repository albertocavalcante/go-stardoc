package stardoc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/albertocavalcante/go-stardoc/gen"
	"google.golang.org/protobuf/proto"
)

// ParseProto parses a binary stardoc protobuf (ModuleInfo) into a [Module].
//
// This is the preferred parsing method when protobuf output is available,
// as it provides richer type information (e.g., [AttributeType] enum,
// [OriginKey] cross-references) than the Markdown parser.
//
// The module Name is derived from the ModuleInfo.file label; callers may
// override it after parsing.
func ParseProto(data []byte) (*Module, error) {
	info := &pb.ModuleInfo{}
	if err := proto.Unmarshal(data, info); err != nil {
		return nil, fmt.Errorf("stardoc: unmarshal proto: %w", err)
	}
	return moduleFromProto(info), nil
}

// ParseProtoFile parses a binary stardoc protobuf file (.pb or .binaryproto)
// from disk.
//
// The module Name is set to the base filename without extension if the
// protobuf does not contain a file label.
func ParseProtoFile(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %w", err)
	}
	m, err := ParseProto(data)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %s: %w", path, err)
	}
	if m.Name == "" {
		m.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return m, nil
}

// ParseProtoReader parses a binary stardoc protobuf from an [io.Reader].
func ParseProtoReader(r io.Reader) (*Module, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("stardoc: %w", err)
	}
	return ParseProto(data)
}

// isProtoFile reports whether a filename looks like a stardoc protobuf output.
func isProtoFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".pb" || ext == ".binaryproto" || ext == ".binpb"
}
