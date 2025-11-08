package config

import (
	"bytes"
	"fmt"
	"hocon-go/parser"
	"hocon-go/raw"
	"io"
	"net/http"
)

// Config represents a parsed HOCON document before resolution.
type Config struct {
	rawObj *raw.Object
	opts   parser.ConfigOptions
}

// ParseFile reads the file at path and returns a Config.
func ParseFile(path string, opts *parser.ConfigOptions) (*Config, error) {
	options := normalizeOptions(opts)
	obj, err := parser.ParseFile(path, options)
	if err != nil {
		return nil, err
	}
	return &Config{rawObj: obj, opts: options}, nil
}

// ParseReader reads all data from r and parses it as HOCON.
func ParseReader(r io.Reader, opts *parser.ConfigOptions) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return parseBytes(data, opts)
}

// ParseString parses the given string as HOCON.
func ParseString(s string, opts *parser.ConfigOptions) (*Config, error) {
	return parseBytes([]byte(s), opts)
}

// ParseBytes parses the given byte slice as HOCON.
func parseBytes(data []byte, opts *parser.ConfigOptions) (*Config, error) {
	options := normalizeOptions(opts)
	parser := parser.NewParser(data).WithOptions(options)
	obj, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return &Config{rawObj: obj, opts: options}, nil
}

// ParseURL downloads the resource located at url and parses it as HOCON.
func ParseURL(url string, opts *parser.ConfigOptions) (*Config, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch %s: %s", url, resp.Status)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, err
	}
	return parseBytes(buf.Bytes(), opts)
}

// Resolve converts the configuration into regular Go values (maps, slices, scalars).
func (c *Config) Resolve() (map[string]interface{}, error) {
	if c == nil || c.rawObj == nil {
		return map[string]interface{}{}, nil
	}
	obj, err := buildMergeObject(nil, c.rawObj)
	if err != nil {
		return nil, err
	}
	if err := obj.Substitute(); err != nil {
		return nil, err
	}
	obj.ResolveAddAssign()
	obj.TryBecomeMerged()
	res, err := objectToInterface(obj)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MustResolve resolves the configuration and panics on failure.
func (c *Config) MustResolve() map[string]interface{} {
	res, err := c.Resolve()
	if err != nil {
		panic(err)
	}
	return res
}

// Load resolves the configuration at path and returns the resulting structure.
func Load(path string, opts *parser.ConfigOptions) (map[string]interface{}, error) {
	cfg, err := ParseFile(path, opts)
	if err != nil {
		return nil, err
	}
	return cfg.Resolve()
}

func normalizeOptions(opts *parser.ConfigOptions) parser.ConfigOptions {
	if opts == nil {
		return parser.DefaultConfigOptions()
	}
	resolved := *opts
	def := parser.DefaultConfigOptions()
	if resolved.MaxDepth == 0 {
		resolved.MaxDepth = def.MaxDepth
	}
	if resolved.MaxIncludeDepth == 0 {
		resolved.MaxIncludeDepth = def.MaxIncludeDepth
	}
	return resolved
}
