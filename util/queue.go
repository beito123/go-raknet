package util

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import "github.com/damnever/goqueue"

func NewQueue() (que *Queue) {
	que = &Queue{
		list: goqueue.New(0),
	}

	return que
}

// Queue is a simple wrapper for goqueue
// Thank you: https://github.com/damnever/goqueue
type Queue struct {
	list *goqueue.Queue
}

func (queue *Queue) Size() int {
	return queue.list.Size()
}

func (queue *Queue) Add(data interface{}) {
	queue.list.PutNoWait(data)
}

func (queue *Queue) Poll() (item interface{}, exist bool) {
	item, err := queue.list.GetNoWait()
	if err != nil {
		return nil, false
	}

	return item, true
}
