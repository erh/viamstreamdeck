package viamstreamdeck

import (
	"embed"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype/truetype"
)

//go:embed assets/*
var assetsFS embed.FS

var assetImages map[string]image.Image
var assetFonts map[string]*truetype.Font

func init() {
	var err error
	assetImages, err = loadImages()
	if err != nil {
		panic(err)
	}

	assetFonts, err = loadFonts()
	if err != nil {
		panic(err)
	}
}

func loadImages() (map[string]image.Image, error) {
	imageMap := make(map[string]image.Image)

	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	// Walk through all files in the embedded assets directory
	err := fs.WalkDir(assetsFS, "assets/images", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		f, err := assetsFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			return fmt.Errorf("failed to decode image %s: %w", path, err)
		}

		filename := filepath.Base(path)
		imageMap[filename] = img

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk assets directory: %w", err)
	}

	return imageMap, nil
}

func loadFonts() (map[string]*truetype.Font, error) {
	fontMap := make(map[string]*truetype.Font)

	supportedExts := map[string]bool{
		".ttf": true,
		".otf": true,
	}

	err := fs.WalkDir(assetsFS, "assets/fonts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		f, err := assetsFS.Open(path)
		if err != nil {
			return fmt.Errorf("failed to read font file %s: %w", path, err)
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return fmt.Errorf("failed to read font data %s: %w", path, err)
		}

		font, err := truetype.Parse(data)
		if err != nil {
			return fmt.Errorf("failed to parse font %s: %w", path, err)
		}

		filename := filepath.Base(path)
		fontMap[filename] = font

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk fonts directory: %w", err)
	}

	return fontMap, nil
}

// GetFont returns a font by filename, or nil if not found
func GetFont(filename string) *truetype.Font {
	return assetFonts[filename]
}

// LoadExternalFont loads a font from an absolute file path and adds it to assetFonts
func LoadExternalFont(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read font file %s: %w", path, err)
	}

	font, err := truetype.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse font %s: %w", path, err)
	}

	filename := filepath.Base(path)
	assetFonts[filename] = font
	return nil
}

// LoadExternalImage loads an image from an absolute file path and adds it to assetImages
func LoadExternalImage(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open image file %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode image %s: %w", path, err)
	}

	filename := filepath.Base(path)
	assetImages[filename] = img
	return nil
}

// loadFontsFromPath loads fonts from a file or directory
func loadFontsFromPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if !info.IsDir() {
		// Single file
		return LoadExternalFont(path)
	}

	// Directory - walk and load all supported font files
	supportedExts := map[string]bool{
		".ttf": true,
		".otf": true,
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !supportedExts[ext] {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		if err := LoadExternalFont(fullPath); err != nil {
			return err
		}
	}

	return nil
}

// loadImagesFromPath loads images from a file or directory
func loadImagesFromPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if !info.IsDir() {
		// Single file
		return LoadExternalImage(path)
	}

	// Directory - walk and load all supported image files
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !supportedExts[ext] {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		if err := LoadExternalImage(fullPath); err != nil {
			return err
		}
	}

	return nil
}

// LoadExternalAssets loads fonts and images from the AssetsConfig
// Supports both individual files and directories
func LoadExternalAssets(assetsConfig *AssetsConfig) error {
	if assetsConfig == nil {
		return nil
	}

	for _, fontPath := range assetsConfig.Fonts {
		if err := loadFontsFromPath(fontPath); err != nil {
			return err
		}
	}

	for _, imagePath := range assetsConfig.Images {
		if err := loadImagesFromPath(imagePath); err != nil {
			return err
		}
	}

	return nil
}
