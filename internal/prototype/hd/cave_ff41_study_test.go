package hd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 僅研究：複製解碼輸入以隔離FF41守衛，不改原版機器或保存來源。
func TestCaveFF41SourceStudy(t *testing.T) {
	out := os.Getenv("ZELIARD_FF41_STUDY_OUTPUT")
	if out == "" {
		t.Skip("須指定新的本機研究輸出；skip不算驗收")
	}
	files, err := os.ReadDir(out)
	if err != nil || len(files) != 0 {
		t.Fatal("須使用既有空研究目錄", err)
	}
	base := filepath.Join(iceStudy(t), "cave-ff41-source-v1")
	read := func(ext string) []byte { return iceRead(t, filepath.Join(base, "source."+ext)) }
	memory, body, raw := read("memory"), read("body.memory"), read("indexed")
	if townBackgroundHash(memory) != "c43f08361e64e6abaa0b9f0f1818d42305556a89c5652937e28e60cbc7b52a6c" || townBackgroundHash(body) != "0fa4da56ec7b063d0185e449410f04a562d224283e3a29ba5f4aa2856472dffc" || townBackgroundHash(raw) != "b9c80da87e0095ebf2488acdc3e9672b9c2b7c841c8c1151820b486c6ba57988" {
		t.Fatal("固定被拒來源身份不符")
	}
	var paints []CavePaint
	if err := json.Unmarshal(read("paints.json"), &paints); err != nil {
		t.Fatal(err)
	}
	bp := body[0x1b10:0x11b10]
	if bp[0xff41] != 1 || bp[0xff40] != 0 {
		t.Fatal("研究條件改變")
	}
	isolated := cloneTownBytes(body)
	isolated[0x1b10+0xff41] = 0
	d, decodeErr := decodeCaveFrameWithPaints(memory, isolated, raw, paints)
	report := map[string]any{"schema": 1, "status": "DRAFT_FF41_GUARD_ISOLATION", "original_body_sha256": townBackgroundHash(body), "decoder_body_sha256": townBackgroundHash(isolated), "decoder_input_only_override": "01B1:FF41 1→0; 原版機器與來源不變", "error": nil}
	if decodeErr != nil {
		report["error"] = decodeErr.Error()
	}
	if d != nil {
		report["viewport_differences"] = d.Differences
		report["actor_key"] = townBackgroundHash(d.ActorRaster[:])
		report["weapon_key"] = townBackgroundHash(d.WeaponRaster[:])
		report["objects"] = d.Objects
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(out, "receipt.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
