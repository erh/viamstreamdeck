package viamstreamdeck

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/multierr"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"

	"github.com/dh1tw/streamdeck"

	"github.com/erh/vmodutils"

	_ "go.viam.com/rdk/components/arm"
	_ "go.viam.com/rdk/components/base"
	_ "go.viam.com/rdk/components/board"
	_ "go.viam.com/rdk/components/button"
	_ "go.viam.com/rdk/components/camera"
	_ "go.viam.com/rdk/components/generic"
	_ "go.viam.com/rdk/components/gripper"
	_ "go.viam.com/rdk/components/motor"
	_ "go.viam.com/rdk/components/movementsensor"
	_ "go.viam.com/rdk/components/sensor"
	_ "go.viam.com/rdk/services/generic"
	_ "go.viam.com/rdk/services/motion"
	_ "go.viam.com/rdk/services/vision"
)

func NewStreamDeck(ctx context.Context, name resource.Name, deps resource.Dependencies, ms *ModelSetup, conf *Config, logger logging.Logger) (resource.Resource, error) {

	_, _, err := conf.Validate("")
	if err != nil {
		return nil, err
	}

	sdc := &streamdeckComponent{
		name:   name,
		logger: logger,
		ms:     ms,
		deps:   deps,
		keys:   map[int]KeyConfig{},
	}

	sdc.sd, err = streamdeck.NewStreamDeck(ms.Conf)
	if err != nil && ms == ModelOriginal {
		// original vs original2 is confusing, try it
		ms = ModelOriginal2
		sdc.ms = ModelOriginal2
		sdc.sd, err = streamdeck.NewStreamDeck(ms.Conf)
	}

	if err != nil {
		return nil, err
	}

	err = sdc.updateBrightness(conf.Brightness)
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

func (sdc *streamdeckComponent) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}

	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	sdc.deps = deps

	err = sdc.updateBrightness(newConf.Brightness)
	if err != nil {
		return err
	}

	return sdc.updateKeys(newConf.Keys)
}

type streamdeckComponent struct {
	name   resource.Name
	logger logging.Logger
	ms     *ModelSetup

	sd *streamdeck.StreamDeck

	configLock sync.Mutex
	deps       resource.Dependencies
	keys       map[int]KeyConfig
}

func (sdc *streamdeckComponent) updateBrightness(level int) error {
	if level <= 0 {
		return nil
	}
	if level > 100 {
		level = 100
	}
	return sdc.sd.SetBrightness(uint16(level))
}

func (sdc *streamdeckComponent) updateKey(k KeyConfig) error {
	_, ok := vmodutils.FindDep(sdc.deps, k.Component)
	if !ok {
		img, ok := assetImages["x.jpg"]
		if !ok {
			return fmt.Errorf("can't find dependency %s nore, the x image :(", k.Component)
		}

		return sdc.sd.WriteTextOnImage(
			k.Key,
			img,
			[]streamdeck.TextLine{{Text: k.Component, PosX: 10, PosY: 30, FontSize: 20, FontColor: getColor("black", "black")}},
		)
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
					sdc.ms.SimpleText(k.Text, k.TextColor),
				)
			}
			return sdc.sd.FillImage(k.Key, img)
		}
		return fmt.Errorf("unknown image [%s]", k.Image)
	}

	if k.Text != "" {
		return sdc.sd.WriteText(k.Key, sdc.ms.SimpleTextButton(k.Text, k.Color, k.TextColor))
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
	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	k, ok := sdc.keys[which]
	if !ok {
		return fmt.Errorf("no key for %v", e)
	}

	r, ok := vmodutils.FindDep(sdc.deps, k.Component)
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
