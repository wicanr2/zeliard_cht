package hd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTownActorsNormalPathsReadOnly(t *testing.T) {
	project, output := os.Getenv("ZELIARD_PROJECT"), os.Getenv("ZELIARD_TOWN_NPC_OUTPUT")
	if project == "" || output == "" || os.Getenv("ZELIARD_TOWN_NPC_ART") == "" {
		t.Skip("須掛載正常鍵序、居民圖面及輸出；skip不算驗收")
	}
	runTownFramesPairWithActors(t, filepath.Join(project, "tools/town-residents-hd.commands"), output, true)
}
