package openapi

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mvanhorn/cli-printing-press/internal/spec"
)

const (
	maxResources            = 50
	maxEndpointsPerResource = 20
)

func Parse(data []byte) (*spec.APISpec, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("loading OpenAPI spec: %w", err)
	}

	doc.InternalizeRefs(context.Background(), nil)

	name := "api"
	description := ""
	version := ""
	if doc.Info != nil {
		if v := toKebabCase(doc.Info.Title); v != "" {
			name = v
		}
		description = strings.TrimSpace(doc.Info.Description)
		version = strings.TrimSpace(doc.Info.Version)
	}

	baseURL := ""
	if len(doc.Servers) > 0 && doc.Servers[0] != nil {
		baseURL = strings.TrimRight(strings.TrimSpace(doc.Servers[0].URL), "/")
	}

	result := &spec.APISpec{
		Name:        name,
		Description: description,
		Version:     version,
		BaseURL:     baseURL,
		Auth:        mapAuth(doc, name),
		Config: spec.ConfigSpec{
			Format: "toml",
			Path:   fmt.Sprintf("~/.config/%s-cli/config.toml", name),
		},
		Resources: map[string]spec.Resource{},
		Types:     map[string]spec.TypeDef{},
	}

	mapResources(doc, result, baseURLPath(baseURL))
	mapTypes(doc, result)

	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("validating parsed spec: %w", err)
	}

	return result, nil
}

func mapAuth(doc *openapi3.T, name string) spec.AuthConfig {
	auth := spec.AuthConfig{Type: "none"}
	schemeName, scheme := selectSecurityScheme(doc)
	if scheme == nil {
		return auth
	}

	auth.Scheme = schemeName

	switch strings.ToLower(scheme.Type) {
	case "http":
		switch strings.ToLower(scheme.Scheme) {
		case "bearer":
			auth.Type = "bearer_token"
			auth.Header = "Authorization"
		case "basic":
			auth.Type = "api_key"
			auth.Header = "Authorization"
			auth.Format = "Basic {username}:{password}"
		}
	case "apikey":
		auth.Type = "api_key"
		auth.Header = strings.TrimSpace(scheme.Name)
		if auth.Header == "" {
			auth.Header = "Authorization"
		}
		auth.In = strings.TrimSpace(scheme.In)
	case "oauth2":
		auth.Type = "bearer_token"
		auth.Header = "Authorization"
	}

	envPrefix := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	switch auth.Type {
	case "api_key":
		auth.EnvVars = []string{envPrefix + "_API_KEY"}
	case "bearer_token":
		auth.EnvVars = []string{envPrefix + "_TOKEN"}
	}

	return auth
}

func selectSecurityScheme(doc *openapi3.T) (string, *openapi3.SecurityScheme) {
	if doc == nil || doc.Components == nil || len(doc.Components.SecuritySchemes) == 0 {
		return "", nil
	}

	orderedNames := orderedSecuritySchemeNames(doc)
	for _, name := range orderedNames {
		scheme := securitySchemeValue(doc.Components.SecuritySchemes[name])
		if scheme == nil {
			continue
		}
		switch strings.ToLower(scheme.Type) {
		case "apikey", "oauth2":
			return name, scheme
		case "http":
			switch strings.ToLower(scheme.Scheme) {
			case "bearer", "basic":
				return name, scheme
			}
		}
	}

	for _, name := range orderedNames {
		scheme := securitySchemeValue(doc.Components.SecuritySchemes[name])
		if scheme != nil {
			return name, scheme
		}
	}

	return "", nil
}

func orderedSecuritySchemeNames(doc *openapi3.T) []string {
	seen := map[string]struct{}{}
	var names []string

	for _, requirement := range doc.Security {
		var requirementNames []string
		for name := range requirement {
			requirementNames = append(requirementNames, name)
		}
		sort.Strings(requirementNames)
		for _, name := range requirementNames {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	var all []string
	for name := range doc.Components.SecuritySchemes {
		all = append(all, name)
	}
	sort.Strings(all)
	for _, name := range all {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	return names
}

func securitySchemeValue(ref *openapi3.SecuritySchemeRef) *openapi3.SecurityScheme {
	if ref == nil {
		return nil
	}
	return ref.Value
}

func mapResources(doc *openapi3.T, out *spec.APISpec, basePath string) {
	if doc == nil || out == nil || doc.Paths == nil {
		return
	}

	tagDescriptions := mapTagDescriptions(doc.Tags)

	pathMap := doc.Paths.Map()
	pathKeys := make([]string, 0, len(pathMap))
	for path := range pathMap {
		pathKeys = append(pathKeys, path)
	}
	sort.Strings(pathKeys)

	for _, path := range pathKeys {
		pathItem := doc.Paths.Value(path)
		if pathItem == nil {
			warnf("skipping path %q: path item is nil", path)
			continue
		}

		operations := pathItem.Operations()
		if len(operations) == 0 {
			warnf("skipping path %q: no valid HTTP methods", path)
			continue
		}

		resourceName := resourceNameFromPath(path, basePath)
		if resourceName == "" {
			warnf("skipping path %q: could not derive resource name", path)
			continue
		}

		resource, ok := out.Resources[resourceName]
		if !ok {
			if len(out.Resources) >= maxResources {
				warnf("skipping path %q: resource limit (%d) reached", path, maxResources)
				continue
			}
			resource = spec.Resource{
				Description: tagDescriptions[resourceName],
				Endpoints:   map[string]spec.Endpoint{},
			}
		}

		methods := make([]string, 0, len(operations))
		for method := range operations {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		for _, method := range methods {
			op := operations[method]
			if op == nil {
				warnf("skipping %s %q: operation is nil", method, path)
				continue
			}

			if len(resource.Endpoints) >= maxEndpointsPerResource {
				warnf("skipping %s %q: endpoint limit (%d) reached for resource %q", method, path, maxEndpointsPerResource, resourceName)
				continue
			}

			endpointName := resolveEndpointName(method, path, op, resource.Endpoints, resourceName, basePath)
			endpoint := spec.Endpoint{
				Method:      strings.ToUpper(method),
				Path:        path,
				Description: firstNonEmpty(strings.TrimSpace(op.Summary), strings.TrimSpace(op.Description)),
				Params:      mapParameters(pathItem, op),
				Body:        mapRequestBody(op.RequestBody, method, path),
			}

			endpoint.Response, endpoint.ResponsePath = mapResponse(op, endpointName)
			if resource.Description == "" {
				resource.Description = resourceDescription(op, tagDescriptions)
			}
			resource.Endpoints[endpointName] = endpoint
		}

		out.Resources[resourceName] = resource
	}
}

func mapTagDescriptions(tags openapi3.Tags) map[string]string {
	out := make(map[string]string, len(tags))
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		name := toSnakeCase(tag.Name)
		if name == "" {
			continue
		}
		if desc := strings.TrimSpace(tag.Description); desc != "" {
			out[name] = desc
		}
	}
	return out
}

func resourceDescription(op *openapi3.Operation, tagDescriptions map[string]string) string {
	if op == nil {
		return ""
	}
	for _, tag := range op.Tags {
		if desc := tagDescriptions[toSnakeCase(tag)]; desc != "" {
			return desc
		}
	}
	return ""
}

func resolveEndpointName(method, path string, op *openapi3.Operation, existing map[string]spec.Endpoint, resourceName, basePath string) string {
	name := operationIDToName(operationID(op))
	if name == "" {
		name = defaultEndpointName(method, path)
	}
	if name == "" {
		name = strings.ToLower(method)
	}

	if _, ok := existing[name]; !ok {
		return name
	}

	suffix := endpointCollisionSuffix(path, resourceName, basePath)
	if suffix == "" {
		suffix = "endpoint"
	}

	candidate := name + "-" + suffix
	if _, ok := existing[candidate]; !ok {
		return candidate
	}

	for i := 2; ; i++ {
		alt := fmt.Sprintf("%s-%s-%d", name, suffix, i)
		if _, ok := existing[alt]; !ok {
			return alt
		}
	}
}

func operationID(op *openapi3.Operation) string {
	if op == nil {
		return ""
	}
	return strings.TrimSpace(op.OperationID)
}

func defaultEndpointName(method, path string) string {
	switch strings.ToUpper(method) {
	case "GET":
		if hasPathParams(path) {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func hasPathParams(path string) bool {
	return strings.Contains(path, "{") && strings.Contains(path, "}")
}

func mapParameters(pathItem *openapi3.PathItem, op *openapi3.Operation) []spec.Param {
	merged := mergeParameters(pathItem, op)
	params := make([]spec.Param, 0, len(merged))
	for _, parameter := range merged {
		if parameter == nil {
			continue
		}
		if parameter.In != openapi3.ParameterInPath && parameter.In != openapi3.ParameterInQuery {
			continue
		}

		schema := schemaRefValue(parameter.Schema)
		param := spec.Param{
			Name:        parameter.Name,
			Type:        mapSchemaType(schema),
			Required:    parameter.Required,
			Positional:  parameter.In == openapi3.ParameterInPath,
			Description: strings.TrimSpace(parameter.Description),
			Enum:        schemaEnum(schema),
			Format:      schemaFormat(schema),
		}
		if schema != nil && schema.Default != nil {
			param.Default = schema.Default
		}
		if param.Positional {
			param.Required = true
		}
		params = append(params, param)
	}
	return params
}

func mergeParameters(pathItem *openapi3.PathItem, op *openapi3.Operation) []*openapi3.Parameter {
	var merged []*openapi3.Parameter
	index := map[string]int{}

	add := func(parameters openapi3.Parameters, override bool) {
		for _, parameterRef := range parameters {
			if parameterRef == nil || parameterRef.Value == nil {
				continue
			}
			parameter := parameterRef.Value
			key := strings.ToLower(parameter.In) + ":" + parameter.Name
			if i, ok := index[key]; ok {
				if override {
					merged[i] = parameter
				}
				continue
			}
			index[key] = len(merged)
			merged = append(merged, parameter)
		}
	}

	if pathItem != nil {
		add(pathItem.Parameters, false)
	}
	if op != nil {
		add(op.Parameters, true)
	}

	return merged
}

func mapRequestBody(requestBodyRef *openapi3.RequestBodyRef, method, path string) []spec.Param {
	requestBody := requestBodyValue(requestBodyRef)
	if requestBody == nil || requestBody.Content == nil {
		return nil
	}

	media := requestBody.Content.Get("application/json")
	if media == nil {
		media = firstJSONMediaType(requestBody.Content)
	}
	if media == nil || media.Schema == nil || media.Schema.Value == nil {
		return nil
	}

	properties := map[string]*openapi3.SchemaRef{}
	required := map[string]struct{}{}
	if collectAllOfProperties(media.Schema, properties, required, map[*openapi3.Schema]struct{}{}) {
		warnf("skipping request body for %s %q: contains oneOf/anyOf", strings.ToUpper(method), path)
		return nil
	}

	if len(properties) == 0 {
		return nil
	}

	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)

	body := make([]spec.Param, 0, len(names))
	for _, name := range names {
		schema := schemaRefValue(properties[name])
		param := spec.Param{
			Name:        name,
			Type:        mapSchemaType(schema),
			Required:    isRequired(required, name),
			Description: schemaDescription(schema),
			Enum:        schemaEnum(schema),
			Format:      schemaFormat(schema),
		}
		if schema != nil && schema.Default != nil {
			param.Default = schema.Default
		}
		body = append(body, param)
	}

	return body
}

func collectAllOfProperties(
	schemaRef *openapi3.SchemaRef,
	properties map[string]*openapi3.SchemaRef,
	required map[string]struct{},
	visited map[*openapi3.Schema]struct{},
) bool {
	if schemaRef == nil || schemaRef.Value == nil {
		return false
	}

	schema := schemaRef.Value
	if _, ok := visited[schema]; ok {
		return false
	}
	visited[schema] = struct{}{}

	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		return true
	}

	for _, field := range schema.Required {
		required[field] = struct{}{}
	}
	for name, prop := range schema.Properties {
		if prop == nil {
			continue
		}
		properties[name] = prop
	}
	for _, sub := range schema.AllOf {
		if collectAllOfProperties(sub, properties, required, visited) {
			return true
		}
	}

	return false
}

func mapResponse(op *openapi3.Operation, fallbackName string) (spec.ResponseDef, string) {
	if op == nil || op.Responses == nil {
		return spec.ResponseDef{}, ""
	}

	success := selectSuccessResponse(op.Responses)
	if success == nil || success.Value == nil {
		return spec.ResponseDef{}, ""
	}

	schemaRef := selectResponseSchema(success.Value)
	if schemaRef == nil || schemaRef.Value == nil {
		return spec.ResponseDef{}, ""
	}

	schema := schemaRef.Value
	if isObjectSchema(schema) {
		if dataRef := schema.Properties["data"]; dataRef != nil && isArraySchema(schemaRefValue(dataRef)) {
			return spec.ResponseDef{
				Type: "array",
				Item: schemaTypeName(schemaRefValue(dataRef).Items, fallbackName+"Item"),
			}, "data"
		}
	}

	if isArraySchema(schema) {
		return spec.ResponseDef{
			Type: "array",
			Item: schemaTypeName(schema.Items, fallbackName+"Item"),
		}, ""
	}

	if isObjectSchema(schema) {
		return spec.ResponseDef{
			Type: "object",
			Item: schemaTypeName(schemaRef, fallbackName+"Response"),
		}, ""
	}

	return spec.ResponseDef{}, ""
}

func selectSuccessResponse(responses *openapi3.Responses) *openapi3.ResponseRef {
	if responses == nil {
		return nil
	}
	if v := responses.Value("200"); v != nil {
		return v
	}
	if v := responses.Value("201"); v != nil {
		return v
	}

	responseMap := responses.Map()
	keys := make([]string, 0, len(responseMap))
	for key := range responseMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if len(key) == 3 && key[0] == '2' {
			return responses.Value(key)
		}
		if strings.EqualFold(key, "2XX") {
			return responses.Value(key)
		}
	}

	return nil
}

func selectResponseSchema(response *openapi3.Response) *openapi3.SchemaRef {
	if response == nil || response.Content == nil {
		return nil
	}
	if media := response.Content.Get("application/json"); media != nil && media.Schema != nil {
		return media.Schema
	}

	contentTypes := make([]string, 0, len(response.Content))
	for contentType := range response.Content {
		contentTypes = append(contentTypes, contentType)
	}
	sort.Strings(contentTypes)
	for _, contentType := range contentTypes {
		media := response.Content[contentType]
		if media != nil && media.Schema != nil {
			return media.Schema
		}
	}

	return nil
}

func mapTypes(doc *openapi3.T, out *spec.APISpec) {
	if doc == nil || doc.Components == nil {
		return
	}

	schemaMap := doc.Components.Schemas
	if len(schemaMap) == 0 {
		return
	}

	names := make([]string, 0, len(schemaMap))
	for name := range schemaMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		schemaRef := schemaMap[name]
		schema := schemaRefValue(schemaRef)
		if schema == nil {
			warnf("skipping schema %q: schema is nil", name)
			continue
		}
		if !isObjectSchema(schema) {
			continue
		}

		properties := map[string]*openapi3.SchemaRef{}
		collectTypeProperties(schemaRef, properties, map[*openapi3.Schema]struct{}{})

		fieldNames := make([]string, 0, len(properties))
		for fieldName := range properties {
			fieldNames = append(fieldNames, fieldName)
		}
		sort.Strings(fieldNames)

		fields := make([]spec.TypeField, 0, len(fieldNames))
		for _, fieldName := range fieldNames {
			fields = append(fields, spec.TypeField{
				Name: fieldName,
				Type: mapSchemaType(schemaRefValue(properties[fieldName])),
			})
		}

		out.Types[name] = spec.TypeDef{Fields: fields}
	}
}

func collectTypeProperties(schemaRef *openapi3.SchemaRef, properties map[string]*openapi3.SchemaRef, visited map[*openapi3.Schema]struct{}) {
	if schemaRef == nil || schemaRef.Value == nil {
		return
	}

	schema := schemaRef.Value
	if _, ok := visited[schema]; ok {
		return
	}
	visited[schema] = struct{}{}

	for name, prop := range schema.Properties {
		if prop == nil {
			continue
		}
		properties[name] = prop
	}
	for _, sub := range schema.AllOf {
		collectTypeProperties(sub, properties, visited)
	}
}

func requestBodyValue(ref *openapi3.RequestBodyRef) *openapi3.RequestBody {
	if ref == nil {
		return nil
	}
	return ref.Value
}

func schemaRefValue(ref *openapi3.SchemaRef) *openapi3.Schema {
	if ref == nil {
		return nil
	}
	return ref.Value
}

func mapSchemaType(schema *openapi3.Schema) string {
	if schema == nil || schema.Type == nil {
		return "string"
	}
	switch {
	case schema.Type.Is(openapi3.TypeString):
		return "string"
	case schema.Type.Is(openapi3.TypeInteger):
		return "int"
	case schema.Type.Is(openapi3.TypeBoolean):
		return "bool"
	case schema.Type.Is(openapi3.TypeNumber):
		return "float"
	default:
		return "string"
	}
}

func schemaEnum(schema *openapi3.Schema) []string {
	if schema == nil || len(schema.Enum) == 0 {
		return nil
	}
	enum := make([]string, 0, len(schema.Enum))
	for _, value := range schema.Enum {
		switch v := value.(type) {
		case string:
			enum = append(enum, v)
		default:
			enum = append(enum, fmt.Sprint(v))
		}
	}
	return enum
}

func schemaFormat(schema *openapi3.Schema) string {
	if schema == nil {
		return ""
	}
	return strings.TrimSpace(schema.Format)
}

func schemaDescription(schema *openapi3.Schema) string {
	if schema == nil {
		return ""
	}
	return strings.TrimSpace(schema.Description)
}

func isArraySchema(schema *openapi3.Schema) bool {
	if schema == nil {
		return false
	}
	if schema.Type != nil && schema.Type.Is(openapi3.TypeArray) {
		return true
	}
	return schema.Items != nil
}

func isObjectSchema(schema *openapi3.Schema) bool {
	if schema == nil {
		return false
	}
	if schema.Type != nil && schema.Type.Is(openapi3.TypeObject) {
		return true
	}
	return len(schema.Properties) > 0 || len(schema.AllOf) > 0
}

func schemaTypeName(schemaRef *openapi3.SchemaRef, fallback string) string {
	if schemaRef == nil {
		return toTypeName(fallback)
	}

	if refName := refComponentName(schemaRef.Ref); refName != "" {
		return toTypeName(refName)
	}

	schema := schemaRef.Value
	if schema == nil {
		return toTypeName(fallback)
	}
	if schema.Title != "" {
		return toTypeName(schema.Title)
	}

	if schema.Type != nil {
		switch {
		case schema.Type.Is(openapi3.TypeString):
			return "string"
		case schema.Type.Is(openapi3.TypeInteger):
			return "int"
		case schema.Type.Is(openapi3.TypeBoolean):
			return "bool"
		case schema.Type.Is(openapi3.TypeNumber):
			return "float"
		}
	}

	return toTypeName(fallback)
}

func refComponentName(ref string) string {
	if ref == "" {
		return ""
	}
	i := strings.LastIndex(ref, "/")
	if i == -1 || i+1 >= len(ref) {
		return ""
	}
	return ref[i+1:]
}

func firstJSONMediaType(content openapi3.Content) *openapi3.MediaType {
	if content == nil {
		return nil
	}

	contentTypes := make([]string, 0, len(content))
	for contentType := range content {
		contentTypes = append(contentTypes, contentType)
	}
	sort.Strings(contentTypes)

	for _, contentType := range contentTypes {
		if strings.Contains(strings.ToLower(contentType), "json") {
			return content[contentType]
		}
	}

	for _, contentType := range contentTypes {
		media := content[contentType]
		if media != nil && media.Schema != nil {
			return media
		}
	}

	return nil
}

func resourceNameFromPath(path, basePath string) string {
	segments := pathSegmentsAfterBase(path, basePath)
	if len(segments) == 0 {
		return ""
	}
	if isPathParamSegment(segments[0]) {
		return ""
	}
	return sanitizeResourceName(toSnakeCase(segments[0]))
}

func endpointCollisionSuffix(path, resourceName, basePath string) string {
	segments := pathSegmentsAfterBase(path, basePath)
	if len(segments) == 0 {
		return ""
	}
	if toSnakeCase(segments[0]) == resourceName {
		segments = segments[1:]
	}

	for _, segment := range segments {
		if isPathParamSegment(segment) {
			continue
		}
		if suffix := toKebabCase(segment); suffix != "" {
			return suffix
		}
	}
	for _, segment := range segments {
		segment = strings.Trim(segment, "{}")
		if suffix := toKebabCase(segment); suffix != "" {
			return suffix
		}
	}

	return ""
}

func pathSegmentsAfterBase(path, basePath string) []string {
	segments := splitPath(path)
	if len(segments) == 0 {
		return nil
	}

	baseSegments := splitPath(basePath)
	if len(baseSegments) > 0 && len(segments) >= len(baseSegments) {
		match := true
		for i := range baseSegments {
			if segments[i] != baseSegments[i] {
				match = false
				break
			}
		}
		if match {
			segments = segments[len(baseSegments):]
		}
	}

	if len(segments) > 0 && isVersionSegment(segments[0]) {
		segments = segments[1:]
	}

	return segments
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	raw := strings.Split(path, "/")
	segments := make([]string, 0, len(raw))
	for _, segment := range raw {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	return segments
}

func isVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	_, err := strconv.Atoi(segment[1:])
	return err == nil
}

func isPathParamSegment(segment string) bool {
	return strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}

func baseURLPath(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return parsed.Path
}

func operationIDToName(operationID string) string {
	if operationID == "" {
		return ""
	}
	return toSnakeCase(operationID)
}

func toSnakeCase(input string) string {
	var b strings.Builder
	var prev rune
	lastUnderscore := true

	for i, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if unicode.IsUpper(r) && i > 0 && (unicode.IsLower(prev) || unicode.IsDigit(prev)) && !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		} else if !lastUnderscore && b.Len() > 0 {
			b.WriteByte('_')
			lastUnderscore = true
		}
		prev = r
	}

	return strings.Trim(b.String(), "_")
}

func sanitizeResourceName(name string) string {
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.Trim(name, "_")
	if name == "" {
		return ""
	}
	return name
}

func toKebabCase(input string) string {
	var b strings.Builder
	lastHyphen := true

	for _, r := range input {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastHyphen = false
		case unicode.IsSpace(r):
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func toTypeName(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "Item"
	}

	var parts []string
	var current strings.Builder

	flush := func() {
		if current.Len() == 0 {
			return
		}
		parts = append(parts, current.String())
		current.Reset()
	}

	for i, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if i > 0 && unicode.IsUpper(r) {
				prev := rune(input[i-1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					flush()
				}
			}
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	if len(parts) == 0 {
		return "Item"
	}

	var b strings.Builder
	for _, part := range parts {
		part = strings.ToLower(part)
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}

	result := b.String()
	if result == "" {
		return "Item"
	}
	if unicode.IsDigit(rune(result[0])) {
		return "Type" + result
	}
	return result
}

func isRequired(required map[string]struct{}, name string) bool {
	_, ok := required[name]
	return ok
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func warnf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "warning: "+format+"\n", args...)
}
