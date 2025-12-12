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
	Brightness int
	Keys       []KeyConfig
	Dials      []DialConfig
}

func (c *Config) Validate(p string) ([]string, []string, error) {
	ret := []string{}

	for _, k := range c.Keys {
		err := k.Validate()
		if err != nil {
			return nil, nil, err
		}

		if !slices.Contains(ret, k.Component) {
			ret = append(ret, k.Component)
		}

	}

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
