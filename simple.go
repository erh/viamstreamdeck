package viamstreamdeck

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/multierr"

	"go.viam.com/rdk/components/switch"
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

	if ms == nil {
		ms = FindAttachedStreamDeck()
		if ms == nil {
			return nil, fmt.Errorf("no streamdeck found")
		}
	}

	sdc := &streamdeckComponent{
		name:   name,
		logger: logger,
		ms:     ms,
		conf:   conf,
		deps:   deps,
		keys:   map[int]KeyConfig{},
	}

	sdc.sd, err = streamdeck.NewStreamDeckWithConfig(&ms.Conf, "")
	if err != nil && ms == ModelOriginal {
		// original vs original2 is confusing, try it
		ms = ModelOriginal2
		sdc.ms = ModelOriginal2
		sdc.sd, err = streamdeck.NewStreamDeckWithConfig(&ms.Conf, "")
	}

	if err != nil {
		return nil, err
	}

	err = sdc.updateBrightness(conf.Brightness)
	if err != nil {
		return nil, err
	}

	err = sdc.updateKeys(ctx, conf.Keys)
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

	go sdc.stateChecker()

	return sdc, nil
}

func (sdc *streamdeckComponent) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}

	return sdc.reconfigure(ctx, deps, newConf)
}

func (sdc *streamdeckComponent) reconfigure(ctx context.Context, deps resource.Dependencies, newConf *Config) error {
	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	sdc.deps = deps
	sdc.conf = newConf

	err := sdc.updateBrightness(newConf.Brightness)
	if err != nil {
		return err
	}

	return sdc.updateKeys(ctx, newConf.Keys)
}

type streamdeckComponent struct {
	name   resource.Name
	logger logging.Logger
	ms     *ModelSetup

	sd *streamdeck.StreamDeck

	configLock sync.Mutex
	deps       resource.Dependencies
	conf       *Config
	keys       map[int]KeyConfig

	closed atomic.Int32
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

func (sdc *streamdeckComponent) updateKey(ctx context.Context, k KeyConfig) error {
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

	if snakeToCamel(k.Method) != "DoCommand" && snakeToCamel(k.Method) != "SetPosition" {
		return fmt.Errorf("only support DoCommand and SetPosition now, not %s", k.Method)
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

	if k.Text == "" && snakeToCamel(k.Method) == "SetPosition" {
		s, err := sdc.findSwitch(ctx, k.Component)
		if err != nil {
			return err
		}
		_, names, err := s.GetNumberOfPositions(ctx, nil)
		if err != nil {
			return err
		}

		n, err := sdc.findSwitchArg(k)
		if err != nil {
			return err
		}

		if n < 0 || int(n) >= len(names) {
			return fmt.Errorf("invalid position %d", n)
		}

		if k.Color == "" && k.TextColor == "" {
			pos, err := s.GetPosition(ctx, nil)
			if err != nil {
				return err
			}

			if pos == n {
				k.TextColor = "black"
				k.Color = "white"
			} else {
				k.TextColor = "white"
				k.Color = "black"
			}
		}

		k.Text = names[n]
	}

	if k.Text != "" {
		return sdc.sd.WriteText(k.Key, sdc.ms.SimpleTextButton(k.Text, k.Color, k.TextColor))
	}

	return fmt.Errorf("nothing to display for key %v", k)
}

func (sdc *streamdeckComponent) updateKeys(ctx context.Context, keys []KeyConfig) error {
	for _, k := range keys {
		err := sdc.updateKey(ctx, k)
		if err != nil {
			return err
		}
		sdc.keys[k.Key] = k
	}
	return nil
}

func (sdc *streamdeckComponent) findSwitchArg(k KeyConfig) (uint32, error) {
	if len(k.Args) == 1 {
		switch v := k.Args[0].(type) {
		case int:
			return uint32(v), nil
		case float64:
			return uint32(v), nil
		case int32:
			return uint32(v), nil
		}
	}
	return 0, fmt.Errorf("need 1 number arg, got: %v", k.Args)
}

func (sdc *streamdeckComponent) findSwitch(ctx context.Context, name string) (toggleswitch.Switch, error) {
	r, ok := vmodutils.FindDep(sdc.deps, name)
	if !ok {
		return nil, fmt.Errorf("no resource %s", name)
	}

	s, ok := r.(toggleswitch.Switch)
	if !ok {
		return nil, fmt.Errorf("%s is a %T not switch", name, r)
	}

	return s, nil

}

func (sdc *streamdeckComponent) getKeyConfig(which int) (*KeyConfig, error) {
	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	k, ok := sdc.keys[which]
	if !ok {
		return nil, fmt.Errorf("no key for %v", which)
	}
	return &k, nil
}

func (sdc *streamdeckComponent) getResourceAndCommandForKey(which int, e streamdeck.Event) (resource.Resource, map[string]interface{}, error) {
	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	k, ok := sdc.keys[which]
	if !ok {
		return nil, nil, fmt.Errorf("no key for %v", e)
	}

	r, ok := vmodutils.FindDep(sdc.deps, k.Component)
	if !ok {
		return nil, nil, fmt.Errorf("no resource %s for %s", k.Component, e)
	}

	cmd := map[string]interface{}{}

	if len(k.Args) > 0 {
		cmd, ok = k.Args[0].(map[string]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("args wrong for %v %v %T", e, k.Args[0], k.Args[0])
		}
	}

	return r, cmd, nil
}

func (sdc *streamdeckComponent) getResourceAndCommandForDial(which int) (resource.Resource, string, error) {
	sdc.configLock.Lock()
	defer sdc.configLock.Unlock()

	for _, dc := range sdc.conf.Dials {
		if which != dc.Dial {
			continue
		}

		r, ok := vmodutils.FindDep(sdc.deps, dc.Component)
		if !ok {
			return nil, "", fmt.Errorf("no resource %s for %s", dc.Component)
		}

		return r, dc.Command, nil
	}

	return nil, "", fmt.Errorf("no config for dial %d", which)
}

func (sdc *streamdeckComponent) handleKeyPress(ctx context.Context, s streamdeck.State, e streamdeck.Event, which int) error {
	k, err := sdc.getKeyConfig(which)
	if err != nil {
		return err
	}

	if k.snakeMethod() == "DoCommand" {
		r, cmd, err := sdc.getResourceAndCommandForKey(which, e)
		if err != nil {
			return err
		}

		res, err := r.DoCommand(ctx, cmd)
		if err != nil {
			return err
		}
		sdc.logger.Infof("event %v got result %v", e, res)
		return nil
	} else if k.snakeMethod() == "SetPosition" {
		s, err := sdc.findSwitch(ctx, k.Component)
		if err != nil {
			return err
		}

		n, err := sdc.findSwitchArg(*k)
		if err != nil {
			return err
		}

		return s.SetPosition(ctx, n, nil)

	} else {
		return fmt.Errorf("can't handle command %v", k.snakeMethod())
	}
}

func (sdc *streamdeckComponent) handleDialTurn(ctx context.Context, s streamdeck.State, which int) error {
	sdc.logger.Infof("handleDialTurn called which: %v state: %v", which, s.DialPos[which])
	r, c, err := sdc.getResourceAndCommandForDial(which)
	if err != nil {
		return err
	}
	res, err := r.DoCommand(ctx, map[string]interface{}{c: float64(s.DialPos[which])})
	if err != nil {
		return err
	}
	sdc.logger.Infof("res: %v", res)
	return nil
}

func (sdc *streamdeckComponent) HandleEvent(ctx context.Context, s streamdeck.State, e streamdeck.Event) error {
	sdc.logger.Infof("got event %v", e)

	switch e.Kind {
	case streamdeck.EventKeyPressed:
		return nil
	case streamdeck.EventKeyReleased:
		return sdc.handleKeyPress(ctx, s, e, e.Which)
	case streamdeck.EventDialTurn:
		return sdc.handleDialTurn(ctx, s, e.Which)
	}

	return fmt.Errorf("HandleEvent for %v not done", e)
}

func (sdc *streamdeckComponent) Name() resource.Name {
	return sdc.name
}

func (sdc *streamdeckComponent) stateChecker() {
	for sdc.closed.Load() == 0 {

		err := sdc.reconfigure(context.Background(), sdc.deps, sdc.conf)
		if err != nil {
			sdc.logger.Errorf("can't reconfigure: %v", err)
		}

		time.Sleep(time.Second)
	}
}

func (sdc *streamdeckComponent) Close(ctx context.Context) error {
	sdc.closed.Store(1)
	return multierr.Combine(sdc.sd.ClearAllBtns(), sdc.sd.Close())
}

func (sdc *streamdeckComponent) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}
