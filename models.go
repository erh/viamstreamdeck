package viamstreamdeck

import (
	"context"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"

	"github.com/dh1tw/streamdeck"
)

var NamespaceFamily = resource.ModelNamespace("erh").WithFamily("viam-streamdeck")

type ModelSetup struct {
	Model resource.Model
	Conf  streamdeck.Config
}

var Models = []ModelSetup{
	{NamespaceFamily.WithModel("streamdeck-plus"), streamdeck.Plus},
	{NamespaceFamily.WithModel("streamdeck-original"), streamdeck.Original},
	{NamespaceFamily.WithModel("streamdeck-original2"), streamdeck.Original2},
}

func init() {
	for _, ms := range Models {
		resource.RegisterService(generic.API, ms.Model, resource.Registration[resource.Resource, *Config]{
			Constructor: func(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {

				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}

				return NewStreamDeck(ctx, conf.ResourceName(), deps, ms.Conf, newConf, logger)
			},
		})
	}
}
