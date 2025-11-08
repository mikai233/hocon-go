package parser

const (
	defaultMaxDepth        = 64
	defaultMaxIncludeDepth = 64
)

// ConfigOptions contains parser level configuration flags.
type ConfigOptions struct {
	UseSystemEnvironment bool
	Classpath            []string
	MaxDepth             int
	MaxIncludeDepth      int
}

func DefaultConfigOptions() ConfigOptions {
	return ConfigOptions{
		UseSystemEnvironment: false,
		Classpath:            nil,
		MaxDepth:             defaultMaxDepth,
		MaxIncludeDepth:      defaultMaxIncludeDepth,
	}
}

func normalizeOptions(opts ConfigOptions) ConfigOptions {
	if opts.MaxDepth == 0 {
		opts.MaxDepth = defaultMaxDepth
	}
	if opts.MaxIncludeDepth == 0 {
		opts.MaxIncludeDepth = defaultMaxIncludeDepth
	}
	return opts
}
