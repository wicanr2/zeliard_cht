package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/hd"
)

type hdLoadJob struct {
	name        string
	when        func() bool
	load, apply func()
	done        chan error
	finished    bool
	started     time.Time
}

func (g *game) deferHD(name string, when func() bool, load, apply func()) {
	g.hdJobs = append(g.hdJobs, &hdLoadJob{name: name, when: when, load: load, apply: apply})
}

func (g *game) updateHDLoading() bool {
	if job := g.hdLoading; job != nil {
		select {
		case err := <-job.done:
			if err != nil {
				log.Printf("%s未載入：%v", job.name, err)
			} else if job.apply != nil {
				job.apply()
			}
			job.finished = true
			g.hdLoading = nil
			g.releaseGate = true
			g.clock.Reset()
			g.lastRenderedAt = time.Time{}
			log.Printf("場景HD準備完成：%s，耗時 %.2fs", job.name, time.Since(job.started).Seconds())
		default:
		}
		return true
	}
	if !g.hdEnabled || g.panel || g.paused || !ebiten.IsFocused() {
		return false
	}
	for _, job := range g.hdJobs {
		if job.finished || !job.when() {
			continue
		}
		g.edges.Sync(g, nil)
		g.clock.Reset()
		g.releaseGate = true
		job.started = time.Now()
		job.done = make(chan error, 1)
		g.hdLoading = job
		log.Printf("按需載入場景HD：%s", job.name)
		// 載入期間Update不推進VM，Draw只畫等待頁。Oracle掛鉤在主執行緒apply。
		go func() {
			var err error
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
				job.done <- err
			}()
			job.load()
		}()
		return true
	}
	return false
}

func (g *game) drawHDLoading(screen *ebiten.Image) {
	job := g.hdLoading
	if job == nil {
		return
	}
	draw.Draw(g.canvas, image.Rect(120, 230, 840, 450), &image.Uniform{color.RGBA{25, 32, 45, 255}}, image.Point{}, draw.Src)
	g.text(image.Rect(150, 250, 810, 290), "正在準備"+job.name)
	g.text(image.Rect(150, 315, 810, 350), fmt.Sprintf("已等待 %.1f 秒，遊戲暫停。", time.Since(job.started).Seconds()))
	g.text(image.Rect(150, 370, 810, 410), "只載入目前場景所需圖組，完成後繼續。")
	g.texture.WritePixels(g.canvas.Pix)
	screen.DrawImage(g.texture, nil)
}

func (g *game) caveModule() bool { return hd.CaveRuntimeModule(g.o) }
func (g *game) townModule() bool {
	if g.o.VideoMode() != 0x13 || g.caveModule() {
		return false
	}
	w := g.o.Word(oracle.Far(0x1b1, 0xc002))
	return w >= 32 && w <= 1024
}
func (g *game) roomDialogue() bool {
	if g.dialogue == nil {
		return false
	}
	key, _ := g.dialogue.State()["key"].(string)
	return key != ""
}

func (g *game) caveFamily() string {
	if g.caveTerrainHD == nil || !g.caveModule() {
		return ""
	}
	// 清單識別只在地圖及資源header改變時重算，避免每次Update複製1MiB。
	key := string(g.o.Bytes(oracle.Far(0x1b1, 0xc000), 32)) + string(g.o.Bytes(oracle.Far(0x11b1, 0x8000), 48))
	if key == g.hdFamilySignature {
		return g.hdFamily
	}
	family := hd.CurrentCaveTerrainFamily(g.o)
	if family != "" {
		g.hdFamilySignature, g.hdFamily = key, family
	}
	return family
}
