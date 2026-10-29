package viamstreamdeck

import (
	"context"
	"testing"

	"github.com/dh1tw/streamdeck"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/testutils/inject"
	"go.viam.com/test"
)

// dialTestComponent wires one dial (index 0) to the named dependency and command.
func dialTestComponent(t *testing.T, command string, deps ...resource.Resource) *streamdeckComponent {
	d := resource.Dependencies{}
	for _, r := range deps {
		d[r.Name()] = r
	}
	return &streamdeckComponent{
		logger: logging.NewTestLogger(t),
		deps:   d,
		conf: &Config{
			Dials: []DialConfig{{Dial: 0, Component: "target", Command: command}},
		},
	}
}

func dialState(pos int) streamdeck.State {
	return streamdeck.State{DialPos: []int{pos, 0, 0, 0}}
}

func TestHandleDialTurnCustomCommand(t *testing.T) {
	ctx := context.Background()

	// Dial configs name the DoCommand key to send (the dial-control service uses
	// dial_move_x and friends), so any command must reach DoCommand unchanged.
	for _, command := range []string{"dial_move_x", "DoCommand", "anything"} {
		var got map[string]interface{}
		svc := inject.NewGenericService("target")
		svc.DoFunc = func(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
			got = cmd
			return map[string]interface{}{}, nil
		}

		err := dialTestComponent(t, command, svc).handleDialTurn(ctx, dialState(7), 0)
		test.That(t, err, test.ShouldBeNil)
		test.That(t, got, test.ShouldResemble, map[string]interface{}{command: float64(7)})
	}
}

func TestHandleDialTurnSetPositionOnSwitch(t *testing.T) {
	ctx := context.Background()

	var position uint32
	var doCalled bool
	sw := inject.NewSwitch("target")
	sw.SetPositionFunc = func(ctx context.Context, p uint32, extra map[string]interface{}) error {
		position = p
		return nil
	}
	sw.DoFunc = func(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
		doCalled = true
		return nil, nil
	}

	err := dialTestComponent(t, "SetPosition", sw).handleDialTurn(ctx, dialState(3), 0)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, position, test.ShouldEqual, 3)
	test.That(t, doCalled, test.ShouldBeFalse)
}

func TestHandleDialTurnSetPositionOnNonSwitch(t *testing.T) {
	ctx := context.Background()

	// Before dials could address switches, "SetPosition" was just another DoCommand
	// key, so a resource that is not a switch still gets the DoCommand call.
	var got map[string]interface{}
	svc := inject.NewGenericService("target")
	svc.DoFunc = func(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
		got = cmd
		return map[string]interface{}{}, nil
	}

	err := dialTestComponent(t, "SetPosition", svc).handleDialTurn(ctx, dialState(5), 0)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, got, test.ShouldResemble, map[string]interface{}{"SetPosition": float64(5)})
}

func TestHandleDialTurnMissingComponent(t *testing.T) {
	err := dialTestComponent(t, "dial_move_x").handleDialTurn(context.Background(), dialState(1), 0)
	test.That(t, err, test.ShouldNotBeNil)
}
