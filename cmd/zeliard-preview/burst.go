package main

import "github.com/wicanr2/dosgolem/oracle"

// burstBucket 是 VRAM 寫入統計的時間粒度（原版指令數）。
const burstBucket = 500

// vramBurst 統計最近的 VRAM 寫入量，用來判斷原版是否正在大量重畫畫面。
// 畫到一半的畫面不該送出：城鎮、洞窟與商店建構時，HD 守衛只認完成的畫面。規格066。
type vramBurst struct {
	o       *oracle.Oracle
	buckets [16]int
	start   uint64
}

func newVRAMBurst(o *oracle.Oracle) *vramBurst {
	b := &vramBurst{o: o, start: o.Steps()}
	o.WatchEveryWrite(0xa0000, 0xaf9ff, func(uint32, uint8, uint8) {
		b.advance()
		b.buckets[0]++
	})
	return b
}

func (b *vramBurst) advance() {
	now := b.o.Steps()
	if now >= b.start+burstBucket*uint64(len(b.buckets)) {
		b.buckets = [16]int{}
		b.start = now - now%burstBucket
		return
	}
	for now >= b.start+burstBucket {
		copy(b.buckets[1:], b.buckets[:len(b.buckets)-1])
		b.buckets[0] = 0
		b.start += burstBucket
	}
}

// Recent 回傳最近 n 個桶（含目前未滿的桶）的 VRAM 寫入 byte 數。
func (b *vramBurst) Recent(n int) int {
	b.advance()
	total := 0
	for _, v := range b.buckets[:min(n, len(b.buckets))] {
		total += v
	}
	return total
}
