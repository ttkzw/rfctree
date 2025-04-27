package rfctree

import (
	"cmp"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type RfcIndex struct {
	rfcMap map[string]*RfcLabel
}

const defaultRfcIndexSize = 10000

func NewRfcIndex(rfcsDir string) (*RfcIndex, error) {
	files, err := os.ReadDir(rfcsDir)
	if err != nil {
		return nil, err
	}

	rfcMap := make(map[string]*RfcLabel, defaultRfcIndexSize)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(rfcsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %v", file.Name(), err.Error())
		}

		rfcDoc, err := NewRfcDoc(data)
		if err == ErrRfcNotIssued {
			continue
		}
		if err != nil {
			return nil, err
		}

		rfc, err := NewRfcLabel(rfcDoc)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", file.Name(), err.Error())
		}

		rfcMap[rfc.DocId] = rfc
	}

	rfcIndex := RfcIndex{
		rfcMap: rfcMap,
	}

	return &rfcIndex, nil
}

func (r *RfcIndex) Find(docId string) *RfcLabel {
	return r.rfcMap[docId]
}

func (r *RfcIndex) FindAllByKeywords(keywords []string) []*RfcLabel {
	var rfcs []*RfcLabel
	for _, rfc := range r.Values() {
		for _, keyword := range keywords {
			if slices.Contains(rfc.Keywords, keyword) {
				rfc.IsTarget = true
				break
			}
		}
		if rfc.IsTarget {
			rfcs = append(rfcs, rfc)
		}
	}
	return rfcs
}

func (r *RfcIndex) Exist(docId string) bool {
	_, ok := r.rfcMap[docId]
	return ok
}

func (r *RfcIndex) Keys() []string {
	docIds := slices.SortedFunc(maps.Keys(r.rfcMap), func(a, b string) int {
		return cmp.Compare(a, b)
	})
	return docIds
}

func (r *RfcIndex) Values() []*RfcLabel {
	rfcs := slices.SortedFunc(maps.Values(r.rfcMap), func(a, b *RfcLabel) int {
		return cmp.Compare(a.DocId, b.DocId)
	})
	return rfcs
}
