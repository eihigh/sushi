package sushi

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// libvterm使うまでのつなぎ

// box - ターミナル描画機能のラッパ。libvterm使いだしたら捨てる
type box struct {
	x, y, w, h int
}

func (b *box) clear() {
	for x := 0; x < b.w; x++ {
		for y := 0; y < b.h; y++ {
			screen.SetContent(x+b.x, y+b.y, ' ', nil, tcell.StyleDefault)
		}
	}
}

//
// func (b *box) drawVTerm(vt *vterm.VTerm, vsc *vterm.Screen) {
// 	b.clear()
// 	height, width := vt.Size()
// 	for y := 0; y < height; y++ {
// 		for x := 0; x < width; x++ {
// 			cell, _ := vsc.GetCellAt(y, x)
// 			chs := cell.Chars()
// 			if len(chs) > 0 && chs[0] != 0 {
// 				cbg := cell.Bg()
// 				r, g, b := cbg.GetRGB()
// 				bg := tcell.FromImageColor(color.RGBA{r, g, b, 255})
// 				cfg := cell.Fg()
// 				r, g, b = cfg.GetRGB()
// 				fg := tcell.FromImageColor(color.RGBA{r, g, b, 255})
// 				style := tcell.StyleDefault.Background(bg).Foreground(fg)
// 				screen.SetContent(x, y, chs[0], chs[1:], style)
// 			}
// 		}
// 	}
// }

func (b *box) drawString(s string, cursor bool) {
	b.clear()
	x, y := 0, 0
	r := strings.NewReader(s)
	for {
		if x >= b.w || y >= b.h {
			break
		}
		ch, _, err := r.ReadRune()
		if err != nil {
			break
		}
		screen.SetContent(x+b.x, y+b.y, ch, nil, tcell.StyleDefault)
		if ch == '\n' {
			y++
			x = 0
		} else {
			x += runewidth.RuneWidth(ch)
		}
	}
	if cursor {
		screen.ShowCursor(x+b.x, y+b.y)
	}
}
