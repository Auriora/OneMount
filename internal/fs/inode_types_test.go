package fs

import (
	"testing"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestVirtualInodeContentHelpers(t *testing.T) {
	inode := NewInode(".xdg-volume-info", fuse.S_IFREG|0644, nil)
	inode.SetVirtualContent([]byte("initial"))

	if !inode.IsVirtual() {
		t.Fatalf("expected inode to be virtual")
	}

	if got, err := inode.WriteVirtualContent(len("initial"), []byte(" rename")); err != nil || got != len(" rename") {
		t.Fatalf("WriteVirtualContent failed, bytes=%d err=%v", got, err)
	}

	if content := string(inode.ReadVirtualContent(0, 32)); content != "initial rename" {
		t.Fatalf("unexpected content after write: %s", content)
	}

	if err := inode.TruncateVirtualContent(7); err != nil {
		t.Fatalf("TruncateVirtualContent shrink failed: %v", err)
	}
	if content := string(inode.ReadVirtualContent(0, 32)); content != "initial" {
		t.Fatalf("unexpected content after shrink truncate: %s", content)
	}

	if err := inode.TruncateVirtualContent(10); err != nil {
		t.Fatalf("TruncateVirtualContent grow failed: %v", err)
	}
	data := inode.ReadVirtualContent(0, 10)
	if string(data[:7]) != "initial" {
		t.Fatalf("existing bytes changed unexpectedly: %s", string(data))
	}
	for idx, b := range data[7:] {
		if b != 0 {
			t.Fatalf("expected zero padding at index %d, got %d", 7+idx, b)
		}
	}
}
