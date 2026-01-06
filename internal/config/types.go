package config

type Config struct {
	Provider  string              `mapstructure:"provider"`
	Providers map[string]Provider `mapstructure:"providers"`
	UI        UIConfig            `mapstructure:"ui"`
	System    SystemConfig        `mapstructure:"system"`
}

type Provider struct {
	APIKey     string `mapstructure:"api_key"`
	BaseURL    string `mapstructure:"base_url"`
	Model      string `mapstructure:"model"`
	AuthHeader string `mapstructure:"auth_header"`
	AuthPrefix string `mapstructure:"auth_prefix"`
}

type UIConfig struct {
	Theme              string `mapstructure:"theme"`
	WordWrap           int    `mapstructure:"word_wrap"`
	ShowTokens         bool   `mapstructure:"show_tokens"`
	SyntaxHighlighting bool   `mapstructure:"syntax_highlighting"`
}

type SystemConfig struct {
	Prompt string `mapstructure:"system_prompt"`
}
