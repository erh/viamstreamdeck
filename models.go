package viamstreamdeck

import (
	"context"
	"fmt"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"

	"github.com/bearsh/hid"
	"github.com/dh1tw/streamdeck"
)

var NamespaceFamily = resource.ModelNamespace("erh").WithFamily("viam-streamdeck")

type ModelSetup struct {
	Model resource.Model
	Conf  streamdeck.Config
}

var ModelPlus = &ModelSetup{NamespaceFamily.WithModel("streamdeck-plus"), streamdeck.Plus}
var ModelOriginal = &ModelSetup{NamespaceFamily.WithModel("streamdeck-original"), streamdeck.Original}
var ModelOriginal2 = &ModelSetup{NamespaceFamily.WithModel("streamdeck-original2"), streamdeck.Original2}

var Models = []*ModelSetup{
	ModelPlus,
	ModelOriginal,
	ModelOriginal2,
}

var ModelAny = NamespaceFamily.WithModel("streamdeck-any")

func init() {
	for _, ms := range Models {
		resource.RegisterService(generic.API, ms.Model, resource.Registration[resource.Resource, *Config]{
			Constructor: func(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {

				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}

				return NewStreamDeck(ctx, conf.ResourceName(), deps, ms, newConf, logger)
			},
		})
	}

	resource.RegisterService(generic.API, ModelAny, resource.Registration[resource.Resource, *Config]{
		Constructor: func(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
			ms := FindAttachedStreamDeck()
			if ms == nil {
				return nil, fmt.Errorf("no streamdeck found")
			}

			newConf, err := resource.NativeConfig[*Config](conf)
			if err != nil {
				return nil, err
			}

			return NewStreamDeck(ctx, conf.ResourceName(), deps, ms, newConf, logger)
		},
	})

}

func FindAttachedStreamDeck() *ModelSetup {
	for _, ms := range Models {
		devices := hid.Enumerate(streamdeck.VendorID, ms.Conf.ProductID)
		if len(devices) > 0 {
			return ms
		}
	}
	return nil
}
