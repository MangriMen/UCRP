package openapi

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Spec struct {
	raw      map[string]any
	paths    map[string]any
	baseDir  string
	external map[string]map[string]any
}

func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi %s: %w", path, err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse openapi %s: %w", path, err)
	}
	paths, ok := raw["paths"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("openapi %s has no paths object", path)
	}
	return &Spec{
		raw:      raw,
		paths:    paths,
		baseDir:  filepath.Dir(path),
		external: map[string]map[string]any{},
	}, nil
}

func (s *Spec) Operation(method, path string) (map[string]any, bool) {
	template, ok := s.matchPath(path)
	if !ok {
		return nil, false
	}
	item, ok := s.resolveNode(s.paths[template])
	if !ok {
		return nil, false
	}
	op, ok := s.resolveNode(item[strings.ToLower(method)])
	if !ok {
		return nil, false
	}
	return op, true
}

func (s *Spec) matchPath(path string) (string, bool) {
	if _, ok := s.paths[path]; ok {
		return path, true
	}
	names := make([]string, 0, len(s.paths))
	for name := range s.paths {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if pathMatches(name, path) {
			return name, true
		}
	}
	return "", false
}

func pathMatches(template, actual string) bool {
	tpl := strings.Split(strings.Trim(template, "/"), "/")
	act := strings.Split(strings.Trim(actual, "/"), "/")
	if len(tpl) != len(act) {
		return false
	}
	for i := range tpl {
		if strings.HasPrefix(tpl[i], "{") && strings.HasSuffix(tpl[i], "}") {
			if act[i] == "" {
				return false
			}
			continue
		}
		if tpl[i] != act[i] {
			return false
		}
	}
	return true
}

func (s *Spec) DeclaredStatuses(method, path string) []string {
	op, ok := s.Operation(method, path)
	if !ok {
		return nil
	}
	responses, ok := op["responses"].(map[string]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(responses))
	for code := range responses {
		out = append(out, code)
	}
	sort.Strings(out)
	return out
}

func (s *Spec) ResponseSchema(method, path string, status int) (map[string]any, bool) {
	op, ok := s.Operation(method, path)
	if !ok {
		return nil, false
	}
	responses, ok := op["responses"].(map[string]any)
	if !ok {
		return nil, false
	}
	entry, ok := s.resolveNode(responses[strconv.Itoa(status)])
	if !ok {
		return nil, false
	}
	content, ok := entry["content"].(map[string]any)
	if !ok {
		return nil, false
	}
	media, ok := content["application/json"].(map[string]any)
	if !ok {
		return nil, false
	}
	schema, ok := media["schema"].(map[string]any)
	if !ok {
		return nil, false
	}
	resolved, ok := s.resolveNode(schema)
	return resolved, ok
}

func (s *Spec) resolveNode(node any) (map[string]any, bool) {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil, false
	}
	if ref, ok := obj["$ref"].(string); ok {
		return s.lookup(ref)
	}
	return obj, true
}

func (s *Spec) lookup(ref string) (map[string]any, bool) {
	file, fragment, _ := strings.Cut(ref, "#")
	var current any
	if file == "" {
		current = s.raw
	} else {
		doc, ok := s.loadExternal(file)
		if !ok {
			return nil, false
		}
		current = doc
	}
	if fragment == "" || fragment == "/" {
		out, ok := current.(map[string]any)
		return out, ok
	}
	for _, part := range strings.Split(strings.TrimPrefix(fragment, "/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		obj, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = obj[part]
		if !ok {
			return nil, false
		}
	}
	out, ok := current.(map[string]any)
	if !ok {
		return nil, false
	}
	if refTarget, ok := out["$ref"].(string); ok {
		return s.lookup(refTarget)
	}
	return out, true
}

func (s *Spec) loadExternal(file string) (map[string]any, bool) {
	if doc, ok := s.external[file]; ok {
		return doc, true
	}
	path := file
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.baseDir, filepath.FromSlash(file))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, false
	}
	s.external[file] = doc
	return doc, true
}

func (s *Spec) Validate(schema map[string]any, value any) []string {
	var problems []string
	s.validate(schema, value, "$", &problems)
	return problems
}

func (s *Spec) ValidateResponse(method, path string, status int, body any) []string {
	schema, ok := s.ResponseSchema(method, path, status)
	if !ok {
		return nil
	}
	return s.Validate(schema, body)
}

func (s *Spec) validate(schema map[string]any, value any, at string, problems *[]string) {
	if schema == nil {
		return
	}
	if ref, ok := schema["$ref"].(string); ok {
		target, found := s.lookup(ref)
		if !found {
			*problems = append(*problems, fmt.Sprintf("%s: unresolvable $ref %q", at, ref))
			return
		}
		s.validate(target, value, at, problems)
		return
	}

	if value == nil {
		if nullable, _ := schema["nullable"].(bool); nullable {
			return
		}
		if _, hasType := schema["type"]; hasType {
			*problems = append(*problems, fmt.Sprintf("%s: null is not allowed", at))
		}
		return
	}

	if parts, ok := schema["allOf"].([]any); ok {
		for i, part := range parts {
			sub, ok := part.(map[string]any)
			if !ok {
				continue
			}
			s.validate(sub, value, fmt.Sprintf("%s/allOf[%d]", at, i), problems)
		}
	}

	if branches, ok := schema["oneOf"].([]any); ok {
		var fallback []string
		for _, raw := range branches {
			sub, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			var branchProblems []string
			s.validate(sub, value, at, &branchProblems)
			if len(branchProblems) == 0 {
				fallback = nil
				break
			}
			if fallback == nil || len(branchProblems) < len(fallback) {
				fallback = branchProblems
			}
		}
		if fallback != nil {
			*problems = append(*problems, fallback...)
		}
		return
	}

	if enum, ok := schema["enum"].([]any); ok {
		if !containsValue(enum, value) {
			*problems = append(*problems, fmt.Sprintf("%s: %v is not one of %v", at, value, enum))
		}
	}

	if typ, ok := schema["type"].(string); ok {
		if !typeMatches(typ, value) {
			*problems = append(*problems, fmt.Sprintf("%s: expected %s, got %T (%v)", at, typ, value, value))
			return
		}
	}

	switch typed := value.(type) {
	case map[string]any:
		s.validateObject(schema, typed, at, problems)
	case []any:
		s.validateArray(schema, typed, at, problems)
	case string:
		s.validateString(schema, typed, at, problems)
	case float64:
		s.validateNumber(schema, typed, at, problems)
	}
}

func (s *Spec) validateObject(schema map[string]any, value map[string]any, at string, problems *[]string) {
	if required, ok := schema["required"].([]any); ok {
		for _, name := range required {
			key, _ := name.(string)
			if _, present := value[key]; !present {
				*problems = append(*problems, fmt.Sprintf("%s: missing required property %q", at, key))
			}
		}
	}
	properties, _ := schema["properties"].(map[string]any)
	for key, sub := range properties {
		child, present := value[key]
		if !present {
			continue
		}
		subSchema, ok := sub.(map[string]any)
		if !ok {
			continue
		}
		s.validate(subSchema, child, at+"."+key, problems)
	}
	if additional, ok := schema["additionalProperties"]; ok {
		switch rule := additional.(type) {
		case bool:
			if !rule {
				for key := range value {
					if _, declared := properties[key]; !declared {
						*problems = append(*problems, fmt.Sprintf("%s: unexpected property %q", at, key))
					}
				}
			}
		case map[string]any:
			for key, child := range value {
				if _, declared := properties[key]; declared {
					continue
				}
				s.validate(rule, child, at+"."+key, problems)
			}
		}
	}
}

func (s *Spec) validateArray(schema map[string]any, value []any, at string, problems *[]string) {
	if min, ok := intOf(schema["minItems"]); ok && len(value) < min {
		*problems = append(*problems, fmt.Sprintf("%s: expected at least %d items, got %d", at, min, len(value)))
	}
	if max, ok := intOf(schema["maxItems"]); ok && len(value) > max {
		*problems = append(*problems, fmt.Sprintf("%s: expected at most %d items, got %d", at, max, len(value)))
	}
	items, ok := s.resolveNode(schema["items"])
	if !ok {
		return
	}
	for i, child := range value {
		s.validate(items, child, fmt.Sprintf("%s[%d]", at, i), problems)
	}
}

func (s *Spec) validateString(schema map[string]any, value, at string, problems *[]string) {
	if min, ok := intOf(schema["minLength"]); ok && len([]rune(value)) < min {
		*problems = append(*problems, fmt.Sprintf("%s: length %d is below minLength %d", at, len([]rune(value)), min))
	}
	if max, ok := intOf(schema["maxLength"]); ok && len([]rune(value)) > max {
		*problems = append(*problems, fmt.Sprintf("%s: length %d exceeds maxLength %d", at, len([]rune(value)), max))
	}
	format, _ := schema["format"].(string)
	switch format {
	case "date-time":
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			*problems = append(*problems, fmt.Sprintf("%s: %q is not an RFC 3339 date-time", at, value))
		}
	case "email":
		if !strings.Contains(value, "@") {
			*problems = append(*problems, fmt.Sprintf("%s: %q is not an email", at, value))
		}
	}
}

func (s *Spec) validateNumber(schema map[string]any, value float64, at string, problems *[]string) {
	if min, ok := floatOf(schema["minimum"]); ok && value < min {
		*problems = append(*problems, fmt.Sprintf("%s: %v is below minimum %v", at, value, min))
	}
	if max, ok := floatOf(schema["maximum"]); ok && value > max {
		*problems = append(*problems, fmt.Sprintf("%s: %v exceeds maximum %v", at, value, max))
	}
}

func typeMatches(typ string, value any) bool {
	switch typ {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "integer":
		n, ok := value.(float64)
		return ok && math.Trunc(n) == n
	case "number":
		_, ok := value.(float64)
		return ok
	}
	return true
}

func containsValue(enum []any, value any) bool {
	for _, candidate := range enum {
		if fmt.Sprint(candidate) == fmt.Sprint(value) {
			return true
		}
	}
	return false
}

func intOf(v any) (int, bool) {
	n, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return int(n), true
}

func floatOf(v any) (float64, bool) {
	n, ok := v.(float64)
	return n, ok
}
