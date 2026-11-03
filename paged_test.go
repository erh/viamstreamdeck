package viamstreamdeck

import (
	"context"
	"testing"

	"go.viam.com/test"
)

func pagedConfig() *Config {
	home := []KeyConfig{textKey(0, "home0"), textKey(1, "home1")}
	home[1].Name = "status"
	other := []KeyConfig{textKey(0, "other0"), textKey(2, "other2")}
	other[0].Name = "status"
	other[1].Name = "extra"
	return &Config{
		Pages:       map[string][]KeyConfig{"home": home, "other": other},
		InitialPage: "home",
	}
}

func setPage(t *testing.T, sdc *streamdeckComponent, name string) {
	t.Helper()
	_, err := sdc.DoCommand(context.Background(), map[string]interface{}{"set_page": name})
	test.That(t, err, test.ShouldBeNil)
}

func TestPagedUpdateEditsThePageNotTheFlatKeys(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, keyUpdate("0", map[string]interface{}{"text": "changed"}))
	test.That(t, err, test.ShouldBeNil)

	test.That(t, fd.button(0), test.ShouldEqual, "text:changed")
	test.That(t, sdc.conf.Pages["home"][0].Text, test.ShouldEqual, "changed")
	test.That(t, len(sdc.conf.Keys), test.ShouldEqual, 0)

	// The page-switching that this used to break still works.
	setPage(t, sdc, "other")
	test.That(t, fd.button(0), test.ShouldEqual, "text:other0")
	setPage(t, sdc, "home")
	test.That(t, fd.button(0), test.ShouldEqual, "text:changed")
}

func TestPagedUpdateByNameAndIndex(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, keyUpdate("status", map[string]interface{}{"text": "by name"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(1), test.ShouldEqual, "text:by name")

	_, err = update(sdc, keyUpdate("1", map[string]interface{}{"text": "by index"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(1), test.ShouldEqual, "text:by index")
}

func TestPagedNamesAreScopedToTheTargetPage(t *testing.T) {
	// "status" is key 1 on home and key 0 on other.
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, map[string]interface{}{
		"page": "other",
		"keys": map[string]interface{}{"status": map[string]interface{}{"text": "hidden"}},
	})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, sdc.conf.Pages["other"][0].Text, test.ShouldEqual, "hidden")
	test.That(t, sdc.conf.Pages["home"][1].Text, test.ShouldEqual, "home1")
	test.That(t, fd.button(1), test.ShouldEqual, "text:home1")
}

func TestPagedHiddenPageIsNotDrawnUntilShown(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, map[string]interface{}{
		"page": "other",
		"keys": map[string]interface{}{"0": map[string]interface{}{"text": "later"}},
	})
	test.That(t, err, test.ShouldBeNil)

	// Nothing on the deck changed.
	test.That(t, fd.button(0), test.ShouldEqual, "text:home0")
	test.That(t, fd.button(2), test.ShouldEqual, "")

	// Still true after the once-a-second redraw.
	test.That(t, sdc.reconfigure(context.Background(), sdc.deps, sdc.conf), test.ShouldBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "text:home0")

	setPage(t, sdc, "other")
	test.That(t, fd.button(0), test.ShouldEqual, "text:later")
	test.That(t, fd.button(2), test.ShouldEqual, "text:other2")
}

func TestPagedExplicitPageThatIsShowingDrawsNow(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	_, err := update(sdc, map[string]interface{}{
		"page": "home",
		"keys": map[string]interface{}{"0": map[string]interface{}{"text": "now"}},
	})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "text:now")
}

func TestPagedDefaultsToWhicheverPageIsShowing(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	setPage(t, sdc, "other")

	_, err := update(sdc, keyUpdate("0", map[string]interface{}{"text": "on other"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "text:on other")
	test.That(t, sdc.conf.Pages["other"][0].Text, test.ShouldEqual, "on other")
	test.That(t, sdc.conf.Pages["home"][0].Text, test.ShouldEqual, "home0")
}

func TestPagedAddKeyOnlyToShowingPage(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	newKey := map[string]interface{}{"text": "new", "component": "thing", "method": "do_command"}

	_, err := update(sdc, map[string]interface{}{
		"page": "other",
		"keys": map[string]interface{}{"7": newKey},
	})
	mustErr(t, err, "only be added to the page that is showing")
	test.That(t, len(sdc.conf.Pages["other"]), test.ShouldEqual, 2)

	_, err = update(sdc, keyUpdate("7", newKey))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(7), test.ShouldEqual, "text:new")
	test.That(t, len(sdc.conf.Pages["home"]), test.ShouldEqual, 3)
}

func TestPagedRejectsUndrawableKeysWhenSent(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	before := sdc.conf.Pages["other"][0]

	// On a page that is not showing, so nothing would be drawn until later.
	for name, fields := range map[string]map[string]interface{}{
		"unknown image": {"image": "nope.png"},
		"bad method":    {"method": "get_readings"},
		"blank method":  {"method": ""},
		"no text":       {"text": ""},
	} {
		_, err := update(sdc, map[string]interface{}{
			"page": "other",
			"keys": map[string]interface{}{"0": fields},
		})
		test.That(t, err, test.ShouldNotBeNil)
		test.That(t, sdc.conf.Pages["other"][0], test.ShouldResemble, before)
		_ = name
	}

	// And the page can still be shown.
	setPage(t, sdc, "other")
	test.That(t, fd.button(0), test.ShouldEqual, "text:other0")
}

func TestPagedRejectedUpdateChangesNothing(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, map[string]interface{}{"keys": map[string]interface{}{
		"0": map[string]interface{}{"text": "good"},
		"1": map[string]interface{}{"image": "nope.png"},
	}})
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "text:home0")
	test.That(t, sdc.conf.Pages["home"][0].Text, test.ShouldEqual, "home0")
}

func TestPagedBadReferences(t *testing.T) {
	sdc, _ := newTestComponent(t, pagedConfig(), "thing")

	_, err := update(sdc, keyUpdate("4abc", map[string]interface{}{"text": "x"}))
	mustErr(t, err, "invalid key reference")

	_, err = update(sdc, keyUpdate("nobody", map[string]interface{}{"text": "x"}))
	mustErr(t, err, "invalid key reference")

	_, err = update(sdc, map[string]interface{}{
		"page": "missing",
		"keys": map[string]interface{}{"0": map[string]interface{}{"text": "x"}},
	})
	mustErr(t, err, "page missing not found")

	_, err = update(sdc, map[string]interface{}{"keys": map[string]interface{}{
		"0":      map[string]interface{}{"text": "x"},
		"status": map[string]interface{}{"text": "y"},
	}})
	test.That(t, err, test.ShouldBeNil) // different keys: 0 and 1

	_, err = update(sdc, map[string]interface{}{"keys": map[string]interface{}{
		"1":      map[string]interface{}{"text": "x"},
		"status": map[string]interface{}{"text": "y"},
	}})
	mustErr(t, err, "addressed more than once")
}

func TestAmbiguousName(t *testing.T) {
	conf := flatConfig()
	conf.Keys[0].Name, conf.Keys[1].Name = "dup", "dup"
	sdc, _ := newTestComponent(t, conf, "thing")
	_, err := update(sdc, keyUpdate("dup", map[string]interface{}{"text": "x"}))
	mustErr(t, err, "ambiguous")
}

// With pages the rule is strict: "4abc" is not a key number, so it is looked up
// as a name, and no key has that name.
func TestPagedMalformedKeyReferenceIsAnUnknownName(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	_, err := update(sdc, keyUpdate("1abc", map[string]interface{}{"text": "sneaky"}))
	mustErr(t, err, "no key has that name")
	test.That(t, fd.button(1), test.ShouldEqual, "text:home1")
	test.That(t, sdc.conf.Pages["home"][1].Text, test.ShouldEqual, "home1")
}

func TestPagedMissingComponentIsCrossedCircleOnEveryPage(t *testing.T) {
	conf := pagedConfig()
	conf.Pages["other"][1].Component = "gone"
	conf.Pages["home"][1].Component = "gone"
	sdc, fd := newTestComponent(t, conf, "thing")
	test.That(t, fd.button(1), test.ShouldEqual, "image+text:gone")

	setPage(t, sdc, "other")
	test.That(t, fd.button(2), test.ShouldEqual, "image+text:gone")

	// An update that points a key at a missing component is not an error.
	_, err := update(sdc, keyUpdate("0", map[string]interface{}{"component": "also-gone"}))
	test.That(t, err, test.ShouldBeNil)
	test.That(t, fd.button(0), test.ShouldEqual, "image+text:also-gone")
}

func TestPagedSurvivesAMixedSequence(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := update(sdc, keyUpdate("status", map[string]interface{}{"text": "home"}))
		test.That(t, err, test.ShouldBeNil)
		_, err = update(sdc, map[string]interface{}{
			"page": "other",
			"keys": map[string]interface{}{"status": map[string]interface{}{"text": "other"}},
		})
		test.That(t, err, test.ShouldBeNil)

		setPage(t, sdc, "other")
		test.That(t, fd.button(0), test.ShouldEqual, "text:other")
		test.That(t, sdc.reconfigure(ctx, sdc.deps, sdc.conf), test.ShouldBeNil)
		setPage(t, sdc, "home")
		test.That(t, fd.button(1), test.ShouldEqual, "text:home")
	}
	test.That(t, len(sdc.conf.Keys), test.ShouldEqual, 0)
}

func TestPagedBrightnessAndDialsUnchanged(t *testing.T) {
	sdc, fd := newTestComponent(t, pagedConfig(), "thing")
	res, err := update(sdc, map[string]interface{}{"brightness": 30})
	test.That(t, err, test.ShouldBeNil)
	test.That(t, res, test.ShouldResemble, map[string]interface{}{"brightness": 30})
	test.That(t, fd.brightness, test.ShouldEqual, uint16(30))
}

func TestConfigNameIsOptionalAndNotValidatedAtLoad(t *testing.T) {
	conf := flatConfig()
	conf.Keys[0].Name = "has spaces, symbols & 12"
	conf.Keys[1].Name = "has spaces, symbols & 12" // even a duplicate loads
	_, _, err := conf.Validate("")
	test.That(t, err, test.ShouldBeNil)
}
