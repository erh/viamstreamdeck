package viamstreamdeck

import (
	"context"
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/dh1tw/streamdeck"
)

func NewStreamDeck(ctx context.Context, name resource.Name, deps resource.Dependencies, sdConfig streamdeck.Config, conf *Config, logger logging.Logger) (resource.Resource, error) {

	_, _, err := conf.Validate("")
	if err != nil {
		return nil, err
	}

	sd, err := streamdeck.NewStreamDeck(sdConfig)
	if err != nil {
		return nil, err
	}

	sdc := &streamdeckComponent{
		name: name,
		conf: conf,
		sd:   sd,
	}

	err = sdc.updateKeys(conf.Keys)
	if err != nil {
		return nil, err
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

func (sdc *streamdeckComponent) updateKeys(keys []KeyConfig) error {
	for _, k := range keys {
		tb := streamdeck.TextButton{
			Lines: []streamdeck.TextLine{
				{Text: k.Text, PosX: 10, PosY: 30, FontSize: 20, FontColor: getColor(k.TextColor, "white")},
			},
			BgColor: getColor(k.Color, "black"),
		}
		fmt.Printf("%v\n", tb)
		err := sdc.sd.WriteText(k.Key, tb)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sdc *streamdeckComponent) HandleEvent(ctx context.Context, s streamdeck.State, e streamdeck.Event) error {
	return fmt.Errorf("HandleEvent finish me")
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

func getColor(want, def string) color.Color {
	c, ok := colornames.Map[want]
	if ok {
		return c
	}

	c, ok = colornames.Map[def]
	if ok {
		return c
	}

	panic(fmt.Errorf("default color didn't work [%s]", def))
}
