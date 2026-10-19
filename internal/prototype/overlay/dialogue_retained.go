package overlay

import "fmt"

// 規格021已證實的順序。原版仍執行續頁及捲動，譯文只用於顯示。
var exhaustionReturnKeys = []string{
	"zelres2.file.01d72e", "zelres2.file.01d76c",
	"zelres2.file.01d79c", "zelres2.file.01d7b2",
}

func (d *Dialogue) validateRetainedDialogue() error {
	count := 0
	var full []rune
	for _, key := range exhaustionReturnKeys {
		for _, p := range d.paragraphs {
			if p.key == key {
				count++
				full = append(full, p.text...)
			}
		}
	}
	if count == 0 {
		return nil
	}
	if count != len(exhaustionReturnKeys) {
		return fmt.Errorf("送返長老對話須四個字段完整提供")
	}
	_, err := dialogueLayout(d.font, full)
	return err
}

// 缺前段、前段未完成或失序時，整框保留原文，不猜補前文。
func (d *Dialogue) retainedDialogueLead(next int) ([]rune, bool) {
	key := d.paragraphs[next].key
	for i, candidate := range exhaustionReturnKeys {
		if key != candidate {
			continue
		}
		if i == 0 {
			return nil, true
		}
		if !d.active || d.current < 0 || d.current >= len(d.paragraphs) {
			return nil, false
		}
		previous := d.paragraphs[d.current]
		if previous.key != exhaustionReturnKeys[i-1] || d.visible != len(previous.text) {
			return nil, false
		}
		return append(append([]rune(nil), d.lead...), previous.text...), true
	}
	if d.current >= 0 && d.paragraphs[d.current].key == "zelres2.file.01af36" && key == "zelres2.file.01aeee" {
		return append([]rune(nil), d.paragraphs[d.current].text...), true
	}
	return nil, true
}

// State提供觀察器唯讀進度，正常GUI可依原版字段等待。
func (d *Dialogue) State() map[string]any {
	if d == nil {
		return nil
	}
	state := map[string]any{"active": d.active, "key": "", "visible": 0, "total": 0, "retained": len(d.lead), "prompt": d.prompt, "pending": d.pending, "complete": false, "glyphs": d.Glyphs, "scrolls": d.Scrolls}
	if d.active && d.current >= 0 && d.current < len(d.paragraphs) {
		p := d.paragraphs[d.current]
		state["key"], state["visible"], state["total"], state["complete"] = p.key, d.visible, len(p.text), d.visible == len(p.text)
	}
	return state
}
