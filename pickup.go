package viamstreamdeck

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dh1tw/streamdeck"
	"go.uber.org/multierr"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/components/gripper"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/referenceframe"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot/framesystem"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/services/vision"
	"go.viam.com/rdk/spatialmath"
	viz "go.viam.com/rdk/vision"

	"github.com/erh/vmodutils/touch"
)

const STATE_PREP = 0
const STATE_WAITING_FOR_CHOICE = 1

var PickupModel = NamespaceFamily.WithModel("pickup")

func init() {
	resource.RegisterService(generic.API, PickupModel, resource.Registration[resource.Resource, *PickupConfig]{
		Constructor: func(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {

			newConf, err := resource.NativeConfig[*PickupConfig](conf)
			if err != nil {
				return nil, err
			}

			return NewPickup(ctx, conf.ResourceName(), deps, newConf, logger)
		}})

}

type PickupConfig struct {
	Arm        string
	Gripper    string
	Finder     string
	Motion     string
	WatchPose  string `json:"watch_pose"`
	Brightness uint16
}

func (pc *PickupConfig) Validate(_ string) ([]string, []string, error) {
	deps := []string{}
	if pc.Arm == "" {
		return nil, nil, fmt.Errorf("need an arm")
	}
	deps = append(deps, pc.Arm)

	if pc.Gripper == "" {
		return nil, nil, fmt.Errorf("need a gripper")
	}
	deps = append(deps, pc.Gripper)

	if pc.Finder == "" {
		return nil, nil, fmt.Errorf("need a finder (vision service)")
	}
	deps = append(deps, pc.Finder)

	if pc.WatchPose == "" {
		return nil, nil, fmt.Errorf("need a watch_pose (arm saver)")
	}
	deps = append(deps, pc.WatchPose)

	if pc.Motion == "" {
		return nil, nil, fmt.Errorf("need a motion")
	}
	deps = append(deps, motion.Named(pc.Motion).String())

	return deps, nil, nil
}

func NewPickup(ctx context.Context, name resource.Name, deps resource.Dependencies, conf *PickupConfig, logger logging.Logger) (*Pickup, error) {
	ms := FindAttachedStreamDeck()
	if ms == nil {
		return nil, fmt.Errorf("no streamdeck found")
	}

	p := &Pickup{
		name:   name,
		logger: logger,
		conf:   conf,
		ms:     ms,
	}

	var err error

	p.arm, err = arm.FromDependencies(deps, conf.Arm)
	if err != nil {
		return nil, err
	}

	p.gripper, err = gripper.FromDependencies(deps, conf.Gripper)
	if err != nil {
		return nil, err
	}

	p.motion, err = motion.FromDependencies(deps, conf.Motion)
	if err != nil {
		return nil, err
	}

	p.rfs, err = framesystem.FromDependencies(deps)
	if err != nil {
		return nil, err
	}

	p.watchPose, err = toggleswitch.FromDependencies(deps, conf.WatchPose)
	if err != nil {
		return nil, err
	}

	p.finder, err = vision.FromDependencies(deps, conf.Finder)
	if err != nil {
		return nil, err
	}
	prop, err := p.finder.GetProperties(ctx, nil)
	if err != nil {
		return nil, err
	}
	if !prop.ObjectPCDsSupported {
		return nil, fmt.Errorf("ObjectPCDsSupported not supported by %v", conf.Finder)
	}

	// keep this last so we don't have to close it
	p.sd, err = streamdeck.NewStreamDeckWithConfig(&ms.Conf, "")
	if err != nil {
		return nil, err
	}

	p.sd.SetBtnEventCb(func(s streamdeck.State, e streamdeck.Event) {
		logger.Debugf("got event %v", e)
		err := p.HandleEvent(context.Background(), s, e)
		if err != nil {
			logger.Errorf("event handler failed for event %v: %v", e, err)
		}
	})

	err = p.Prep(ctx)
	if err != nil {
		return nil, multierr.Combine(err, p.sd.Close())
	}

	return p, nil

}

type Pickup struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	conf   *PickupConfig
	ms     *ModelSetup

	sd *streamdeck.StreamDeck

	arm       arm.Arm
	gripper   gripper.Gripper
	motion    motion.Service
	rfs       framesystem.Service
	watchPose toggleswitch.Switch
	finder    vision.Service

	stateLock    sync.Mutex
	currentState int
	lastObjects  []*viz.Object
}

func (p *Pickup) Name() resource.Name {
	return p.name
}

func (p *Pickup) Close(ctx context.Context) error {
	return multierr.Combine(p.sd.ClearAllBtns(), p.sd.Close())
}

func (p *Pickup) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (p *Pickup) Prep(ctx context.Context) error {
	if p.conf.Brightness > 0 {
		err := p.sd.SetBrightness(p.conf.Brightness)
		if err != nil {
			return err
		}
	}

	p.stateLock.Lock()
	defer p.stateLock.Unlock()

	p.currentState = STATE_PREP

	err := p.sd.ClearAllBtns()
	if err != nil {
		return err
	}

	return multierr.Combine(
		p.sd.WriteText(0, p.ms.SimpleTextButton("move to watch", "", "", nil)),
		p.sd.WriteText(1, p.ms.SimpleTextButton("move and image", "", "", nil)),
	)
}

func (p *Pickup) imageAndChangeButtonsInLock(ctx context.Context) error {
	err := p.watchPose.SetPosition(ctx, 2, nil)
	if err != nil {
		return err
	}

	time.Sleep(time.Millisecond * 500)

	objs, err := p.finder.GetObjectPointClouds(ctx, "", nil)
	if err != nil {
		return err
	}

	err = p.sd.ClearAllBtns()
	if err != nil {
		return err
	}

	for idx, o := range objs {
		if (idx + 1) >= p.ms.Conf.NumButtons() {
			p.logger.Info("too many objects %d %v", idx, o)
			continue
		}
		img := touch.PCToImage(o)
		err = p.sd.FillImage(idx, img)
		if err != nil {
			return err
		}
	}

	p.currentState = STATE_WAITING_FOR_CHOICE
	p.lastObjects = objs

	return p.sd.WriteText(p.ms.Conf.NumButtons()-1, p.ms.SimpleTextButton("image again", "", "", nil))
}

func (p *Pickup) HandleEvent(ctx context.Context, s streamdeck.State, e streamdeck.Event) error {
	if e.Kind != streamdeck.EventKeyReleased {
		p.logger.Debugf("ignoring %v %v", s, e)
		return nil
	}

	p.stateLock.Lock()
	defer p.stateLock.Unlock()

	switch p.currentState {
	case STATE_PREP:
		switch e.Which {
		case 0:
			return p.watchPose.SetPosition(ctx, 2, nil)
		case 1:
			return p.imageAndChangeButtonsInLock(ctx)
		default:
			p.logger.Infof("in prep state, button %v does nothing", e.Which)
			return nil
		}
	case STATE_WAITING_FOR_CHOICE:
		if e.Which == p.ms.Conf.NumButtons()-1 {
			return p.imageAndChangeButtonsInLock(ctx)
		}

		return p.pickupInLock(ctx, e.Which)
	default:
		return fmt.Errorf("unknown state %v", p.currentState)
	}
}

func (p *Pickup) pickupInLock(ctx context.Context, which int) error {

	if which >= len(p.lastObjects) {
		return fmt.Errorf("no object for button %v", which)
	}

	md := p.lastObjects[which].MetaData()
	theSpot := md.Center()
	theSpot.Z = md.MaxZ

	theSpot = touch.GetApproachPoint(theSpot, 0, &spatialmath.OrientationVectorDegrees{OZ: -1})

	current, err := p.rfs.GetPose(ctx, p.conf.Gripper, "world", nil, nil)
	if err != nil {
		return err
	}

	goalPose := referenceframe.NewPoseInFrame("world",
		spatialmath.NewPose(
			theSpot,
			&spatialmath.OrientationVectorDegrees{OZ: -1, Theta: current.Pose().Orientation().OrientationVectorDegrees().Theta},
		),
	)

	p.logger.Infof("want to go to %v", goalPose)

	err = p.gripper.Open(ctx, nil)
	if err != nil {
		return fmt.Errorf("can't open gripper %w", err)
	}

	_, err = p.motion.Move(
		ctx,
		motion.MoveReq{
			ComponentName: p.conf.Gripper,
			Destination:   goalPose,
		},
	)
	if err != nil {
		return err
	}

	return fmt.Errorf("finish me")
}
