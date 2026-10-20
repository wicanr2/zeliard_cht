// capture_music 在Docker內錄取原版AdLib配樂，不跳過正在播放的曲段。
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/zeliard_cht/internal/prototype/fm"
	"github.com/wicanr2/zeliard_cht/internal/prototype/sound"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root, out, library := os.Args[1], os.Args[2], os.Args[3]
	if _, e := os.Stat(out); !os.IsNotExist(e) {
		panic("輸出必須是新目錄")
	}
	if e := os.MkdirAll(filepath.Join(out, "scratch"), 0755); e != nil {
		panic(e)
	}
	config, e := os.ReadFile(filepath.Join(root, "RESOURCE.CFG"))
	if e != nil {
		panic(e)
	}
	config = []byte(strings.ReplaceAll(string(config), "MSCSTD.DRV", "MSCADLIB.DRV"))
	if e = os.WriteFile(filepath.Join(out, "scratch/RESOURCE.CFG"), config, 0644); e != nil {
		panic(e)
	}
	o, e := oracle.Load(filepath.Join(root, "ZELIARD.EXE"), root)
	if e != nil {
		panic(e)
	}
	defer o.Close()
	o.SetScratch(filepath.Join(out, "scratch"))
	o.SetDOSBoxCycles(3000)
	o.SetAdLib(true)
	synth, e := fm.Open(library, 44100)
	if e != nil {
		panic(e)
	}
	defer synth.Close()
	audio := o.NewAudio(44100)
	audio.SetFMSynth(synth)
	wav, e := sound.CreateWAV(filepath.Join(out, "source.wav"))
	if e != nil {
		panic(e)
	}
	notes, writes, samples, peak := 0, 0, 0, 0
	for i := 0; i < 800; i++ {
		if i == 80 {
			o.KeyDown(0x39)
		}
		if i == 84 {
			o.KeyUp(0x39)
		}
		if e = o.RunCycles(150000); e != nil {
			panic(e)
		}
		for _, w := range o.OPLWrites() {
			writes++
			if w.Reg >= 0xb0 && w.Reg <= 0xb8 && w.Val&32 != 0 {
				notes++
			}
		}
		data := audio.Render()
		samples += len(data)
		for _, v := range data {
			a := int(v)
			if a < 0 {
				a = -a
			}
			if a > peak {
				peak = a
			}
		}
		if e = wav.Write(data); e != nil {
			panic(e)
		}
	}
	if e = wav.Close(); e != nil {
		panic(e)
	}
	data, e := os.ReadFile(filepath.Join(out, "source.wav"))
	if e != nil {
		panic(e)
	}
	meta := map[string]any{"source": "原版MSCAdLib驅動，冷開機後4秒按下Space，200毫秒後放開，繼續40秒虛擬時間，不再跳曲", "rate": 44100, "samples": samples, "peak": peak, "opl_writes": writes, "key_ons": notes, "steps": o.Steps(), "cycles": o.Cycles(), "sha256": fmt.Sprintf("%x", sha256.Sum256(data)), "unimplemented": o.Unimplemented(), "rights": "original music, local only"}
	b, _ := json.MarshalIndent(meta, "", "  ")
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(b, '\n'), 0644); e != nil {
		panic(e)
	}
	fmt.Println("CAPTURED_ORIGINAL_MUSIC", notes, writes, samples, peak)
}
