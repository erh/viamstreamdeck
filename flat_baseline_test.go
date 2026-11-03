package viamstreamdeck

import (
	"context"
	"testing"

	"go.viam.com/test"
)

// These tests pin the behavior of a flat `keys` config.

func flatConfig() *Config {
	return &Config{
		Keys: []KeyConfig{textKey(0, "zero"), textKey(1, "one")},
	}
}

func TestFlatLoadDraws(t *testing.T) {
	_, fd := newTestComponent(t, flatConfig(), "thing")
	test.That(t, fd.button(0), test.ShouldEqual, "text:zero")
	test.That(t, fd.button(1), test.ShouldEqual, "text:one")
	test.That(t, fd.button(2), test.ShouldEqual, "")
}

func TestFlatUpdateExistingKey(t *testing.T) {
	sdc, fd := newTestComponent(t, flatConfig(), "thing")

	res, err := update(sdc, keyUpdate("1", map[string]interface{}{"text": "uno"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, res, test.ShouldResemble, map[string]interface{}{"keys": []int{1}})

	test.That(t, fd.button(1), test.ShouldEqual, "text:uno")
	test.That(t, fd.button(0), test.ShouldEqual, "text:zero")
	test.That(t, sdc.conf.Keys[1].Text, test.ShouldEqual, "uno")
	test.That(t, len(sdc.conf.Keys), test.ShouldEqual, 2)
}

func TestFlatUpdateNewKey(t *testing.T) {
	sdc, fd := newTestComponent(t, flatConfig(), "thing")

	_, err := update(sdc, keyUpdate("5", map[string]interface{}{
		"text": "five", "component": "thing", "method": "do_command",
	}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(5), test.ShouldEqual, "text:five")
	test.That(t, len(sdc.conf.Keys), test.ShouldEqual, 3)
}

func TestFlatUpdateSurvivesRedraw(t *testing.T) {
	// stateChecker redraws every key from the config once a second.
	sdc, fd := newTestComponent(t, flatConfig(), "thing")
	_, err := update(sdc, keyUpdate("0", map[string]interface{}{"text": "changed"}))
	test.That(t, err, test.ShouldBeNil)

	test.That(t, sdc.reconfigure(context.Background(), sdc.deps, sdc.conf), test.ShouldBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "text:changed")
}

func TestFlatBrightnessAndDials(t *testing.T) {
	sdc, fd := newTestComponent(t, flatConfig(), "thing")

	res, err := update(sdc, map[string]interface{}{"brightness": 40})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, res, test.ShouldResemble, map[string]interface{}{"brightness": 40})
	test.That(t, fd.brightness, test.ShouldEqual, uint16(40))

	res, err = update(sdc, map[string]interface{}{"dials": map[string]interface{}{
		"2": map[string]interface{}{"component": "thing", "command": "DoCommand"},
	}})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, res, test.ShouldResemble, map[string]interface{}{"dials": []int{2}})
	test.That(t, sdc.conf.Dials, test.ShouldResemble, []DialConfig{{Dial: 2, Component: "thing", Command: "DoCommand"}})
}

func TestFlatMissingComponentDrawsCrossedCircle(t *testing.T) {
	conf := &Config{Keys: []KeyConfig{textKey(0, "zero")}}
	conf.Keys[0].Component = "gone"
	_, fd := newTestComponent(t, conf) // no deps at all
	test.That(t, fd.button(0), test.ShouldEqual, "image+text:gone")
}

func TestFlatRejections(t *testing.T) {
	sdc, _ := newTestComponent(t, flatConfig(), "thing")

	_, err := update(sdc, map[string]interface{}{})
	mustErr(t, err, "no valid updates provided")

	_, err = update(sdc, keyUpdate("0", map[string]interface{}{"method": "get_readings"}))
	mustErr(t, err, "only support DoCommand and SetPosition")

	_, err = update(sdc, keyUpdate("0", map[string]interface{}{"image": "nope.png"}))
	mustErr(t, err, "unknown image")

	_, err = update(sdc, keyUpdate("x", map[string]interface{}{"text": "hi"}))
	test.That(t, err, test.ShouldNotBeNil)

	_, err = sdc.DoCommand(context.Background(), map[string]interface{}{"update_display": "nope"})
	mustErr(t, err, "must be an object")

	_, err = sdc.DoCommand(context.Background(), map[string]interface{}{"bogus": 1})
	mustErr(t, err, "unknown command")
}

func TestFlatSetPageIsRejected(t *testing.T) {
	sdc, _ := newTestComponent(t, flatConfig(), "thing")
	_, err := sdc.DoCommand(context.Background(), map[string]interface{}{"set_page": "p"})
	mustErr(t, err, "pages not supported")
}

// A reference is read with Sscanf, which takes the leading integer of a
// malformed one, so "1abc" addresses key 1.
func TestFlatMalformedKeyReferenceReadAsLeadingInteger(t *testing.T) {
	sdc, fd := newTestComponent(t, flatConfig(), "thing")
	_, err := update(sdc, keyUpdate("1abc", map[string]interface{}{"text": "sneaky"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(1), test.ShouldEqual, "text:sneaky")
}
