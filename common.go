package viamstreamdeck

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/dh1tw/streamdeck"

	"golang.org/x/image/colornames"
)

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	result := ""

	for _, part := range parts {
		result += strings.ToUpper(string(part[0])) + part[1:]
	}

	return result
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

func (ms *ModelSetup) SimpleText(text string, clr string, textFont *string) []streamdeck.TextLine {
	return ms.simpleText(text, clr, 20, textFont)
}

func (ms *ModelSetup) simpleText(text string, clr string, fontSize float64, textFont *string) []streamdeck.TextLine {
	if fontSize <= 0 {
		panic(fontSize)
	}

	maxLine := int(1.6 * float64(ms.Conf.ButtonSize) / fontSize)

	lines := []string{""}

	for _, s := range strings.Split(text, " ") {

		if len(s) >= maxLine && fontSize > 4 {
			return ms.simpleText(text, clr, fontSize-2, textFont)
		}

		last := lines[len(lines)-1]

		if last != "" && (len(last)+1+len(s)) > maxLine {
			lines = append(lines, s)
		} else {
			if last == "" {
				last = s
			} else {
				last = last + " " + s
			}
			lines[len(lines)-1] = last
		}
	}

	if fontSize > 4 && len(lines) >= ms.Conf.ButtonSize/int(fontSize) {
		return ms.simpleText(text, clr, fontSize-2, textFont)
	}

	tls := []streamdeck.TextLine{}

	for idx, l := range lines {
		tl := streamdeck.TextLine{
			Text:      l,
			PosX:      5,
			PosY:      (idx * int(fontSize)),
			FontSize:  fontSize,
			FontColor: getColor(clr, "white"),
		}
		if textFont != nil {
			tl.Font = GetFont(*textFont)
		}
		tls = append(tls, tl)
	}

	return tls
}

func (ms *ModelSetup) SimpleTextButton(text string, bgColor, textClr string, textFont *string) streamdeck.TextButton {
	return streamdeck.TextButton{
		Lines:   ms.SimpleText(text, textClr, textFont),
		BgColor: getColor(bgColor, "black"),
	}
}
