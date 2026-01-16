package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// Schema represents a JSON Schema
type Schema struct {
	Schema      string             `json:"$schema,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Type        interface{}        `json:"type,omitempty"` // Can be string or []string for nullable
	Format      string             `json:"format,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Definitions map[string]*Schema `json:"definitions,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty"`
	AnyOf       []*Schema          `json:"anyOf,omitempty"`
	AllOf       []*Schema          `json:"allOf,omitempty"`
	Description string             `json:"description,omitempty"`
	// Custom extensions for complex types
	XGoType string `json:"x-go-type,omitempty"`
}

// ContractSchema represents the full schema for a contract
type ContractSchema struct {
	Schema      string             `json:"$schema"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	OneOf       []*Schema          `json:"oneOf,omitempty"`
	Definitions map[string]*Schema `json:"definitions"`
}

// TypeInfo holds parsed type information
type TypeInfo struct {
	Name   string
	Fields []FieldInfo
	IsMsg  bool // Is this a Msg* type (message type)
}

// FieldInfo holds field information
type FieldInfo struct {
	Name     string
	JSONName string
	Type     string
	Optional bool
	Comment  string
}

// Known type mappings for custom types
var typeMapping = map[string]*Schema{
	"sdkmath.Int": {
		Type:    "string",
		XGoType: "bigint",
		Format:  "bigint",
	},
	"math.Int": {
		Type:    "string",
		XGoType: "bigint",
		Format:  "bigint",
	},
	"Bech32String": {
		Type:   "string",
		Format: "bech32",
	},
	"wasmx.Bech32String": {
		Type:   "string",
		Format: "bech32",
	},
	"HexString": {
		Type:   "string",
		Format: "hex",
	},
	"wasmx.HexString": {
		Type:   "string",
		Format: "hex",
	},
	"ConsensusAddressString": {
		Type:   "string",
		Format: "consensus-address",
	},
	"wasmx.ConsensusAddressString": {
		Type:   "string",
		Format: "consensus-address",
	},
	"ValidatorAddressString": {
		Type:   "string",
		Format: "validator-address",
	},
	"wasmx.ValidatorAddressString": {
		Type:   "string",
		Format: "validator-address",
	},
	"Coin": {
		Type: "object",
		Properties: map[string]*Schema{
			"denom": {Type: "string"},
			"amount": {
				Type:    "string",
				XGoType: "bigint",
				Format:  "bigint",
			},
		},
		Required: []string{"denom", "amount"},
	},
	"wasmx.Coin": {
		Type: "object",
		Properties: map[string]*Schema{
			"denom": {Type: "string"},
			"amount": {
				Type:    "string",
				XGoType: "bigint",
				Format:  "bigint",
			},
		},
		Required: []string{"denom", "amount"},
	},
	"time.Duration": {
		Type:    "integer",
		Format:  "int64",
		XGoType: "duration-ns",
	},
}

func main() {
	var (
		modulePath string
		outputPath string
		moduleName string
	)

	flag.StringVar(&modulePath, "module", "", "Path to the TinyGo module directory")
	flag.StringVar(&outputPath, "output", "", "Output path for JSON schema (default: <module>/schema.json)")
	flag.StringVar(&moduleName, "name", "", "Module name for schema title (default: derived from path)")
	flag.Parse()

	if modulePath == "" {
		fmt.Fprintln(os.Stderr, "Usage: jsonschema -module <path> [-output <path>] [-name <name>]")
		os.Exit(1)
	}

	// Find lib directory
	libPath := filepath.Join(modulePath, "lib")
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		// Try gov subdirectory (for wasmx-gov, wasmx-gov-continuous)
		govPath := filepath.Join(modulePath, "gov")
		if _, err := os.Stat(govPath); err == nil {
			libPath = govPath
		} else {
			// Just use module path
			libPath = modulePath
		}
	}

	// Derive module name
	if moduleName == "" {
		moduleName = filepath.Base(modulePath)
	}

	// Default output path
	if outputPath == "" {
		outputPath = filepath.Join(modulePath, "schema.json")
	}

	// Parse all Go files in lib directory
	types, callDataType, instantiateType, err := parseGoFiles(libPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Go files: %v\n", err)
		os.Exit(1)
	}

	// Generate schema
	schema := generateSchema(moduleName, types, callDataType, instantiateType)

	// Write schema to file
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated JSON schema: %s\n", outputPath)
}

func parseGoFiles(dirPath string) (map[string]*TypeInfo, *TypeInfo, *TypeInfo, error) {
	fset := token.NewFileSet()
	types := make(map[string]*TypeInfo)

	var callDataType *TypeInfo
	var instantiateType *TypeInfo
	importDirs := map[string]struct{}{}
	aliasTypes := map[string]string{}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", filePath, err)
			continue
		}

		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if localDir, ok := resolveTinyGoImport(dirPath, path); ok {
				importDirs[localDir] = struct{}{}
			}
		}

		// Extract types from AST
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					typeInfo := parseStructType(typeSpec.Name.Name, structType)
					types[typeSpec.Name.Name] = typeInfo

					// Check for special types
					if typeSpec.Name.Name == "CallData" || typeSpec.Name.Name == "Calldata" {
						callDataType = typeInfo
					}
					if typeSpec.Name.Name == "CallDataInstantiate" || typeSpec.Name.Name == "CalldataInstantiate" || typeSpec.Name.Name == "InstantiateMsg" {
						instantiateType = typeInfo
					}
					continue
				}

				aliasTypes[typeSpec.Name.Name] = typeToString(typeSpec.Type)
			}
		}
	}

	for importDir := range importDirs {
		extraTypes, err := parseGoTypes(importDir)
		if err != nil {
			continue
		}
		for name, typeInfo := range extraTypes {
			if _, exists := types[name]; !exists {
				types[name] = typeInfo
			}
		}
	}

	for aliasName, targetType := range aliasTypes {
		targetName := targetType
		if strings.Contains(targetName, ".") {
			parts := strings.Split(targetName, ".")
			targetName = parts[len(parts)-1]
		}
		if target, ok := types[targetName]; ok {
			if _, exists := types[aliasName]; !exists {
				clone := *target
				clone.Name = aliasName
				clone.Fields = append([]FieldInfo(nil), target.Fields...)
				types[aliasName] = &clone
			}
		}
	}

	return types, callDataType, instantiateType, nil
}

func parseGoTypes(dirPath string) (map[string]*TypeInfo, error) {
	fset := token.NewFileSet()
	types := make(map[string]*TypeInfo)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				typeInfo := parseStructType(typeSpec.Name.Name, structType)
				types[typeSpec.Name.Name] = typeInfo
			}
		}
	}

	return types, nil
}

func resolveTinyGoImport(moduleDir, importPath string) (string, bool) {
	const marker = "/tests/testdata/tinygo/"
	moduleDir = filepath.ToSlash(moduleDir)
	idx := strings.Index(moduleDir, marker)
	if idx == -1 {
		return "", false
	}

	root := moduleDir[:idx+len(marker)]
	impIdx := strings.Index(importPath, marker)
	if impIdx == -1 {
		return "", false
	}
	subpath := importPath[impIdx+len(marker):]
	localDir := filepath.Join(filepath.FromSlash(root), filepath.FromSlash(subpath))
	if _, err := os.Stat(localDir); err != nil {
		return "", false
	}
	return localDir, true
}

func parseStructType(name string, structType *ast.StructType) *TypeInfo {
	typeInfo := &TypeInfo{
		Name:  name,
		IsMsg: strings.HasPrefix(name, "Msg"),
	}

	if structType.Fields == nil {
		return typeInfo
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldInfo := FieldInfo{
			Name: field.Names[0].Name,
			Type: typeToString(field.Type),
		}

		// Parse JSON tag
		if field.Tag != nil {
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			jsonTag := tag.Get("json")
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				fieldInfo.JSONName = parts[0]
				for _, part := range parts[1:] {
					if part == "omitempty" {
						fieldInfo.Optional = true
					}
				}
			}
		}

		if fieldInfo.JSONName == "" {
			fieldInfo.JSONName = fieldInfo.Name
		}

		// Check if it's a pointer type (optional)
		if strings.HasPrefix(fieldInfo.Type, "*") {
			fieldInfo.Optional = true
			fieldInfo.Type = strings.TrimPrefix(fieldInfo.Type, "*")
		}

		typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
	}

	return typeInfo
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

func generateSchema(moduleName string, types map[string]*TypeInfo, callDataType *TypeInfo, instantiateType *TypeInfo) *ContractSchema {
	schema := &ContractSchema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Title:       moduleName,
		Description: fmt.Sprintf("JSON Schema for %s TinyGo contract", moduleName),
		Type:        "object",
		Definitions: make(map[string]*Schema),
	}

	// Add all type definitions
	for name, typeInfo := range types {
		// Skip CallData itself (it's the root)
		if callDataType != nil && name == callDataType.Name {
			continue
		}
		if instantiateType != nil && name == instantiateType.Name {
			continue
		}
		schema.Definitions[name] = typeInfoToSchema(typeInfo, types)
	}

	// Generate oneOf for CallData (execute messages)
	if callDataType != nil {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		schema.Definitions["main"] = typeInfoToSchema(callDataType, types)
		schema.Properties["main"] = &Schema{Ref: "#/definitions/main"}

		for _, field := range callDataType.Fields {
			if field.JSONName == "-" {
				continue
			}

			msgSchema := &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					field.JSONName: fieldToSchema(field, types),
				},
				Required: []string{field.JSONName},
			}
			schema.OneOf = append(schema.OneOf, msgSchema)
		}
	}

	// Add instantiate type if exists
	if instantiateType != nil {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		instantiateSchema := typeInfoToSchema(instantiateType, types)
		schema.Definitions["instantiate"] = instantiateSchema
		schema.Definitions["InstantiateMsg"] = instantiateSchema
		schema.Properties["instantiate"] = &Schema{Ref: "#/definitions/instantiate"}
	}

	return schema
}

func typeInfoToSchema(typeInfo *TypeInfo, allTypes map[string]*TypeInfo) *Schema {
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}

	for _, field := range typeInfo.Fields {
		if field.JSONName == "-" {
			continue
		}
		schema.Properties[field.JSONName] = fieldToSchema(field, allTypes)
		if !field.Optional {
			schema.Required = append(schema.Required, field.JSONName)
		}
	}

	return schema
}

func fieldToSchema(field FieldInfo, allTypes map[string]*TypeInfo) *Schema {
	return goTypeToSchema(field.Type, allTypes)
}

func goTypeToSchema(goType string, allTypes map[string]*TypeInfo) *Schema {
	// Check known type mappings first
	if mapped, ok := typeMapping[goType]; ok {
		// Return a copy to avoid mutations
		return copySchema(mapped)
	}

	// Handle arrays/slices
	if strings.HasPrefix(goType, "[]") {
		elemType := strings.TrimPrefix(goType, "[]")
		return &Schema{
			Type:  "array",
			Items: goTypeToSchema(elemType, allTypes),
		}
	}

	// Handle maps
	if strings.HasPrefix(goType, "map[") {
		re := regexp.MustCompile(`map\[([^\]]+)\](.+)`)
		matches := re.FindStringSubmatch(goType)
		if len(matches) == 3 {
			return &Schema{
				Type: "object",
				// JSON only supports string keys
				Properties: map[string]*Schema{},
			}
		}
	}

	// Handle basic Go types
	switch goType {
	case "string":
		return &Schema{Type: "string"}
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return &Schema{Type: "integer"}
	case "float32", "float64":
		return &Schema{Type: "number"}
	case "bool":
		return &Schema{Type: "boolean"}
	case "byte":
		return &Schema{Type: "integer"}
	case "[]byte":
		return &Schema{Type: "string", Format: "base64"}
	case "interface{}", "any":
		return &Schema{} // Empty schema accepts anything
	}

	// Check if it's a known struct type
	if _, exists := allTypes[goType]; exists {
		return &Schema{Ref: "#/definitions/" + goType}
	}

	// Handle selector types (package.Type) that aren't in mapping
	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		typeName := parts[len(parts)-1]
		// Check if base type name is mapped
		if mapped, ok := typeMapping[typeName]; ok {
			return copySchema(mapped)
		}
		// Check if it's a known struct
		if _, exists := allTypes[typeName]; exists {
			return &Schema{Ref: "#/definitions/" + typeName}
		}
	}

	// Default to string for unknown types
	return &Schema{Type: "string", XGoType: goType}
}

func copySchema(s *Schema) *Schema {
	if s == nil {
		return nil
	}
	copy := *s
	if s.Properties != nil {
		copy.Properties = make(map[string]*Schema)
		for k, v := range s.Properties {
			copy.Properties[k] = copySchema(v)
		}
	}
	if s.Required != nil {
		copy.Required = append([]string{}, s.Required...)
	}
	return &copy
}
