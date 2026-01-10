package main

import (
	"bufio"
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
	"sort"
	"strings"
)

type Schema struct {
	Schema                 string             `json:"$schema,omitempty"`
	Ref                    string             `json:"$ref,omitempty"`
	Type                   interface{}        `json:"type,omitempty"`
	Format                 string             `json:"format,omitempty"`
	Properties             map[string]*Schema `json:"properties,omitempty"`
	Items                  *Schema            `json:"items,omitempty"`
	Required               []string           `json:"required,omitempty"`
	OneOf                  []*Schema          `json:"oneOf,omitempty"`
	Enum                   []string           `json:"enum,omitempty"`
	Description            string             `json:"description,omitempty"`
	AdditionalProperties   *Schema            `json:"additionalProperties,omitempty"`
	XProtoFullName         string             `json:"x-proto-full-name,omitempty"`
	XProtoPackage          string             `json:"x-proto-package,omitempty"`
	XProtoMessageShortName string             `json:"x-proto-message,omitempty"`
}

type ContractSchema struct {
	Schema      string             `json:"$schema"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Type        string             `json:"type"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Definitions map[string]*Schema `json:"definitions,omitempty"`
}

type FieldInfo struct {
	Name     string
	JSONName string
	Type     string
	Optional bool
}

type TypeInfo struct {
	Name   string
	Fields []FieldInfo
}

type ModuleConfig struct {
	Name           string
	ProtoTxPath    string
	ProtoQueryPath string
	GoTypesDir     string
	UseProto       bool
	Allowlist      map[string]bool
}

type messageEntry struct {
	Key    string
	Name   string
	Module string
	Kind   string
}

type moduleData struct {
	Config  ModuleConfig
	Types   map[string]*TypeInfo
	Entries []messageEntry
}

func main() {
	outPath := flag.String("out", "", "output JSON schema path (file)")
	outDir := flag.String("out-dir", "", "output directory path (writes per module/tx/query)")
	title := flag.String("title", "wasmx-message-schemas", "schema title")
	flag.Parse()

	if *outPath == "" && *outDir == "" {
		fmt.Fprintln(os.Stderr, "missing -out or -out-dir")
		os.Exit(1)
	}

	modules, err := collectModules()
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect modules: %v\n", err)
		os.Exit(1)
	}

	if *outDir != "" {
		if err := writePerModuleSchemas(*outDir, *title, modules); err != nil {
			fmt.Fprintf(os.Stderr, "write schemas: %v\n", err)
			os.Exit(1)
		}
	}

	if *outPath != "" {
		if err := writeCombinedSchema(*outPath, *title, modules); err != nil {
			fmt.Fprintf(os.Stderr, "write schema: %v\n", err)
			os.Exit(1)
		}
	}
}

func collectModules() ([]moduleData, error) {
	modules := []ModuleConfig{
		{
			Name:           "wasmx",
			ProtoTxPath:    filepath.Join("wasmx", "proto", "mythos", "wasmx", "v1", "tx.proto"),
			ProtoQueryPath: filepath.Join("wasmx", "proto", "mythos", "wasmx", "v1", "query.proto"),
			GoTypesDir:     filepath.Join("wasmx", "x", "wasmx", "types"),
			UseProto:       true,
		},
		{
			Name:           "cosmosmod",
			ProtoTxPath:    "",
			ProtoQueryPath: "",
			GoTypesDir:     filepath.Join("wasmx", "x", "cosmosmod", "types"),
			UseProto:       false,
		},
		{
			Name:           "network",
			ProtoTxPath:    filepath.Join("wasmx", "proto", "mythos", "network", "v1", "tx.proto"),
			ProtoQueryPath: filepath.Join("wasmx", "proto", "mythos", "network", "v1", "query.proto"),
			GoTypesDir:     filepath.Join("wasmx", "x", "network", "types"),
			UseProto:       true,
			Allowlist: map[string]bool{
				"MsgMultiChainWrap":         true,
				"MsgExecuteAtomicTxRequest": true,
				"QueryMultiChainRequest":    true,
			},
		},
	}

	results := make([]moduleData, 0, len(modules))
	for _, cfg := range modules {
		types, err := parseGoTypes(cfg.GoTypesDir)
		if err != nil {
			return nil, fmt.Errorf("parse go types for %s: %w", cfg.Name, err)
		}

		entries, err := collectEntries(cfg, types)
		if err != nil {
			return nil, fmt.Errorf("collect entries for %s: %w", cfg.Name, err)
		}

		results = append(results, moduleData{
			Config:  cfg,
			Types:   types,
			Entries: entries,
		})
	}

	return results, nil
}

func collectEntries(cfg ModuleConfig, types map[string]*TypeInfo) ([]messageEntry, error) {
	var entries []messageEntry

	if cfg.UseProto {
		txMsgs, err := parseServiceMessages(cfg.ProtoTxPath, "Msg")
		if err != nil {
			return nil, err
		}
		queryMsgs, err := parseServiceMessages(cfg.ProtoQueryPath, "Query")
		if err != nil {
			return nil, err
		}
		entries = append(entries, buildEntries(cfg, txMsgs, "tx")...)
		entries = append(entries, buildEntries(cfg, queryMsgs, "query")...)
	} else {
		var txMsgs []string
		var queryMsgs []string
		for name := range types {
			if strings.HasPrefix(name, "Msg") && !strings.HasSuffix(name, "Response") {
				txMsgs = append(txMsgs, name)
			}
			if strings.HasPrefix(name, "Query") && strings.HasSuffix(name, "Request") {
				queryMsgs = append(queryMsgs, name)
			}
		}
		sort.Strings(txMsgs)
		sort.Strings(queryMsgs)
		entries = append(entries, buildEntries(cfg, txMsgs, "tx")...)
		entries = append(entries, buildEntries(cfg, queryMsgs, "query")...)
	}

	filtered := entries[:0]
	for _, entry := range entries {
		if _, ok := types[entry.Name]; !ok {
			fmt.Fprintf(os.Stderr, "warning: %s missing Go type for %s\n", cfg.Name, entry.Name)
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered, nil
}

func buildEntries(cfg ModuleConfig, names []string, kind string) []messageEntry {
	entries := make([]messageEntry, 0, len(names))
	for _, name := range names {
		if cfg.Allowlist != nil && !cfg.Allowlist[name] {
			continue
		}
		entries = append(entries, messageEntry{
			Key:    cfg.Name + "." + name,
			Name:   name,
			Module: cfg.Name,
			Kind:   kind,
		})
	}
	return entries
}

func parseServiceMessages(path string, serviceName string) ([]string, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open proto %s: %w", path, err)
	}
	defer file.Close()

	serviceRe := regexp.MustCompile(`^\s*service\s+(\w+)\b`)
	rpcRe := regexp.MustCompile(`^\s*rpc\s+\w+\s*\(\s*([.\w]+)\s*\)`)

	var messages []string
	inService := false
	depth := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !inService {
			match := serviceRe.FindStringSubmatch(line)
			if len(match) == 2 && match[1] == serviceName {
				inService = true
				depth = braceDelta(line)
			}
			continue
		}

		if match := rpcRe.FindStringSubmatch(line); len(match) == 2 {
			msg := match[1]
			if idx := strings.LastIndex(msg, "."); idx >= 0 {
				msg = msg[idx+1:]
			}
			messages = append(messages, msg)
		}

		depth += braceDelta(line)
		if depth <= 0 {
			inService = false
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan proto %s: %w", path, err)
	}
	return messages, nil
}

func braceDelta(line string) int {
	return strings.Count(line, "{") - strings.Count(line, "}")
}

func parseGoTypes(dirPath string) (map[string]*TypeInfo, error) {
	types := make(map[string]*TypeInfo)
	fset := token.NewFileSet()

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
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", filePath, err)
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

				info := parseStructType(typeSpec.Name.Name, structType)
				types[info.Name] = info
			}
		}
	}

	return types, nil
}

func parseStructType(name string, structType *ast.StructType) *TypeInfo {
	info := &TypeInfo{Name: name}
	if structType.Fields == nil {
		return info
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldInfo := FieldInfo{
			Name: field.Names[0].Name,
			Type: typeToString(field.Type),
		}

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

		if strings.HasPrefix(fieldInfo.Type, "*") {
			fieldInfo.Optional = true
			fieldInfo.Type = strings.TrimPrefix(fieldInfo.Type, "*")
		}

		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
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

func writePerModuleSchemas(outDir string, title string, modules []moduleData) error {
	for _, module := range modules {
		if len(module.Entries) == 0 {
			continue
		}

		grouped := make(map[string][]messageEntry)
		for _, entry := range module.Entries {
			grouped[entry.Kind] = append(grouped[entry.Kind], entry)
		}

		for kind, entries := range grouped {
			output := filepath.Join(outDir, module.Config.Name, kind+".json")
			schema := buildSchema(title+"-"+module.Config.Name+"-"+kind, module.Config.Name, entries, module.Types)
			if err := writeSchemaFile(output, schema); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeCombinedSchema(outPath string, title string, modules []moduleData) error {
	var entries []messageEntry
	typesByModule := make(map[string]map[string]*TypeInfo)

	for _, module := range modules {
		entries = append(entries, module.Entries...)
		typesByModule[module.Config.Name] = module.Types
	}

	if len(entries) == 0 {
		return fmt.Errorf("no messages selected")
	}

	definitions := make(map[string]*Schema)
	mainProps := make(map[string]*Schema)

	for _, entry := range entries {
		moduleTypes := typesByModule[entry.Module]
		if moduleTypes == nil {
			continue
		}
		ensureDefinition(entry.Module, entry.Name, moduleTypes, definitions)
		mainProps[entry.Key] = &Schema{
			Ref:                    "#/definitions/" + definitionKey(entry.Module, entry.Name),
			XProtoFullName:         entry.Key,
			XProtoPackage:          entry.Module,
			XProtoMessageShortName: entry.Name,
		}
	}

	definitions["main"] = &Schema{
		Type:       "object",
		Properties: mainProps,
	}

	schema := &ContractSchema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Title:       title,
		Description: "JSON schema for governance message encoding",
		Type:        "object",
		Properties: map[string]*Schema{
			"main": {Ref: "#/definitions/main"},
		},
		Definitions: definitions,
	}

	return writeSchemaFile(outPath, schema)
}

func buildSchema(title string, module string, entries []messageEntry, types map[string]*TypeInfo) *ContractSchema {
	definitions := make(map[string]*Schema)
	mainProps := make(map[string]*Schema)

	for _, entry := range entries {
		ensureDefinition(module, entry.Name, types, definitions)
		mainProps[entry.Key] = &Schema{
			Ref:                    "#/definitions/" + definitionKey(module, entry.Name),
			XProtoFullName:         entry.Key,
			XProtoPackage:          module,
			XProtoMessageShortName: entry.Name,
		}
	}

	definitions["main"] = &Schema{
		Type:       "object",
		Properties: mainProps,
	}

	return &ContractSchema{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Title:       title,
		Description: "JSON schema for governance message encoding",
		Type:        "object",
		Properties: map[string]*Schema{
			"main": {Ref: "#/definitions/main"},
		},
		Definitions: definitions,
	}
}

func writeSchemaFile(outPath string, schema *ContractSchema) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}
	payload, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema: %w", err)
	}
	if err := os.WriteFile(outPath, payload, 0o644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}
	return nil
}

func definitionKey(module string, typeName string) string {
	return module + "." + typeName
}

func ensureDefinition(module string, typeName string, types map[string]*TypeInfo, definitions map[string]*Schema) {
	key := definitionKey(module, typeName)
	if _, ok := definitions[key]; ok {
		return
	}

	info, ok := types[typeName]
	if !ok {
		return
	}

	// Pre-seed to break recursion.
	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]*Schema),
	}
	definitions[key] = schema

	required := make([]string, 0)
	for _, field := range info.Fields {
		fieldSchema := schemaForType(module, field.Type, types, definitions)
		if fieldSchema != nil {
			schema.Properties[field.JSONName] = fieldSchema
		}
		if !field.Optional {
			required = append(required, field.JSONName)
		}
	}

	if len(required) > 0 {
		schema.Required = required
	}
}

func schemaForType(module string, typeName string, types map[string]*TypeInfo, definitions map[string]*Schema) *Schema {
	if typeName == "[]byte" {
		return &Schema{Type: "string", Format: "base64"}
	}

	if strings.HasPrefix(typeName, "[]") {
		return &Schema{
			Type:  "array",
			Items: schemaForType(module, strings.TrimPrefix(typeName, "[]"), types, definitions),
		}
	}

	if strings.HasPrefix(typeName, "map[") {
		value := parseMapValue(typeName)
		return &Schema{
			Type:                 "object",
			AdditionalProperties: schemaForType(module, value, types, definitions),
		}
	}

	if strings.HasPrefix(typeName, "*") {
		return schemaForType(module, strings.TrimPrefix(typeName, "*"), types, definitions)
	}

	if mapped := mapKnownType(typeName); mapped != nil {
		return mapped
	}

	if _, ok := types[typeName]; ok {
		ensureDefinition(module, typeName, types, definitions)
		return &Schema{Ref: "#/definitions/" + definitionKey(module, typeName)}
	}

	return &Schema{Type: "string"}
}

func parseMapValue(typeName string) string {
	if !strings.HasPrefix(typeName, "map[") {
		return "string"
	}
	closeIdx := strings.Index(typeName, "]")
	if closeIdx == -1 || closeIdx+1 >= len(typeName) {
		return "string"
	}
	return typeName[closeIdx+1:]
}

func mapKnownType(typeName string) *Schema {
	switch typeName {
	case "string":
		return &Schema{Type: "string"}
	case "bool":
		return &Schema{Type: "boolean"}
	case "float32", "float64":
		return &Schema{Type: "number"}
	case "int32":
		return &Schema{Type: "string", Format: "int32"}
	case "uint32":
		return &Schema{Type: "string", Format: "uint32"}
	case "int64":
		return &Schema{Type: "string", Format: "int64"}
	case "uint64":
		return &Schema{Type: "string", Format: "uint64"}
	}

	if strings.HasSuffix(typeName, ".Any") || typeName == "Any" {
		return &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"type_url": {Type: "string"},
				"value":    {Type: "string", Format: "base64"},
			},
		}
	}

	if isCoinType(typeName) {
		return coinSchema()
	}

	if isCoinsType(typeName) {
		return &Schema{
			Type:  "array",
			Items: coinSchema(),
		}
	}

	if isDecCoinType(typeName) {
		return decCoinSchema()
	}

	if isDecCoinsType(typeName) {
		return &Schema{
			Type:  "array",
			Items: decCoinSchema(),
		}
	}

	return nil
}

func coinSchema() *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"denom":  {Type: "string"},
			"amount": {Type: "string", Format: "bigint"},
		},
		Required: []string{"denom", "amount"},
	}
}

func decCoinSchema() *Schema {
	return &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"denom":  {Type: "string"},
			"amount": {Type: "string", Format: "decimal"},
		},
		Required: []string{"denom", "amount"},
	}
}

func isCoinType(typeName string) bool {
	return typeName == "Coin" || strings.HasSuffix(typeName, ".Coin")
}

func isCoinsType(typeName string) bool {
	return typeName == "Coins" || strings.HasSuffix(typeName, ".Coins")
}

func isDecCoinType(typeName string) bool {
	return typeName == "DecCoin" || strings.HasSuffix(typeName, ".DecCoin")
}

func isDecCoinsType(typeName string) bool {
	return typeName == "DecCoins" || strings.HasSuffix(typeName, ".DecCoins")
}
