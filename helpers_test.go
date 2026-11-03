package viamstreamdeck

import (
	"context"
	"image"
	"strings"
	"sync"
	"testing"

	"github.com/dh1tw/streamdeck"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/test"
)

// fakeDeck records what the component draws, in place of a physical deck.
type fakeDeck struct {
	mu         sync.Mutex
	buttons    map[int]string
	brightness uint16
	clearedAll int
}

func newFakeDeck() *fakeDeck {
	return &fakeDeck{buttons: map[int]string{}}
}

func linesText(lines []streamdeck.TextLine) string {
	parts := []string{}
	for _, l := range lines {
		parts = append(parts, l.Text)
	}
	return strings.Join(parts, " ")
}

func (f *fakeDeck) set(idx int, what string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buttons[idx] = what
	return nil
}

// button returns what is drawn on a key, or "" if it is blank.
func (f *fakeDeck) button(idx int) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.buttons[idx]
}

func (f *fakeDeck) SetBtnEventCb(streamdeck.BtnEvent) {}

func (f *fakeDeck) SetBrightness(b uint16) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.brightness = b
	return nil
}

func (f *fakeDeck) ClearBtn(idx int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.buttons, idx)
	return nil
}

func (f *fakeDeck) ClearAllBtns() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buttons = map[int]string{}
	f.clearedAll++
	return nil
}

func (f *fakeDeck) FillImage(idx int, _ image.Image) error {
	return f.set(idx, "image")
}

func (f *fakeDeck) WriteText(idx int, tb streamdeck.TextButton) error {
	return f.set(idx, "text:"+linesText(tb.Lines))
}

func (f *fakeDeck) WriteTextOnImage(idx int, _ image.Image, lines []streamdeck.TextLine) error {
	return f.set(idx, "image+text:"+linesText(lines))
}

func (f *fakeDeck) Close() error { return nil }

// fakeThing is a stand-in for any component a key points at.
type fakeThing struct {
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	name resource.Name
}

func (f *fakeThing) Name() resource.Name { return f.name }

func (f *fakeThing) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (f *fakeThing) Status(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// newTestComponent builds a component on a fakeDeck, the way NewStreamDeck
// does, without starting stateChecker. things are the components present in deps.
func newTestComponent(t *testing.T, conf *Config, things ...string) (*streamdeckComponent, *fakeDeck) {
	t.Helper()

	_, _, err := conf.Validate("")
	test.That(t, err, test.ShouldBeNil)

	deps := resource.Dependencies{}
	for _, n := range things {
		name := generic.Named(n)
		deps[name] = &fakeThing{name: name}
	}

	fd := newFakeDeck()
	sdc := &streamdeckComponent{
		name:   generic.Named("deck"),
		logger: logging.NewTestLogger(t),
		ms:     ModelOriginal2,
		sd:     fd,
		conf:   conf,
		deps:   deps,
		keys:   map[int]KeyConfig{},
	}
	if len(conf.Pages) > 0 {
		sdc.currentPage = conf.InitialPage
	}
	test.That(t, sdc.updateKeys(context.Background()), test.ShouldBeNil)
	return sdc, fd
}

// update sends an update_display command.
func update(sdc *streamdeckComponent, body map[string]interface{}) (map[string]interface{}, error) {
	return sdc.DoCommand(context.Background(), map[string]interface{}{"update_display": body})
}

// keyUpdate is shorthand for an update_display body that edits one key.
func keyUpdate(ref string, fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"keys": map[string]interface{}{ref: fields}}
}

func textKey(idx int, text string) KeyConfig {
	return KeyConfig{Key: idx, Text: text, Component: "thing", Method: "do_command"}
}

func mustErr(t *testing.T, err error, contains string) {
	t.Helper()
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, contains)
}
