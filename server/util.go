package server

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"math"
	"time"

	"github.com/beito123/go-raknet/binary"
)

const (
	PermanentExpire = -1
)

type Expire struct {
	Time     time.Time
	Duration time.Duration
}

func (exp *Expire) IsPermanent() bool {
	return exp.Duration < 0
}

func BumpTriad(id *binary.Triad) (result binary.Triad) {
	result = *id
	*id = id.Bump()
	return result
}

func BumpUInt16(id *uint16) (result uint16) {
	result = *id
	*id = (*id % math.MaxUint16) + 1
	return result
}
