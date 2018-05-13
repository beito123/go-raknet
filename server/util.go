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
	"time"
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
