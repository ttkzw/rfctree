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
	rfcMap map[string]*Rfc
}

const defaultRfcIndexSize = 10000

func NewRfcIndex(rfcsDir string) (RfcIndex, error) {
	files, err := os.ReadDir(rfcsDir)
	if err != nil {
		return RfcIndex{}, err
	}

	rfcMap := make(map[string]*Rfc, defaultRfcIndexSize)
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(rfcsDir, file.Name()))
		if err != nil {
			return RfcIndex{}, fmt.Errorf("%s: %v", file.Name(), err.Error())
		}

		rfc, err := NewRfc(data)
		if err == ErrRfcNotIssued {
			continue
		}
		if err != nil {
			return RfcIndex{}, fmt.Errorf("%s: %v", file.Name(), err.Error())
		}

		rfcMap[rfc.DocId] = rfc
	}

	rfcIndex := RfcIndex{
		rfcMap: rfcMap,
	}

	return rfcIndex, nil
}

func (r RfcIndex) Find(docId string) *Rfc {
	return r.rfcMap[docId]
}

func (r RfcIndex) FindAllByKeywords(keywords []string) []*Rfc {
	var rfcs []*Rfc
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

func (r RfcIndex) Exist(docId string) bool {
	_, ok := r.rfcMap[docId]
	return ok
}

func (r RfcIndex) Keys() []string {
	docIds := slices.SortedFunc(maps.Keys(r.rfcMap), func(a, b string) int {
		return cmp.Compare(a, b)
	})
	return docIds
}

func (r RfcIndex) Values() []*Rfc {
	rfcs := slices.SortedFunc(maps.Values(r.rfcMap), func(a, b *Rfc) int {
		return cmp.Compare(a.Doc.DocId, b.Doc.DocId)
	})
	return rfcs
}
