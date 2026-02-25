package viamstreamdeck

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"go.viam.com/rdk/logging"
)

type KeyConfig struct {
	Key int

	Text      string
	TextColor string  `json:"text_color"`
	TextFont  *string `json:"text_font,omitempty"`

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
		return fmt.Errorf("need a method")
	}

	// Validate font exists (if specified)
	if kc.TextFont != nil {
		if _, ok := assetFonts[*kc.TextFont]; !ok {
			availableFonts := []string{}
			for fontName := range assetFonts {
				availableFonts = append(availableFonts, fontName)
			}
			sort.Strings(availableFonts)
			return fmt.Errorf("unknown font %s. Available fonts: %s", *kc.TextFont, strings.Join(availableFonts, ", "))
		}
	}

	// Validate image exists (if specified)
	if kc.Image != "" {
		if _, ok := assetImages[kc.Image]; !ok {
			availableImages := []string{}
			for imageName := range assetImages {
				availableImages = append(availableImages, imageName)
			}
			sort.Strings(availableImages)
			return fmt.Errorf("unknown image %s. Available images: %s", kc.Image, strings.Join(availableImages, ", "))
		}
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

type AssetsConfig struct {
	Fonts  []string `json:"fonts,omitempty"`
	Images []string `json:"images,omitempty"`
}

type Config struct {
	Brightness int
	Keys       []KeyConfig
	Dials      []DialConfig
	Assets     *AssetsConfig `json:"assets,omitempty"`
}

type UpdateDisplayCommand struct {
	Brightness *int                              `mapstructure:"brightness"`
	Keys       map[string]map[string]interface{} `mapstructure:"keys"`
	Dials      map[string]map[string]interface{} `mapstructure:"dials"`
}

func (c *Config) Validate(p string) ([]string, []string, error) {
	// Create logger for validation
	logger := logging.NewLogger("viamstreamdeck-config")

	// Load external assets FIRST before validating
	if c.Assets != nil {
		err := LoadExternalAssets(c.Assets)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load external assets: %w", err)
		}

		// Debug log all loaded fonts
		fontNames := []string{}
		for fontName := range assetFonts {
			fontNames = append(fontNames, fontName)
		}
		sort.Strings(fontNames)
		logger.Debugf("Loaded fonts: %s", strings.Join(fontNames, ", "))

		// Debug log all loaded images
		imageNames := []string{}
		for imageName := range assetImages {
			imageNames = append(imageNames, imageName)
		}
		sort.Strings(imageNames)
		logger.Debugf("Loaded images: %s", strings.Join(imageNames, ", "))
	}

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
