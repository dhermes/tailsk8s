// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tka

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/hdevalence/ed25519consensus"
	"tailscale.com/types/tkatype"
)

// KeyKind describes the different varieties of a Key.
type KeyKind uint8

// Valid KeyKind values.
const (
	KeyInvalid KeyKind = iota
	Key25519
)

func (k KeyKind) String() string {
	switch k {
	case KeyInvalid:
		return "invalid"
	case Key25519:
		return "25519"
	default:
		return fmt.Sprintf("Key?<%d>", int(k))
	}
}

// Key describes the public components of a key known to network-lock.
type Key struct {
	Kind KeyKind `cbor:"1,keyasint"`

	// Votes describes the weight applied to signatures using this key.
	// Weighting is used to deterministically resolve branches in the AUM
	// chain (i.e. forks, where two AUMs exist with the same parent).
	Votes uint `cbor:"2,keyasint"`

	// Public encodes the public key of the key. For 25519 keys,
	// this is simply the point on the curve representing the public
	// key.
	Public []byte `cbor:"3,keyasint"`

	// Meta describes arbitrary metadata about the key. This could be
	// used to store the name of the key, for instance.
	Meta map[string]string `cbor:"12,keyasint,omitempty"`
}

// Clone makes an independent copy of Key.
//
// NOTE: There is a difference between a nil slice and an empty slice for encoding purposes,
// so an implementation of Clone() must take care to preserve this.
func (k Key) Clone() Key {
	out := k

	if k.Public != nil {
		out.Public = make([]byte, len(k.Public))
		copy(out.Public, k.Public)
	}

	if k.Meta != nil {
		out.Meta = make(map[string]string, len(k.Meta))
		for k, v := range k.Meta {
			out.Meta[k] = v
		}
	}

	return out
}

func (k Key) ID() tkatype.KeyID {
	switch k.Kind {
	// Because 25519 public keys are so short, we just use the 32-byte
	// public as their 'key ID'.
	case Key25519:
		return tkatype.KeyID(k.Public)
	default:
		panic("unsupported key kind")
	}
}

const maxMetaBytes = 512

func (k Key) StaticValidate() error {
	if k.Votes > 4096 {
		return fmt.Errorf("excessive key weight: %d > 4096", k.Votes)
	}
	if k.Votes == 0 {
		return errors.New("key votes must be non-zero")
	}

	// We have an arbitrary upper limit on the amount
	// of metadata that can be associated with a key, so
	// people don't start using it as a key-value store and
	// causing pathological cases due to the number + size of
	// AUMs.
	var metaBytes uint
	for k, v := range k.Meta {
		metaBytes += uint(len(k) + len(v))
	}
	if metaBytes > maxMetaBytes {
		return fmt.Errorf("key metadata too big (%d > %d)", metaBytes, maxMetaBytes)
	}

	switch k.Kind {
	case Key25519:
	default:
		return fmt.Errorf("unrecognized key kind: %v", k.Kind)
	}
	return nil
}

// Verify returns a nil error if the signature is valid over the
// provided AUM BLAKE2s digest, using the given key.
func signatureVerify(s *tkatype.Signature, aumDigest tkatype.AUMSigHash, key Key) error {
	// NOTE(tom): Even if we can compute the public from the KeyID,
	//            its possible for the KeyID to be attacker-controlled
	//            so we should use the public contained in the state machine.
	switch key.Kind {
	case Key25519:
		if ed25519consensus.Verify(ed25519.PublicKey(key.Public), aumDigest[:], s.Signature) {
			return nil
		}
		return errors.New("invalid signature")

	default:
		return fmt.Errorf("unhandled key type: %v", key.Kind)
	}
}
