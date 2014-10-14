// This packege is utilities for LTSV(Labeled Tab Separated Value)
// if you want more LTSV information, see http://ltsv.org/
package goltsv

import (
	"bytes"
	"strings"
)

const (
	KV_SEP   = ":"
	ITEM_SEP = "\t"
)

// Parse line LTSV to map[string]string
func ParseLtsv(line string, filter_label ...string) map[string]string {
	pMap := make(map[string]string)
	if line == "" {
		return pMap
	}

	for _, i := range strings.Split(line, ITEM_SEP) {
		kv := strings.SplitN(i, KV_SEP, 2)

		if len(kv) != 2 {
			pMap[kv[0]] = ""
		} else {
			pMap[kv[0]] = kv[1]
		}
	}

	return pMap
}

func ParseLtsvFilter(line string, filter_label ...string) map[string]string {
	pMap := make(map[string]string)
	if line == "" || filter_label == nil || len(filter_label) == 0 {
		return pMap
	}

	labels := array2set(filter_label)
	for _, i := range strings.Split(line, ITEM_SEP) {
		kv := strings.SplitN(i, KV_SEP, 2)

		_, exists := labels[kv[0]]
		if !exists {
			continue
		}

		if len(kv) != 2 {
			pMap[kv[0]] = ""
		} else {
			pMap[kv[0]] = kv[1]
		}
	}

	return pMap
}

func array2set(keys []string) map[string]bool {
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return set
}

// Convert map[string]string to LTSV line
func Map2Ltsv(items map[string]string) string {
	if items == nil || len(items) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	for k, v := range items {
		buffer.WriteString(k)
		buffer.WriteString(KV_SEP)
		buffer.WriteString(v)
		buffer.WriteString(ITEM_SEP)
	}

	buf := bytes.TrimRight(buffer.Bytes(), ITEM_SEP)
	return string(buf)
}

// Convert map[string]string to LTSV line. this func can specify order.
func Map2OrderedLtsv(items map[string]string, labels ...string) string {
	if items == nil || len(items) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	for _, l := range labels {
		buffer.WriteString(l)
		buffer.WriteString(KV_SEP)
		buffer.WriteString(items[l])
		buffer.WriteString(ITEM_SEP)
	}

	buf := bytes.TrimRight(buffer.Bytes(), ITEM_SEP)
	return string(buf)
}

// Convert map[string]string to TSV line. this func can specify order.
func Map2OrderedTsv(items map[string]string, labels ...string) string {
	if items == nil || len(items) == 0 {
		return ""
	}

	var buffer bytes.Buffer
	for _, l := range labels {
		buffer.WriteString(items[l])
		buffer.WriteString(ITEM_SEP)
	}

	buf := bytes.TrimRight(buffer.Bytes(), ITEM_SEP)
	return string(buf)
}
