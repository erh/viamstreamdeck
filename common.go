package viamstreamdeck

import (
	"context"
	"fmt"
	"image/color"
	"strings"

	"golang.org/x/image/colornames"

	"go.uber.org/multierr"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/dh1tw/streamdeck"
)

func NewStreamDeck(ctx context.Context, name resource.Name, deps resource.Dependencies, sdConfig streamdeck.Config, conf *Config, logger logging.Logger) (resource.Resource, error) {

	_, _, err := conf.Validate("")
	if err != nil {
		return nil, err
	}

	sdc := &streamdeckComponent{
		name:   name,
		conf:   conf,
		logger: logger,
		deps:   deps,
		keys:   map[int]KeyConfig{},
	}

	sdc.sd, err = streamdeck.NewStreamDeck(sdConfig)
	if err != nil {
		return nil, err
	}

	err = sdc.updateKeys(conf.Keys)
	if err != nil {
		return nil, err
	}

	sdc.sd.SetBtnEventCb(func(s streamdeck.State, e streamdeck.Event) {
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

	name   resource.Name
	conf   *Config
	deps   resource.Dependencies
	logger logging.Logger

	sd   *streamdeck.StreamDeck
	keys map[int]KeyConfig
}

func findDep(deps resource.Dependencies, n string) (resource.Resource, bool) {
	for nn, r := range deps {
		if nn.ShortName() == n {
			return r, true
		}
	}
	return nil, false
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	result := ""

	for _, part := range parts {
		result += strings.ToUpper(string(part[0])) + part[1:]
	}

	return result
}

func (sdc *streamdeckComponent) updateKey(k KeyConfig) error {
	_, ok := findDep(sdc.deps, k.Component)
	if !ok {
		return fmt.Errorf("can't find component [%s]", k.Component)
	}

	if snakeToCamel(k.Method) != "DoCommand" {
		return fmt.Errorf("only support DoCommand now, not %s", k.Method)
	}

	if k.Image != "" {
		img, ok := assetImages[k.Image]
		if ok {
			if k.Text != "" {
				return sdc.sd.WriteTextOnImage(
					k.Key,
					img,
					[]streamdeck.TextLine{{Text: k.Text, PosX: 10, PosY: 30, FontSize: 20, FontColor: getColor(k.TextColor, "white")}},
				)
			}
			return sdc.sd.FillImage(k.Key, img)
		}
		return fmt.Errorf("unknown image [%s]", k.Image)
	}

	if k.Text != "" {
		tb := streamdeck.TextButton{
			Lines: []streamdeck.TextLine{
				{Text: k.Text, PosX: 10, PosY: 30, FontSize: 20, FontColor: getColor(k.TextColor, "white")},
			},
			BgColor: getColor(k.Color, "black"),
		}
		return sdc.sd.WriteText(k.Key, tb)
	}

	return fmt.Errorf("nothing to display for key %v", k)
}

func (sdc *streamdeckComponent) updateKeys(keys []KeyConfig) error {
	for _, k := range keys {
		err := sdc.updateKey(k)
		if err != nil {
			return err
		}
		sdc.keys[k.Key] = k
	}
	return nil
}

func (sdc *streamdeckComponent) handleKeyPress(ctx context.Context, s streamdeck.State, e streamdeck.Event, which int) error {
	k, ok := sdc.keys[which]
	if !ok {
		return fmt.Errorf("no key for %v", e)
	}

	r, ok := findDep(sdc.deps, k.Component)
	if !ok {
		return fmt.Errorf("no resource %s for %s", k.Component, e)
	}

	cmd := map[string]interface{}{}

	if len(k.Args) > 0 {
		cmd, ok = k.Args[0].(map[string]interface{})
		if !ok {
			return fmt.Errorf("args wrong for %v %v %T", e, k.Args[0], k.Args[0])
		}
	}

	res, err := r.DoCommand(ctx, cmd)
	if err != nil {
		return err
	}
	sdc.logger.Infof("event %v got result %v", e, res)
	return nil
}

func (sdc *streamdeckComponent) HandleEvent(ctx context.Context, s streamdeck.State, e streamdeck.Event) error {
	sdc.logger.Infof("got event %v", e)

	switch e.Kind {
	case streamdeck.EventKeyPush:
		return nil
	case streamdeck.EventKeyUnpush:
		return sdc.handleKeyPress(ctx, s, e, e.Which)
	}

	return fmt.Errorf("HandleEvent for %v not done", e)
}

func (sdc *streamdeckComponent) Name() resource.Name {
	return sdc.name
}

func (sdc *streamdeckComponent) Close(ctx context.Context) error {
	return multierr.Combine(sdc.sd.ClearAllBtns(), sdc.sd.Close())
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
