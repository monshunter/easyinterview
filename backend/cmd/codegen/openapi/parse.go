package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Conventions mirrors the subset of shared/conventions.yaml the openapi
// codegen needs (enum names + values, documented error codes, jobStatuses,
// PageInfo / ApiError structures).
type Conventions struct {
	Errors      []conventionError           `yaml:"errors"`
	JobStatuses []string                    `yaml:"jobStatuses"`
	Enums       []conventionEnum            `yaml:"enums"`
	Structures  map[string]conventionStruct `yaml:"structures"`
}

type conventionError struct {
	Code      string `yaml:"code"`
	Message   string `yaml:"message"`
	Retryable bool   `yaml:"retryable"`
}

type conventionEnum struct {
	Name          string   `yaml:"name"`
	SourceSection string   `yaml:"sourceSection"`
	JSONField     string   `yaml:"jsonField"`
	Values        []string `yaml:"values"`
}

type conventionStruct struct {
	Fields []conventionField `yaml:"fields"`
}

type conventionField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Optional bool   `yaml:"optional"`
	Nullable bool   `yaml:"nullable"`
}

func loadConventions(path string) (*Conventions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Conventions
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// EnumNames returns convention enum names in the order they appear in conventions.yaml.
func (c *Conventions) EnumNames() []string {
	names := make([]string, 0, len(c.Enums))
	for _, e := range c.Enums {
		names = append(names, e.Name)
	}
	return names
}

// ErrorCodes returns sorted documented error codes plus PRIVACY_EXPORT_NOT_AVAILABLE.
// The privacy-export code is co-owned by spec D-12 and surfaces in OpenAPI even
// though B1 conventions.yaml does not list it.
func (c *Conventions) ErrorCodes() []string {
	seen := map[string]struct{}{
		"PRIVACY_EXPORT_NOT_AVAILABLE": {},
	}
	for _, e := range c.Errors {
		seen[e.Code] = struct{}{}
	}
	codes := make([]string, 0, len(seen))
	for code := range seen {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// EnumByName returns the enum spec, or nil if absent.
func (c *Conventions) EnumByName(name string) *conventionEnum {
	for i := range c.Enums {
		if c.Enums[i].Name == name {
			return &c.Enums[i]
		}
	}
	return nil
}

// OpenAPI is a parsed view of openapi.yaml as a yaml.Node tree, with helpers
// to extract structured information (paths, operations, schemas).
//
// We deliberately keep the parser narrow: only the structural data the Go
// and TS templates actually consume is materialised, and unsupported shapes
// (free-form `oneOf` without nullability, dynamic `additionalProperties`)
// fall through as `any` rather than blocking codegen.
type OpenAPI struct {
	Raw     []byte
	Doc     map[string]any
	Version string
}

func loadOpenAPI(path string) (*OpenAPI, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	info, _ := doc["info"].(map[string]any)
	version, _ := info["version"].(string)
	if version == "" {
		return nil, fmt.Errorf("info.version missing")
	}
	return &OpenAPI{Raw: data, Doc: doc, Version: version}, nil
}

// Paths returns spec §3.1.1 operations in **declaration order** (the YAML
// map order preserved by yaml.v3 is good enough for our golden files).
func (o *OpenAPI) Paths() []OpsForPath {
	out := []OpsForPath{}
	rawPaths, _ := o.Doc["paths"].(map[string]any)
	keys := make([]string, 0, len(rawPaths))
	for k := range rawPaths {
		keys = append(keys, k)
	}
	// Stable sort keys to keep output byte-stable across runs even if the
	// underlying map iteration order changes.
	sort.Strings(keys)
	for _, k := range keys {
		entry, _ := rawPaths[k].(map[string]any)
		out = append(out, OpsForPath{Path: k, Item: entry})
	}
	return out
}

// OpsForPath bundles a path with its raw method→operation map.
type OpsForPath struct {
	Path string
	Item map[string]any
}

// Methods returns the verb keys (get/post/...) in canonical sorted order.
func (p *OpsForPath) Methods() []string {
	verbs := []string{}
	for k := range p.Item {
		switch k {
		case "get", "post", "put", "patch", "delete", "head", "options":
			verbs = append(verbs, k)
		}
	}
	sort.Strings(verbs)
	return verbs
}

// Schemas returns components.schemas as raw map.
func (o *OpenAPI) Schemas() map[string]any {
	comps, _ := o.Doc["components"].(map[string]any)
	schemas, _ := comps["schemas"].(map[string]any)
	return schemas
}

// SortedSchemaNames returns components.schemas keys in deterministic
// (declaration-aware) order.
//
// We use the order they appear in the YAML node tree by re-parsing as a
// yaml.Node — that keeps the generated Go file order stable and matches
// what humans see when reading openapi.yaml top-to-bottom.
func (o *OpenAPI) SortedSchemaNames() ([]string, error) {
	var rootNode yaml.Node
	if err := yaml.Unmarshal(o.Raw, &rootNode); err != nil {
		return nil, err
	}
	if rootNode.Kind != yaml.DocumentNode || len(rootNode.Content) == 0 {
		return nil, fmt.Errorf("unexpected yaml root")
	}
	mapping := rootNode.Content[0]
	componentsNode := findMappingValue(mapping, "components")
	if componentsNode == nil {
		return nil, nil
	}
	schemasNode := findMappingValue(componentsNode, "schemas")
	if schemasNode == nil {
		return nil, nil
	}
	names := []string{}
	for i := 0; i < len(schemasNode.Content); i += 2 {
		names = append(names, schemasNode.Content[i].Value)
	}
	return names, nil
}

func findMappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// extractRefName returns "Foo" for "#/components/schemas/Foo".
func extractRefName(ref string) string {
	prefix := "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return ""
	}
	return strings.TrimPrefix(ref, prefix)
}
