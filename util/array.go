package util

/*
 * go-raknet
 *
 * Copyright (c) 2018 beito
 *
 * This software is released under the MIT License.
 * http://opensource.org/licenses/mit-license.php
 */

// SplitBytesSlice splits the bytes with size
// Thanks you: https://qiita.com/suin/items/d0deb76ff03373b22a0b
func SplitBytesSlice(slice []byte, size int) [][]byte {
	var chunks [][]byte

	sliceSize := len(slice)

	for i := 0; i < sliceSize; i += size {
		end := i + size
		if sliceSize < end {
			end = sliceSize
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Strings is a simple sort type for []string
type Strings []string

func (s Strings) Len() int {
	return len(s)
}

func (s Strings) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s Strings) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
