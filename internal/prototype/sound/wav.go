package sound

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"
)

// WAV僅保存更新執行緒取得的原版單聲道取樣，不代表裝置實際播放內容。
type WAV struct {
	file   *os.File
	hash   hash.Hash
	bytes  uint64
	closed bool
}

func CreateWAV(path string) (*WAV, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	w := &WAV{file: f, hash: sha256.New()}
	if _, err = f.Write(w.header()); err != nil {
		_ = f.Close()
		return nil, err
	}
	return w, nil
}

func (w *WAV) header() []byte {
	b := []byte("RIFF")
	b = binary.LittleEndian.AppendUint32(b, uint32(36+w.bytes))
	b = append(b, []byte("WAVEfmt ")...)
	b = binary.LittleEndian.AppendUint32(b, 16)
	b = binary.LittleEndian.AppendUint16(b, 1)
	b = binary.LittleEndian.AppendUint16(b, 1)
	b = binary.LittleEndian.AppendUint32(b, Rate)
	b = binary.LittleEndian.AppendUint32(b, Rate*2)
	b = binary.LittleEndian.AppendUint16(b, 2)
	b = binary.LittleEndian.AppendUint16(b, 16)
	b = append(b, []byte("data")...)
	return binary.LittleEndian.AppendUint32(b, uint32(w.bytes))
}

func (w *WAV) Write(values []int16) error {
	if w.closed {
		return fmt.Errorf("原版音訊記錄已關閉")
	}
	if uint64(len(values))*2+w.bytes > uint64(^uint32(0))-36 {
		return fmt.Errorf("原版音訊記錄超過WAV容量")
	}
	b := make([]byte, 0, len(values)*2)
	for _, v := range values {
		b = binary.LittleEndian.AppendUint16(b, uint16(v))
	}
	n, err := w.file.Write(b)
	w.bytes += uint64(n)
	_, _ = w.hash.Write(b[:n])
	if err == nil && n != len(b) {
		err = io.ErrShortWrite
	}
	return err
}

func (w *WAV) State() map[string]any {
	if w == nil {
		return nil
	}
	return map[string]any{"rate": Rate, "channels": 1, "samples": w.bytes / 2, "pcm_sha256": fmt.Sprintf("%x", w.hash.Sum(nil)), "closed": w.closed, "scope": "原版生成取樣；不是裝置監聽錄音"}
}

func (w *WAV) Close() error {
	if w == nil || w.closed {
		return nil
	}
	w.closed = true
	_, err := w.file.Seek(0, io.SeekStart)
	if err == nil {
		_, err = w.file.Write(w.header())
	}
	closeErr := w.file.Close()
	if err != nil {
		return err
	}
	return closeErr
}
