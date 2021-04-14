/* YaNFD - Yet another NDN Forwarding Daemon
 *
 * Copyright (C) 2020-2021 Eric Newberry.
 *
 * This file is licensed under the terms of the MIT License, as found in LICENSE.md.
 */

package table

import (
	"container/list"
	"time"

	"github.com/cespare/xxhash"
	"github.com/eric135/YaNFD/ndn"
)

// DeadNonceList represents the Dead Nonce List for a forwarding thread.
type DeadNonceList struct {
	list            map[uint64]byte
	ExpirationTimer chan bool
	expiringEntries list.List
}

// NewDeadNonceList creates a new Dead Nonce List for a forwarding thread.
func NewDeadNonceList() *DeadNonceList {
	d := new(DeadNonceList)
	d.list = make(map[uint64]byte)
	d.ExpirationTimer = make(chan bool, tableQueueSize)
	return d
}

// Find returns whether the specified name and nonce combination are present in the Dead Nonce List.
func (d *DeadNonceList) Find(name *ndn.Name, nonce []byte) (uint64, bool) {
	hash := xxhash.Sum64String(name.String() + string(nonce))
	_, ok := d.list[hash]
	return hash, ok
}

// Insert inserts an entry in the Dead Nonce List with the specified name and nonce. Returns hash and whether nonce already present.
func (d *DeadNonceList) Insert(name *ndn.Name, nonce []byte) (uint64, bool) {
	hash, exists := d.Find(name, nonce)

	if !exists {
		d.expiringEntries.PushBack(hash)
		go func() {
			time.Sleep(deadNonceListLifetime)
			d.ExpirationTimer <- true
		}()
	}
	return hash, exists
}

// RemoveExpiredEntry removes the front entry from Dead Nonce List.
func (d *DeadNonceList) RemoveExpiredEntry() {
	if d.expiringEntries.Len() > 0 {
		hash := d.expiringEntries.Front().Value.(uint64)
		delete(d.list, hash)
		d.expiringEntries.Remove(d.expiringEntries.Front())
	}
}
