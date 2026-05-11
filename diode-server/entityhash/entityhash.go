package entityhash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	diodepb "github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
)

// marshalBufPool reuses byte buffers for protobuf serialization. Without
// pooling, GenerateEntityHash allocates a fresh buffer sized to the entity's
// wire format on every call — significant GC pressure on the hot ingestion
// path, where this function runs once per ingested entity.
var marshalBufPool = sync.Pool{
	New: func() any { b := make([]byte, 0, 4096); return &b },
}

const metadataFieldName = "metadata"

// EntityFingerprinter generates deterministic content hashes for entities.
//
// Hashes are computed over the entity's protobuf binary encoding (with
// Deterministic=true) after metadata fields are cleared. Binary protobuf is
// substantially cheaper to produce than canonical JSON (RFC 8785), which the
// prior protojson + JCS path required.
//
// Hash bytes are not stable across proto schema changes (Deterministic mode is
// stable within a build but not guaranteed across versions). That's acceptable
// here because content_hash is recomputed on every ingestion: a schema change
// regenerates hashes naturally on the next pass, and the ON CONFLICT path
// uses (node_type, external_id), not content_hash, as the upsert key.
type EntityFingerprinter struct {
	marshaler proto.MarshalOptions
}

// NewEntityFingerprinter creates a new entity fingerprinter.
func NewEntityFingerprinter() *EntityFingerprinter {
	return &EntityFingerprinter{
		marshaler: proto.MarshalOptions{Deterministic: true},
	}
}

// savedMetadataPool reuses the slice of saved metadata fields collected per
// hash call. Each call typically saves only a handful of entries (one per
// nested message that has metadata set), but the slice keeps capacity
// across calls instead of allocating a fresh one every entity.
var savedMetadataPool = sync.Pool{
	New: func() any {
		s := make([]savedMetadata, 0, 8)
		return &s
	},
}

// savedMetadata captures one metadata field that was temporarily cleared on
// a proto message and needs to be restored after marshaling.
type savedMetadata struct {
	msg   protoreflect.Message
	fd    protoreflect.FieldDescriptor
	value protoreflect.Value
}

// GenerateEntityHash returns a SHA256 hex digest of the entity's content,
// excluding metadata fields.
//
// Implementation note: this mutates the caller's entity in-place
// (clearing metadata fields), marshals, then restores the saved values.
// This avoids deep-cloning the entity on every hash, which would otherwise
// dominate the function's cost. Callers must ensure no other goroutine reads
// the same entity concurrently — this function is not safe to call in
// parallel on a shared entity.
func (f *EntityFingerprinter) GenerateEntityHash(entity *diodepb.Entity) (string, error) {
	if entity == nil {
		return "", fmt.Errorf("entity cannot be nil")
	}
	if entity.GetEntity() == nil {
		return "", fmt.Errorf("entity content cannot be nil")
	}

	// Wrap inner entity content (no timestamp) without cloning. The wrapper
	// shares the inner pointer with the caller's entity; mutations to
	// nested messages below propagate to the caller, which is why we
	// restore metadata before returning.
	wrapper := &diodepb.Entity{Entity: entity.GetEntity()}

	savedPtr := savedMetadataPool.Get().(*[]savedMetadata)
	saved := (*savedPtr)[:0]
	defer func() {
		// Restore in reverse order in case the same message has nested
		// metadata fields collected after the parent's. Reverse order is
		// not strictly necessary today (each savedMetadata has its own
		// Message handle), but keeps the operation symmetric.
		for i := len(saved) - 1; i >= 0; i-- {
			s := saved[i]
			s.msg.Set(s.fd, s.value)
			saved[i] = savedMetadata{} // help the GC drop the protoreflect.Value's ref
		}
		*savedPtr = saved[:0]
		savedMetadataPool.Put(savedPtr)
	}()

	saved = saveAndClearMetadata(wrapper.ProtoReflect(), saved)

	bufPtr := marshalBufPool.Get().(*[]byte)
	buf := (*bufPtr)[:0]
	defer func() {
		*bufPtr = buf
		marshalBufPool.Put(bufPtr)
	}()

	buf, err := f.marshaler.MarshalAppend(buf, wrapper)
	if err != nil {
		return "", fmt.Errorf("marshal entity: %w", err)
	}

	hash := sha256.Sum256(buf)
	return hex.EncodeToString(hash[:]), nil
}

// saveAndClearMetadata walks m (and its nested messages/lists/maps) and
// records every set "metadata" field, clearing each as it goes. Returns the
// updated slice so callers can pre-allocate via sync.Pool.
func saveAndClearMetadata(m protoreflect.Message, saved []savedMetadata) []savedMetadata {
	if !m.IsValid() {
		return saved
	}

	if mfd := m.Descriptor().Fields().ByName(metadataFieldName); mfd != nil && m.Has(mfd) {
		saved = append(saved, savedMetadata{msg: m, fd: mfd, value: m.Get(mfd)})
		m.Clear(mfd)
	}

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			if fd.Kind() == protoreflect.MessageKind {
				list := v.List()
				for i := 0; i < list.Len(); i++ {
					saved = saveAndClearMetadata(list.Get(i).Message(), saved)
				}
			}
		case fd.IsMap():
			if fd.MapValue().Kind() == protoreflect.MessageKind {
				v.Map().Range(func(_ protoreflect.MapKey, mv protoreflect.Value) bool {
					saved = saveAndClearMetadata(mv.Message(), saved)
					return true
				})
			}
		case fd.Kind() == protoreflect.MessageKind:
			saved = saveAndClearMetadata(v.Message(), saved)
		}
		return true
	})

	return saved
}
