package viamstreamdeck

import (
	"embed"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed assets/*
var assetsFS embed.FS

var assetImages map[string]image.Image

func init() {
	var err error
	assetImages, err = loadImages()
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
	err := fs.WalkDir(assetsFS, "assets", func(path string, d fs.DirEntry, err error) error {
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
