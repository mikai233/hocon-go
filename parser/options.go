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
