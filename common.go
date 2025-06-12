package viamstreamdeck

import (
	"context"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/dh1tw/streamdeck"
)

func NewStreamDeck(ctx context.Context, name resource.Name, deps resource.Dependencies, sdConfig streamdeck.Config, conf *Config, logger logging.Logger) (resource.Resource, error) {
	sd, err := streamdeck.NewStreamDeck(sdConfig)
	if err != nil {
		return nil, err
	}

	sdc := &streamdeckComponent{
		name: name,
		conf: conf,
		sd:   sd,
	}

	sd.SetBtnEventCb(func(s streamdeck.State, e streamdeck.Event) {
		logger.Infof("got event %v", e)
		err := sdc.HandleEvent(context.Background(), s, e)
		if err != nil {
			logger.Errorf("event handler failed for event %v: %v", e, err)
		}
	})

	return sdc, nil
}

type streamdeckComponent struct {
	resource.AlwaysRebuild

	name resource.Name
	conf *Config
	sd   *streamdeck.StreamDeck
}

func (sdc *streamdeckComponent) HandleEvent(ctx context.Context, s streamdeck.State, e streamdeck.Event) error {
	panic(1)
}

func (sdc *streamdeckComponent) Name() resource.Name {
	return sdc.name
}

func (sdc *streamdeckComponent) Close(ctx context.Context) error {
	return sdc.sd.Close()
}

func (sdc *streamdeckComponent) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
