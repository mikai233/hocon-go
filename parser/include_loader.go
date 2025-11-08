package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"hocon-go/raw"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type includeContext struct {
	chain []string
}

func (ctx includeContext) push(path string) (includeContext, error) {
	for _, existing := range ctx.chain {
		if existing == path {
			return ctx, fmt.Errorf("include cycle detected for %s", path)
		}
	}
	newChain := make([]string, len(ctx.chain)+1)
	copy(newChain, ctx.chain)
	newChain[len(ctx.chain)] = path
	return includeContext{chain: newChain}, nil
}

func ParseFile(path string, opts ConfigOptions) (*raw.Object, error) {
	opts = normalizeOptions(opts)
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	ctx := includeContext{}
	ctx, err = ctx.push(abs)
	if err != nil {
		return nil, err
	}
	parser := newParser(data, opts, filepath.Dir(abs), ctx)
	return parser.Parse()
}

func (p *Parser) parseInclusion(inclusion *raw.Inclusion) error {
	loader := includeLoader{parser: p}
	obj, err := loader.load(inclusion)
	if err != nil {
		if inclusion.Required || !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	inclusion.Val = obj
	return nil
}

type includeLoader struct {
	parser *Parser
}

func (l includeLoader) load(inclusion *raw.Inclusion) (*raw.Object, error) {
	path := strings.TrimSpace(inclusion.Path)
	if path == "" {
		return nil, fmt.Errorf("include path is empty")
	}
	location := inclusion.Location
	switch {
	case location == nil:
		return l.loadFromFile(path)
	case *location == raw.File:
		return l.loadFromFile(path)
	case *location == raw.Classpath:
		return l.loadFromClasspath(path)
	default:
		return nil, fmt.Errorf("include location %s is not supported", location.String())
	}
}

func (l includeLoader) loadFromClasspath(path string) (*raw.Object, error) {
	if filepath.IsAbs(path) {
		return nil, fmt.Errorf("classpath include %q must be relative", path)
	}
	if len(l.parser.options.Classpath) == 0 {
		return nil, os.ErrNotExist
	}
	return l.loadFromBases(path, l.parser.options.Classpath)
}

func (l includeLoader) loadFromFile(path string) (*raw.Object, error) {
	var bases []string
	if filepath.IsAbs(path) {
		bases = []string{""}
	} else {
		if l.parser.baseDir != "" {
			bases = append(bases, l.parser.baseDir)
		}
		bases = append(bases, "")
	}
	return l.loadFromBases(path, bases)
}

func (l includeLoader) loadFromBases(path string, bases []string) (*raw.Object, error) {
	candidates := buildFileCandidates(path)
	seen := map[string]struct{}{}
	for _, base := range bases {
		for _, cand := range candidates {
			full := l.makeAbsolute(base, cand.path)
			key := fmt.Sprintf("%s|%d", full, cand.syntax)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			obj, err := l.openFile(full, cand.syntax)
			if err == nil {
				return obj, nil
			}
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}
	return nil, os.ErrNotExist
}

func (l includeLoader) makeAbsolute(base, path string) string {
	var combined string
	if filepath.IsAbs(path) {
		combined = path
	} else if base != "" {
		combined = filepath.Join(base, path)
	} else {
		combined = path
	}
	if abs, err := filepath.Abs(combined); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(combined)
}

func (l includeLoader) openFile(path string, syntax fileSyntax) (*raw.Object, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, os.ErrNotExist
	}
	switch syntax {
	case syntaxHocon:
		return l.parseHoconFile(path)
	case syntaxJSON:
		return parseJSONFile(path)
	default:
		return nil, fmt.Errorf("unsupported include syntax for %s", path)
	}
}

func (l includeLoader) parseHoconFile(path string) (*raw.Object, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	abs := filepath.Clean(path)
	childCtx, err := l.parser.ctx.push(abs)
	if err != nil {
		return nil, err
	}
	parser := newParser(data, l.parser.options, filepath.Dir(abs), childCtx)
	return parser.Parse()
}

func parseJSONFile(path string) (*raw.Object, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	var data interface{}
	if err := decoder.Decode(&data); err != nil && err != io.EOF {
		return nil, err
	}
	rawValue, err := jsonValueToRaw(data)
	if err != nil {
		return nil, err
	}
	obj, ok := rawValue.(*raw.Object)
	if !ok {
		return nil, fmt.Errorf("JSON root must be an object")
	}
	return obj, nil
}

func jsonValueToRaw(v interface{}) (raw.Value, error) {
	switch val := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fields := make([]raw.ObjectField, 0, len(keys))
		for _, k := range keys {
			child, err := jsonValueToRaw(val[k])
			if err != nil {
				return nil, err
			}
			fields = append(fields, raw.NewKeyValueField(raw.NewQuotedString(k), child))
		}
		return raw.NewObject(fields), nil
	case []interface{}:
		values := make([]raw.Value, len(val))
		for i, item := range val {
			child, err := jsonValueToRaw(item)
			if err != nil {
				return nil, err
			}
			values[i] = child
		}
		return raw.NewRawArray(values), nil
	case json.Number:
		if i64, err := val.Int64(); err == nil {
			if i64 >= 0 {
				return raw.NewPosInt(uint64(i64)), nil
			}
			return raw.NewNegInt(i64), nil
		}
		if f64, err := val.Float64(); err == nil {
			return raw.NewFloat(f64), nil
		}
		return nil, fmt.Errorf("invalid JSON number %q", val.String())
	case float64:
		if math.Trunc(val) == val {
			if val >= 0 {
				return raw.NewPosInt(uint64(val)), nil
			}
			return raw.NewNegInt(int64(val)), nil
		}
		return raw.NewFloat(val), nil
	case string:
		return raw.NewQuotedString(val), nil
	case bool:
		return raw.NewBoolean(val), nil
	case nil:
		return &raw.NULL, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value %T", v)
	}
}

type fileSyntax int

const (
	syntaxHocon fileSyntax = iota
	syntaxJSON
)

type fileCandidate struct {
	path   string
	syntax fileSyntax
}

func buildFileCandidates(path string) []fileCandidate {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".conf", ".hocon":
		return []fileCandidate{{path: path, syntax: syntaxHocon}}
	case ".json":
		return []fileCandidate{{path: path, syntax: syntaxJSON}}
	default:
		if ext != "" {
			return []fileCandidate{{path: path, syntax: syntaxHocon}}
		}
		return []fileCandidate{
			{path: path, syntax: syntaxHocon},
			{path: path + ".conf", syntax: syntaxHocon},
			{path: path + ".json", syntax: syntaxJSON},
		}
	}
}
