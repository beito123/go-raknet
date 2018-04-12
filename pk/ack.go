package pk

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

import (
	"github.com/beito123/go-raknet"
	"github.com/beito123/go-raknet/binary"
	"sort"
)

type ACK struct {
	Acknowledge
}

func (ACK) ID() byte {
	return IDACK
}

func (pk *ACK) Encode() error {
	return pk.Acknowledge.Encode(pk)
}

func (pk *ACK) Decode() error {
	return pk.Acknowledge.Decode(pk)
}

func (ACK) New() raknet.Packet {
	return new(ACK)
}

type NACK struct {
	Acknowledge
}

func (NACK) ID() byte {
	return IDNACK
}

func (pk *NACK) Encode() error {
	return pk.Acknowledge.Encode(pk)
}

func (pk *NACK) Decode() error {
	return pk.Acknowledge.Decode(pk)
}

func (NACK) New() raknet.Packet {
	return new(NACK)
}

type Acknowledge struct {
	BasePacket

	Records []raknet.Record
}

func (ack *Acknowledge) Encode(pk raknet.Packet) error {
	err := ack.BasePacket.Encode(pk)
	if err != nil {
		return err
	}

	ack.Records = CondenseRecords(ack.Records)

	err = ack.PutShort(uint16(len(ack.Records)))
	if err != nil {
		return err
	}

	for _, rec := range ack.Records {
		noRange := !rec.IsRanged() // 0 = ranged, 1 = no ranged

		err = ack.PutBool(noRange)
		if err != nil {
			return err
		}

		err = ack.PutLTriad(binary.Triad(rec.Index))
		if err != nil {
			return err
		}

		if !noRange { // ranged
			err = ack.PutLTriad(binary.Triad(rec.EndIndex))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ack *Acknowledge) Decode(pk raknet.Packet) error {
	err := ack.BasePacket.Decode(pk)
	if err != nil {
		return err
	}

	var recLen uint16
	err = ack.Short(&recLen)
	if err != nil {
		return err
	}

	ack.Records = []raknet.Record{}
	for i := 0; i < int(recLen); i++ {
		var noRange bool
		err = ack.Bool(&noRange)
		if err != nil {
			return err
		}

		var index binary.Triad

		err = ack.LTriad(&index)
		if err != nil {
			return err
		}

		var endIndex binary.Triad
		if noRange {
			err = ack.LTriad(&endIndex)
			if err != nil {
				return err
			}
		}

		ack.Records = append(ack.Records, raknet.Record{
			Index:    int(index),
			EndIndex: int(endIndex),
		})
	}

	ack.Records = simplifyRecords(ack.Records)

	return nil
}

// CondenseRecords returns condensed records.
// For example (No need sort): [0, 2, 3, 5, 8, 9, 10, 15] -> [0, [2:3], 5, [8:10], 15]
func CondenseRecords(records []raknet.Record) []raknet.Record {
	var ids []int
	for _, record := range records {
		ids = append(ids, record.Numbers()...)
	}

	sort.Ints(ids)

	ln := len(ids)

	var nRecords []raknet.Record
	for i := 0; i < ln; i++ {
		rec := ids[i]
		last := rec

		// find
		if i+1 < ln {
			for last+1 == ids[i+1] {
				last = ids[i+1]
				i++
				if i+1 >= ln {
					break
				}
			}
		}

		end := last

		if rec == end { // no ranged
			nRecords = append(nRecords, raknet.Record{
				Index: rec,
			})
		} else { // ranged
			nRecords = append(nRecords, raknet.Record{
				Index:    rec,
				EndIndex: end,
			})
		}
	}

	return nRecords
}

func simplifyRecords(records []raknet.Record) []raknet.Record {
	var ids []int
	for _, rec := range records {
		ids = append(ids, rec.Numbers()...)
	}

	var recs []raknet.Record
	for _, rec := range ids {
		recs = append(recs, raknet.Record{
			Index: rec,
		})
	}

	return recs
}
