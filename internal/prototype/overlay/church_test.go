package overlay

import (
	"bytes"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

func TestChurchNormalServiceReadOnly(t *testing.T) {
	if os.Getenv("ZELIARD_BANK_SAVE") == "" || os.Getenv("ZELIARD_CHURCH_SOURCE") == "" {
		t.Skip("須掛載自建存檔與正常教堂原版收據")
	}
	if out := os.Getenv("ZELIARD_SHOP_OUTPUT"); out != "" {
		t.Setenv("ZELIARD_SHOP_OUTPUT", filepath.Join(out, "church"))
	}
	testWeaponShopNormalPath(t, "../../../tools/church.commands", 16)
}

func assertChurchCheckpoint(t *testing.T, name string, d *Dialogue, o *oracle.Oracle, dst *image.RGBA, drawn bool) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(os.Getenv("ZELIARD_CHURCH_SOURCE"), name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Steps, Cycles uint64
		Regs          oracle.Regs
		Memory        string `json:"memory_sha256"`
		Indexed       string `json:"indexed_sha256"`
		Gold          uint16
	}
	if err = json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	if o.Steps() != source.Steps || o.Cycles() != source.Cycles || o.Regs() != source.Regs || fmtHash(o.Bytes(oracle.Far(0, 0), 1<<20)) != source.Memory || fmtHash(o.Indexed()) != source.Indexed || source.Gold != 1000 || o.Word(oracle.Far(0x1b1, 0x86)) != 1000 {
		t.Fatal("教堂配對未重現正常原版來源", name)
	}
	keys := map[string]string{
		"church-entered":          "zelres2.file.01a508",
		"church-healing-page2":    "zelres2.file.01a568",
		"church-second-visit":     "zelres2.file.01a508",
		"church-second-final":     "zelres2.file.01a568",
		"church-wounded-greeting": "zelres2.file.01a4cd",
		"church-wounded-page1":    "zelres2.file.01a4cd",
		"church-wounded-page2":    "zelres2.file.01a568",
	}
	if key := keys[name]; key != "" {
		if !drawn || !d.active || d.current < 0 || d.paragraphs[d.current].key != key || len(d.lead) == 0 {
			t.Fatal("教堂中文來源或前文遺失", name, d.State(), d.Rejections)
		}
		if name == "church-healing-page2" || name == "church-second-final" || name == "church-wounded-page2" {
			if d.visible != len(d.paragraphs[d.current].text) || o.Word(oracle.Far(0x1b1, 0x90)) != 80 {
				t.Fatal("教堂祝福未完整或原版回復值不同", name)
			}
		}
		isolated := *d
		unknown := append([]byte(nil), o.Indexed()...)
		unknown[100*320+55] ^= 128
		canvas := image.NewRGBA(dst.Bounds())
		before := append([]byte(nil), canvas.Pix...)
		if isolated.Draw(canvas, unknown, o.Palette(), image.Point{}) || isolated.active || !bytes.Equal(before, canvas.Pix) {
			t.Fatal("未知教堂畫面未完整退回", name)
		}
	} else if name == "church-healing-after" || name == "church-actually-outside" || name == "church-wounded-final" || name == "church-wounded-outside" {
		if drawn || d.Active() {
			t.Fatal("教堂離店後殘留中文", name)
		}
	}
	if value := map[string]uint16{"climb-return-attempt": 75, "wounded-town-settled": 77, "church-wounded-door": 77, "church-wounded-greeting": 77, "church-wounded-page1": 80}[name]; value != 0 {
		if o.Word(oracle.Far(0x1b1, 0x90)) != value {
			t.Fatal("未使用正常受傷存活返城的狀態", name)
		}
	}
}
