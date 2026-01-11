package viamstreamdeck

import (
	"fmt"
	"slices"
)

type KeyConfig struct {
	Key int

	Text      string
	TextColor string `json:"text_color"`

	Color string
	Image string

	Component string
	Method    string
	Args      []interface{}
}

func (kc *KeyConfig) Validate() error {
	if kc.Component == "" {
		return fmt.Errorf("need a component")
	}
	if kc.Method == "" {
		return fmt.Errorf("need a component")
	}
	return nil
}

func (kc *KeyConfig) snakeMethod() string {
	return snakeToCamel(kc.Method)
}

type DialConfig struct {
	Dial      int
	Component string
	Command   string
}

func (dc *DialConfig) Validate() error {
	if dc.Component == "" {
		return fmt.Errorf("need a component")
	}
	if dc.Command == "" {
		return fmt.Errorf("need a command")
	}
	return nil
}

type Config struct {
	Brightness  int
	Keys        []KeyConfig            `json:"keys,omitempty"`
	Pages       map[string][]KeyConfig `json:"pages,omitempty"`
	InitialPage string                 `json:"initial_page,omitempty"`
	Dials       []DialConfig
}

func (c *Config) Validate(p string) ([]string, []string, error) {
	ret := []string{}

	// Validate that we have either Keys or Pages, but not both
	if len(c.Keys) > 0 && len(c.Pages) > 0 {
		return nil, nil, fmt.Errorf("cannot specify both 'keys' and 'pages' in config")
	}

	// If neither is specified, that's an error
	if len(c.Keys) == 0 && len(c.Pages) == 0 {
		return nil, nil, fmt.Errorf("must specify either 'keys' or 'pages' in config")
	}

	// Validate keys (old format)
	for _, k := range c.Keys {
		err := k.Validate()
		if err != nil {
			return nil, nil, err
		}

		if !slices.Contains(ret, k.Component) {
			ret = append(ret, k.Component)
		}
	}

	// Validate pages (new format)
	for pageName, keys := range c.Pages {
		if pageName == "" {
			return nil, nil, fmt.Errorf("page name cannot be empty")
		}
		for _, k := range keys {
			err := k.Validate()
			if err != nil {
				return nil, nil, fmt.Errorf("page %s: %w", pageName, err)
			}

			if !slices.Contains(ret, k.Component) {
				ret = append(ret, k.Component)
			}
		}
	}

	// Validate initial_page - required when using pages
	if len(c.Pages) > 0 {
		if c.InitialPage == "" {
			return nil, nil, fmt.Errorf("initial_page is required when using pages")
		}
		if _, ok := c.Pages[c.InitialPage]; !ok {
			return nil, nil, fmt.Errorf("initial_page '%s' not found in pages", c.InitialPage)
		}
	}

	// Validate dials
	for _, d := range c.Dials {
		err := d.Validate()
		if err != nil {
			return nil, nil, err
		}

		if !slices.Contains(ret, d.Component) {
			ret = append(ret, d.Component)
		}
	}

	return nil, ret, nil
}

// GetPageNames returns a sorted list of page names
func (c *Config) GetPageNames() []string {
	names := make([]string, 0, len(c.Pages))
	for name := range c.Pages {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// GetKeysForPage returns the keys for the given page, or the default keys if using old format
func (c *Config) GetKeysForPage(pageName string) ([]KeyConfig, error) {
	// Old format - return the keys directly
	if len(c.Keys) > 0 {
		if pageName != "" {
			return nil, fmt.Errorf("pages not supported in this config")
		}
		return c.Keys, nil
	}

	// New format - lookup the page
	keys, ok := c.Pages[pageName]
	if !ok {
		return nil, fmt.Errorf("page %s not found", pageName)
	}
	return keys, nil
}
