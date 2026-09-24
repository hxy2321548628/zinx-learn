package protocol

import (
	"errors"
	"io"
	"testing"
)

func TestDataPackRoundTrip(t *testing.T) {
	t.Parallel()

	packer := NewDataPack(1024)
	want := NewMessage(7, []byte("hello"))
	packet, err := packer.Pack(want)
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	headerLen := int(packer.HeaderLen())
	messageID, dataLen, err := packer.UnpackHeader(packet[:headerLen])
	if err != nil {
		t.Fatalf("UnpackHeader() error = %v", err)
	}
	if messageID != want.ID() {
		t.Fatalf("message ID = %d, want %d", messageID, want.ID())
	}
	if dataLen != uint32(len(want.Data())) {
		t.Fatalf("data length = %d, want %d", dataLen, len(want.Data()))
	}
	if got := string(packet[headerLen:]); got != string(want.Data()) {
		t.Fatalf("data = %q, want %q", got, want.Data())
	}
}

func TestDataPackRejectsInvalidMessages(t *testing.T) {
	t.Parallel()

	packer := NewDataPack(4)
	if _, err := packer.Pack(nil); err == nil {
		t.Fatal("Pack(nil) error = nil")
	}
	if _, err := packer.Pack(NewMessage(1, []byte("hello"))); err == nil {
		t.Fatal("Pack() accepted an oversized message")
	}
	if _, _, err := packer.UnpackHeader(make([]byte, headerLen-1)); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("UnpackHeader() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}

func TestDataPackRejectsOversizedHeader(t *testing.T) {
	t.Parallel()

	unlimited := NewDataPack(0)
	packet, err := unlimited.Pack(NewMessage(1, []byte("hello")))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	limited := NewDataPack(4)
	if _, _, err := limited.UnpackHeader(packet[:limited.HeaderLen()]); err == nil {
		t.Fatal("UnpackHeader() accepted an oversized message")
	}
}
