package sound

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

func TestAudioOriginalPrefixFullStatePair(t *testing.T) {
	golden, metadata := originalPCM(t)
	orig, journalPath, out := os.Getenv("ZELIARD_ORIG"), os.Getenv("ZELIARD_AUDIO_JOURNAL"), os.Getenv("ZELIARD_AUDIO_PAIR_OUTPUT")
	if orig == "" || journalPath == "" || out == "" {
		t.Skip("須提供正常原版日誌與新的配對輸出；skip不計驗收")
	}
	journal, err := os.ReadFile(journalPath)
	if err != nil {
		t.Fatal(err)
	}
	hash := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	if hash(journal) != metadata["source_journal_sha256"] {
		t.Fatal("原生日誌身份不符")
	}
	if b, e := os.ReadFile(filepath.Join(orig, "ZELIARD.EXE")); e != nil || hash(b) != metadata["original_exe_sha256"] {
		t.Fatal("原版EXE身份不符", e)
	}
	scanner := bufio.NewScanner(bytes.NewReader(journal))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	if !scanner.Scan() {
		t.Fatal("缺少原生日誌標頭")
	}
	var header struct {
		Cycles uint64      `json:"cycles_per_ms"`
		Rate   int         `json:"audio_sample_rate"`
		Memory string      `json:"initial_memory_sha256"`
		Regs   oracle.Regs `json:"initial_regs"`
	}
	if err = json.Unmarshal(scanner.Bytes(), &header); err != nil {
		t.Fatal(err)
	}
	var machines [2]*oracle.Oracle
	var audio [2]*oracle.Audio
	for i := range machines {
		o, e := oracle.Load(filepath.Join(orig, "ZELIARD.EXE"), orig)
		if e != nil {
			t.Fatal(e)
		}
		defer o.Close()
		o.SetScratch(t.TempDir())
		o.SetDOSBoxCycles(header.Cycles)
		o.TraceFiles()
		machines[i], audio[i] = o, o.NewAudio(header.Rate)
		if hash(o.Bytes(oracle.Far(0, 0), 1<<20)) != header.Memory || o.Regs() != header.Regs {
			t.Fatal("冷開機身份不符")
		}
	}
	a, b := machines[0], machines[1]
	checks := 0
	equal := func() {
		checks++
		if a.Regs() != b.Regs() || a.Steps() != b.Steps() || a.Cycles() != b.Cycles() || !bytes.Equal(a.Bytes(oracle.Far(0, 0), 1<<20), b.Bytes(oracle.Far(0, 0), 1<<20)) || !bytes.Equal(a.Indexed(), b.Indexed()) || a.Palette() != b.Palette() || !bytes.Equal(a.HeldKeys(), b.HeldKeys()) || !reflect.DeepEqual(a.FileOps(), b.FileOps()) || !reflect.DeepEqual(a.KeyReads(), b.KeyReads()) {
			t.Fatal("播放緩衝改變完整原版", checks)
		}
	}
	equal()
	q := NewStream(Capacity)
	var actual []byte
	converted := 0
	records := 1
	wanted := int(metadata["records_including_header"].(float64))
	for records < wanted && scanner.Scan() {
		var r struct {
			Kind         string
			Scan         uint8
			Pressed      bool
			Budget       uint64      `json:"cycle_budget"`
			BeforeSteps  uint64      `json:"before_steps"`
			BeforeCycles uint64      `json:"before_cycles"`
			AfterSteps   uint64      `json:"after_steps"`
			AfterCycles  uint64      `json:"after_cycles"`
			Regs         oracle.Regs `json:"after_regs"`
			Audio        bool        `json:"audio_render"`
			Error        string
		}
		if err = json.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		for i, o := range machines {
			if o.Steps() != r.BeforeSteps || o.Cycles() != r.BeforeCycles {
				t.Fatal("原生日誌前置狀態")
			}
			switch r.Kind {
			case "key":
				if r.Pressed {
					o.KeyDown(r.Scan)
				} else {
					o.KeyUp(r.Scan)
				}
			case "run":
				if r.Budget > 5000000 || o.Steps() > 40000000 {
					t.Fatal("原版前綴預算")
				}
				e := o.RunCycles(r.Budget)
				message := ""
				if e != nil {
					message = e.Error()
				}
				if message != r.Error || o.Regs() != r.Regs {
					t.Fatal("原生日誌暫存器或錯誤")
				}
				if r.Audio {
					values := audio[i].Render()
					if i == 0 {
						for _, v := range values {
							actual = binary.LittleEndian.AppendUint16(actual, uint16(v))
						}
					} else {
						q.Push(values)
						p := make([]byte, len(values)*8)
						if _, e := io.ReadFull(q, p); e != nil {
							t.Fatal(e)
						}
						if !bytes.Equal(p, stereo(values)) {
							t.Fatal("原版取樣轉換錯誤")
						}
						converted += len(p)
					}
				}
			default:
				t.Fatal("未知原生日誌")
			}
			if o.Steps() != r.AfterSteps || o.Cycles() != r.AfterCycles {
				t.Fatal("原生日誌後置狀態")
			}
		}
		records++
		equal()
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if records != wanted || !bytes.Equal(actual, golden) || converted != len(golden)*4 || a.Steps() != uint64(metadata["steps"].(float64)) {
		t.Fatal("真實前綴或完整PCM不符", records, converted)
	}
	if s := q.Snapshot(); s.Underrun != 0 || s.Dropped != 0 || s.Queued != 0 {
		t.Fatal(s)
	}
	for _, o := range machines {
		if len(o.HeldKeys()) != 0 || len(o.Wrote()) != 0 || len(o.Unimplemented()) != 0 {
			t.Fatal("原版寫入、持鍵或未實作項")
		}
	}
	report := map[string]any{"status": "PASS_AUDIO_ORIGINAL_PREFIX_FULL_STATE_PAIR", "records_including_header": records, "full_original_comparisons": checks, "source_pcm_samples": len(actual) / 2, "source_pcm_sha256": hash(actual), "stereo_float_bytes": converted, "stream_stats": q.Snapshot(), "final_steps": a.Steps(), "final_cycles": a.Cycles(), "final_memory_sha256": hash(a.Bytes(oracle.Far(0, 0), 1<<20)), "state_injection": false, "snapshot_restore": false, "scope": "same normal cold-start prefix, Render discard versus bounded output conversion; actual device and full GUI separately verified"}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(out, "pair.json"), append(data, 10), 0644); err != nil {
		t.Fatal(err)
	}
	t.Log(records, checks, converted)
}
