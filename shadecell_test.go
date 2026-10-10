package shadecell

import (
	"math/rand"
	"testing"
	"time"
)

func TestSolids(t *testing.T) {
	q := New()
	if c := q.Pair(RGB{}, RGB{}); c != (Cell{Space, 0}) {
		t.Fatalf("black = %+v", c)
	}
	if c := q.Pair(RGB{255, 255, 255}, RGB{255, 255, 255}); c != (Cell{FullBlock, 15}) {
		t.Fatalf("white = %+v", c)
	}
	if c := q.Pair(RGB{0, 0, 170}, RGB{0, 0, 170}); c != (Cell{Space, 1 << 4}) {
		t.Fatalf("blue = %+v", c)
	}
}

func TestHalfBlocks(t *testing.T) {
	q := New()
	if c := q.Pair(RGB{255, 255, 255}, RGB{}); c != (Cell{HalfUpper, 15}) {
		t.Fatalf("white over black = %+v", c)
	}
	// A bright lower half must become the foreground of a lower half block,
	// since backgrounds stop at the 8 dark colours.
	if c := q.Pair(RGB{}, RGB{255, 85, 85}); c != (Cell{HalfLower, 12}) {
		t.Fatalf("black over light red = %+v", c)
	}
}

func TestShadesAppear(t *testing.T) {
	q := New()
	seen := map[byte]int{}
	for v := 0; v < 256; v += 4 {
		g := RGB{uint8(v), uint8(v), uint8(v)}
		c := q.Pair(g, g)
		seen[c.Ch]++
		fg, bg := c.Fg(), c.Bg()
		// A grey ramp must stay on grey pairs: no stray hues.
		isGrey := func(i byte) bool { return i == 0 || i == 7 || i == 8 || i == 15 }
		if !isGrey(fg) || !isGrey(bg) {
			t.Fatalf("grey %d -> fg %d bg %d (%#x)", v, fg, bg, c.Ch)
		}
	}
	if seen[ShadeLight]+seen[ShadeMedium]+seen[ShadeDark] == 0 {
		t.Fatalf("no shade glyphs on a grey ramp: %v", seen)
	}
	if seen[HalfUpper]+seen[HalfLower] != 0 {
		t.Fatalf("half blocks on a uniform cell: %v", seen)
	}
}

func TestSameHueRamp(t *testing.T) {
	q := New()
	// Mid red between dark red (170,0,0) and light red (255,85,85).
	c := q.Pair(RGB{212, 40, 40}, RGB{212, 40, 40})
	fg, bg := c.Fg(), c.Bg()
	if !(fg == 12 && bg == 4 || fg == 4 && bg == 4) {
		t.Fatalf("mid red = fg %d bg %d ch %#x", fg, bg, c.Ch)
	}
}

func TestDeterministicAndFast(t *testing.T) {
	q := New()
	r := rand.New(rand.NewSource(1))
	pairs := make([][2]RGB, 20000)
	for i := range pairs {
		pairs[i] = [2]RGB{{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256))}, {uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256))}}
	}
	start := time.Now()
	first := make([]Cell, len(pairs))
	for i, p := range pairs {
		first[i] = q.Pair(p[0], p[1])
	}
	cold := time.Since(start)
	start = time.Now()
	for i, p := range pairs {
		if q.Pair(p[0], p[1]) != first[i] {
			t.Fatalf("cell %d changed between calls", i)
		}
	}
	warm := time.Since(start)
	t.Logf("20000 random cells: cold %v, warm %v", cold, warm)
	if cold > 2*time.Second {
		t.Fatalf("too slow cold: %v", cold)
	}
}

func TestSGR(t *testing.T) {
	if s := SGR(4 | 1<<4); s != "0;31;44" { // red on blue
		t.Fatalf("red on blue = %q", s)
	}
	if s := SGR(15 | 0<<4); s != "1;37;40" {
		t.Fatalf("white on black = %q", s)
	}
	if s := SGR(9 | 6<<4); s != "1;34;43" { // light blue on brown
		t.Fatalf("got %q", s)
	}
}

func TestUTF8(t *testing.T) {
	if UTF8(HalfUpper) != "▀" || UTF8(ShadeMedium) != "▒" || UTF8('A') != "A" {
		t.Fatal("utf8 mapping")
	}
}
