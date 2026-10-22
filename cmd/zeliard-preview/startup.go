package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/zeliard_cht/internal/prototype/bitmap"
)

// showStartupError也涵蓋封包定位／參數的早期錯誤，避免Windows隱藏主控台後無聲退出。
func showStartupError(err error) {
	logPath := ""
	if root, e := os.UserConfigDir(); e == nil {
		dir := filepath.Join(root, "Zeliard-CHT")
		if os.MkdirAll(dir, 0700) == nil {
			logPath = filepath.Join(dir, "startup.log")
			if f, e := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600); e == nil {
				defer f.Close()
				log.SetOutput(io.MultiWriter(os.Stderr, f))
			}
		}
	}
	boot := newStartup(logPath, func(report func(string, bitmap.Font)) (*game, error) {
		if exe, e := os.Executable(); e == nil {
			for _, base := range []string{filepath.Dir(exe), filepath.Join(filepath.Dir(exe), "../Resources/game")} {
				if font, e := bitmap.Load(filepath.Join(base, "font.hex.gz")); e == nil {
					report("啟動失敗", font)
					break
				}
			}
		}
		return nil, err
	})
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("神劍封魔：啟動失敗")
	ebiten.SetRunnableOnUnfocused(true)
	_ = ebiten.RunGame(boot)
	_ = boot.Close()
	log.Fatal(err)
}

type startupProgress struct {
	stage string
	font  bitmap.Font
}
type startupResult struct {
	game *game
	err  error
}
type startupCancelled struct{}
type startupGame struct {
	game                      *game
	progress                  chan startupProgress
	result                    chan startupResult
	done, cancel              chan struct{}
	once                      sync.Once
	font                      bitmap.Font
	stage, errorText, logPath string
	stageNumber               int
	failure                   error
	firstFrame                bool
	started                   time.Time
	canvas                    *image.RGBA
	texture                   *ebiten.Image
}

func newStartup(logPath string, load func(func(string, bitmap.Font)) (*game, error)) *startupGame {
	b := &startupGame{progress: make(chan startupProgress, 16), result: make(chan startupResult, 1), done: make(chan struct{}), cancel: make(chan struct{}), started: time.Now(), stage: "正在啟動", logPath: logPath, canvas: image.NewRGBA(image.Rect(0, 0, width, height))}
	go func() {
		defer close(b.done)
		defer func() {
			if r := recover(); r != nil {
				b.result <- startupResult{err: fmt.Errorf("啟動中止：%v", r)}
			}
		}()
		report := func(stage string, font bitmap.Font) {
			select {
			case <-b.cancel:
				panic(startupCancelled{})
			default:
			}
			log.Printf("啟動：%s，耗時 %.2fs", stage, time.Since(b.started).Seconds())
			b.progress <- startupProgress{stage, font}
		}
		g, err := load(report)
		b.result <- startupResult{g, err}
	}()
	return b
}

func (b *startupGame) Update() error {
	if b.game != nil {
		return b.game.Update()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	for {
		select {
		case p := <-b.progress:
			b.stage = p.stage
			b.stageNumber++
			if p.font != nil {
				b.font = p.font
			}
		case r := <-b.result:
			if r.err != nil {
				b.failure = r.err
				b.errorText = r.err.Error()
				log.Printf("啟動失敗：%v", r.err)
				continue
			}
			g := r.game
			g.texture = ebiten.NewImage(width, height)
			var err error
			g.output, err = newAudioOutput(g.audioMode, g.audioPath)
			if err != nil {
				if g.closeResources != nil {
					g.closeResources()
				}
				b.failure = err
				b.errorText = err.Error()
				log.Printf("音訊啟動失敗：%v", err)
				continue
			}
			b.game = g
			ebiten.SetWindowTitle(g.label("title"))
			log.Printf("啟動完成，耗時 %.2fs", time.Since(b.started).Seconds())
			return nil
		default:
			return nil
		}
	}
}

func (b *startupGame) Draw(screen *ebiten.Image) {
	if !b.firstFrame {
		b.firstFrame = true
		log.Printf("啟動視窗已顯示，耗時 %.2fs", time.Since(b.started).Seconds())
	}
	if b.game != nil {
		b.game.Draw(screen)
		return
	}
	if b.font == nil {
		screen.Fill(color.RGBA{18, 23, 32, 255})
		text := fmt.Sprintf("ZELIARD\nStarting... %.1fs\nEsc: cancel", time.Since(b.started).Seconds())
		if b.errorText != "" {
			text = "ZELIARD: STARTUP FAILED\n" + b.errorText + "\nLog: " + b.logPath + "\nEsc: close"
		}
		ebitenutil.DebugPrintAt(screen, text, 48, 96)
		return
	}
	draw.Draw(b.canvas, b.canvas.Bounds(), &image.Uniform{color.RGBA{18, 23, 32, 255}}, image.Point{}, draw.Src)
	draw.Draw(b.canvas, image.Rect(80, 150, 880, 530), &image.Uniform{color.RGBA{35, 46, 62, 255}}, image.Point{}, draw.Src)
	b.font.Draw(b.canvas, image.Rect(110, 175, 850, 220), "神劍封魔 Zeliard", white)
	title := b.stage
	if b.errorText != "" {
		title = "啟動失敗，請查看日誌"
	}
	b.font.Draw(b.canvas, image.Rect(110, 245, 850, 285), title, white)
	b.font.Draw(b.canvas, image.Rect(110, 295, 850, 330), fmt.Sprintf("已等待 %.1f 秒", time.Since(b.started).Seconds()), white)
	if b.errorText == "" {
		draw.Draw(b.canvas, image.Rect(110, 355, 850, 375), &image.Uniform{color.RGBA{18, 23, 32, 255}}, image.Point{}, draw.Src)
		progress := b.stageNumber
		if progress > 9 {
			progress = 9
		}
		draw.Draw(b.canvas, image.Rect(110, 355, 110+740*progress/9, 375), &image.Uniform{color.RGBA{42, 112, 140, 255}}, image.Point{}, draw.Src)
		b.font.Draw(b.canvas, image.Rect(110, 410, 850, 450), "正在準備必要資料，HD補圖將按需載入。", white)
	} else {
		b.font.Draw(b.canvas, image.Rect(110, 340, 850, 380), b.errorText, white)
		b.font.Draw(b.canvas, image.Rect(110, 400, 850, 440), b.logPath, white)
	}
	b.font.Draw(b.canvas, image.Rect(110, 470, 850, 510), "按 Esc 關閉啟動視窗", white)
	if b.texture == nil {
		b.texture = ebiten.NewImage(width, height)
	}
	b.texture.WritePixels(b.canvas.Pix)
	screen.DrawImage(b.texture, nil)
}

func (*startupGame) Layout(int, int) (int, int) { return width, height }

func (b *startupGame) Close() error {
	b.once.Do(func() { close(b.cancel) })
	if b.game != nil {
		err := b.game.output.close()
		if b.game.closeResources != nil {
			b.game.closeResources()
		}
		return err
	}
	select {
	case <-b.done:
		select {
		case r := <-b.result:
			if r.game != nil && r.game.closeResources != nil {
				r.game.closeResources()
			}
			return r.err
		default:
			return b.failure
		}
	case <-time.After(2 * time.Second):
		return errors.New("已取消啟動")
	}
}
