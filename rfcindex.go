package rfctree

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type RfcIndex struct {
	rfcMap             map[string]*RfcLabel
	excludeMap         map[string]bool
	connectionScoreMap map[string]map[string]*ConnectionScore
}

const defaultRfcIndexSize = 10000

func NewRfcIndex(rfcsDir string, targets []*Target, keywords []string, follow bool, excludes []string) (*RfcIndex, error) {
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

	excludeMap := make(map[string]bool)
	relationMap := make(map[string]map[string]*ConnectionScore)

	rfcIndex := RfcIndex{
		rfcMap:             rfcMap,
		excludeMap:         excludeMap,
		connectionScoreMap: relationMap,
	}

	rfcIndex.buildConnectionScore()

	rfcIndex.SetTargets(targets, keywords, follow, excludes)

	return &rfcIndex, nil
}

func (r *RfcIndex) SetTargets(targets []*Target, keywords []string, follow bool, excludes []string) {
	for _, docId := range excludes {
		r.excludeMap[docId] = true
	}

	for _, target := range targets {
		if r.IsExcluded(target.DocId) {
			continue
		}

		rfc := r.Find(target.DocId)
		rfc.IsTarget = true
		rfc.Position.Y = target.PositionY
		rfc.Groups = target.Groups
	}

	for _, rfc := range r.Values() {
		if r.IsExcluded(rfc.DocId) {
			continue
		}

		for _, keyword := range keywords {
			if slices.Contains(rfc.Keywords, keyword) {
				rfc.IsTarget = true
				break
			}
		}
	}

	if follow {
		for _, rfc := range r.GetTargets() {
			if r.IsExcluded(rfc.DocId) {
				continue
			}

			r.followConnection(rfc.DocId)
		}
	}

	//for _, rfc := range r.GetTargets() {
	//	rfc.DescendantYRange = r.getDescendantYRange(rfc.DocId)
	//}

}

func (r *RfcIndex) followConnection(rfcId string) {
	rfc := r.Find(rfcId)

	for _, id := range rfc.Obsoletes {
		if r.IsExcluded(id) {
			continue
		}
		if r.Find(id).IsTarget {
			continue
		}
		r.Find(id).IsTarget = true
		r.followConnection(id)
	}

	for _, id := range rfc.ObsoletedBy {
		if r.IsExcluded(id) {
			continue
		}
		if r.Find(id).IsTarget {
			continue
		}
		r.Find(id).IsTarget = true
		r.followConnection(id)
	}

	for _, id := range rfc.Updates {
		if r.IsExcluded(id) {
			continue
		}
		if r.Find(id).IsTarget {
			continue
		}
		if r.getConnectionScore(rfcId, id).Value() < 1.1 {
			continue
		}
		r.Find(id).IsTarget = true
		r.followConnection(id)
	}

	for _, id := range rfc.UpdatedBy {
		if r.IsExcluded(id) {
			continue
		}
		if r.Find(id).IsTarget {
			continue
		}
		if r.getConnectionScore(rfcId, id).Value() < 1.1 {
			continue
		}
		r.Find(id).IsTarget = true
		r.followConnection(id)
	}
}

func (r *RfcIndex) buildConnectionScore() {
	const (
		obsoletesScore                = 10.0
		updatesBaseScore              = 2.0
		updatesSharedScore            = 6.0
		obsoletesSiblingBaseScore     = 1.0
		obsoletesSiblingSharedScore   = 2.0
		obsoletedBySiblingBaseScore   = 1.0
		obsoletedBySiblingSharedScore = 2.0
		updatesSiblingBaseScore       = 1.0
		updatesSiblingSharedScore     = 2.0
		updatedBySiblingBaseScore     = 1.0
		updatedBySiblingSharedScore   = 2.0
	)

	for _, rfc := range r.Values() {
		if !rfc.IsTarget {
			continue
		}

		for _, rfcId := range rfc.Updates {
			connectionScore := r.getConnectionScore(rfcId, rfc.DocId)
			connectionScore.updates = updatesBaseScore + updatesSharedScore/float64(len(r.rfcMap[rfcId].UpdatedBy))
		}

		for _, rfcId := range rfc.Obsoletes {
			connectionScore := r.getConnectionScore(rfcId, rfc.DocId)
			connectionScore.obsoletes = obsoletesScore
		}

		if len(rfc.Obsoletes) > 1 {
			for _, rfcIdA := range rfc.Obsoletes {
				for _, rfcIdB := range rfc.Obsoletes {
					if rfcIdB <= rfcIdA {
						continue
					}
					connectionScore := r.getConnectionScore(rfcIdA, rfcIdB)
					connectionScore.obsoletesSibling = obsoletesSiblingBaseScore + obsoletesSiblingSharedScore/float64(len(rfc.Obsoletes))
				}
			}
		}

		if len(rfc.ObsoletedBy) > 1 {
			for _, rfcIdA := range rfc.ObsoletedBy {
				for _, rfcIdB := range rfc.ObsoletedBy {
					if rfcIdB <= rfcIdA {
						continue
					}
					connectionScore := r.getConnectionScore(rfcIdA, rfcIdB)
					connectionScore.obsoletedBySibling = obsoletedBySiblingBaseScore + obsoletedBySiblingSharedScore/float64(len(rfc.ObsoletedBy))
				}
			}
		}

		if len(rfc.Updates) > 1 {
			for _, rfcIdA := range rfc.Updates {
				for _, rfcIdB := range rfc.Updates {
					if rfcIdB <= rfcIdA {
						continue
					}
					connectionScore := r.getConnectionScore(rfcIdA, rfcIdB)
					connectionScore.updatesSibling = updatesSiblingBaseScore + updatesSiblingSharedScore/float64(len(rfc.Updates))
				}
			}
		}

		if len(rfc.UpdatedBy) > 1 {
			for _, rfcIdA := range rfc.UpdatedBy {
				for _, rfcIdB := range rfc.UpdatedBy {
					if rfcIdB <= rfcIdA {
						continue
					}
					connectionScore := r.getConnectionScore(rfcIdA, rfcIdB)
					connectionScore.updatedBySibling = updatedBySiblingBaseScore + updatedBySiblingSharedScore/float64(len(rfc.UpdatedBy))
				}
			}
		}
	}
}

func (r *RfcIndex) getConnectionScore(rfcIdA, rfcIdB string) *ConnectionScore {
	var elderRfcId, youngerRfcId string
	if rfcIdA < rfcIdB {
		elderRfcId = rfcIdA
		youngerRfcId = rfcIdB
	} else {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}
	if r.rfcMap[rfcIdB].PubDateInMonth < r.rfcMap[rfcIdA].PubDateInMonth {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}

	connectionScoreMap, ok := r.connectionScoreMap[elderRfcId]
	if !ok {
		connectionScoreMap = make(map[string]*ConnectionScore)
		r.connectionScoreMap[elderRfcId] = connectionScoreMap
	}
	connectionScore, ok := connectionScoreMap[youngerRfcId]
	if !ok {
		connectionScore = &ConnectionScore{}
		connectionScoreMap[youngerRfcId] = connectionScore
	}
	return connectionScore
}

func (r *RfcIndex) getConnectionScoreValue(rfcIdA, rfcIdB string) float64 {
	var elderRfcId, youngerRfcId string
	if rfcIdA < rfcIdB {
		elderRfcId = rfcIdA
		youngerRfcId = rfcIdB
	} else {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}
	if r.rfcMap[rfcIdB].PubDateInMonth < r.rfcMap[rfcIdA].PubDateInMonth {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}

	connectionScoreMap, ok := r.connectionScoreMap[elderRfcId]
	if !ok {
		return 0
	}
	connectionScore, ok := connectionScoreMap[youngerRfcId]
	if !ok {
		return 0
	}
	return connectionScore.Value()
}

func (r *RfcIndex) Find(docId string) *RfcLabel {
	return r.rfcMap[docId]
}

func (r *RfcIndex) IsExcluded(docId string) bool {
	isExcluded, ok := r.excludeMap[docId]
	return ok && isExcluded
}

func (r *RfcIndex) getDescendantYRange(rfcId string) int {
	relationScoreMap, ok := r.connectionScoreMap[rfcId]
	if !ok {
		return 0
	}

	rfcIds := slices.SortedFunc(maps.Keys(relationScoreMap), func(a, b string) int {
		return cmp.Compare(r.Find(a).PubDateInMonth, r.Find(b).PubDateInMonth)
	})
	rfcIds = slices.Insert(rfcIds, 0, rfcId)
	rfcIds = slices.DeleteFunc(rfcIds, func(a string) bool {
		return !r.Find(a).IsTarget
	})

	first := r.Find(rfcId).PubDateInMonth
	last := r.Find(rfcIds[len(rfcIds)-1]).PubDateInMonth

	n := make([]int, last-first+1+int(labelSize.W/cellSize.W))
	for _, id := range rfcIds {
		m := r.Find(id).PubDateInMonth - first
		n[m] = n[m] + 1
	}

	yRange := make([]int, last-first+1)
	for i := range last - first + 1 {
		s := 0
		for j := range int(labelSize.W / cellSize.W) {
			s = s + n[i+j]
		}
		yRange[i] = s
	}
	return slices.Max(yRange)
}

func (r *RfcIndex) GetTargets() []*RfcLabel {
	var rfcs []*RfcLabel
	for _, rfc := range r.Values() {
		if rfc.IsTarget {
			rfcs = append(rfcs, rfc)
		}
	}
	return rfcs
}

func (r *RfcIndex) Clear() {
	for _, rfc := range r.Values() {
		if rfc.IsTarget {
			rfc.IsTarget = false
		}
	}
	r.excludeMap = make(map[string]bool)
	r.connectionScoreMap = make(map[string]map[string]*ConnectionScore)
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

type ConnectionScore struct {
	obsoletes          float64
	updates            float64
	obsoletesSibling   float64
	obsoletedBySibling float64
	updatesSibling     float64
	updatedBySibling   float64
}

func (r *ConnectionScore) Value() float64 {
	return r.obsoletes + r.updates + r.obsoletesSibling + r.obsoletedBySibling + r.updatesSibling + r.updatedBySibling
}

var ErrInvalidCSVRecord = errors.New("invalid csv record")

type Target struct {
	DocId     string
	Title     string
	PositionY int
	Groups    []string
}

func NewTarget(record []string) (target *Target, err error) {
	var (
		docId, title string
		positionY    int
		groups       []string
	)

	l := len(record)

	// DocID, Title, PositionY, Group...
	if l == 0 {
		return nil, ErrInvalidCSVRecord
	}
	docId = record[0]
	if !strings.HasPrefix(docId, "RFC") {
		return nil, ErrInvalidCSVRecord
	}

	if l > 1 {
		title = record[1]
	}

	if l > 2 && record[2] != "" {
		positionY, err = strconv.Atoi(record[2])
		if err != nil {
			return nil, err
		}
	}

	for l > 3 {
		for i := 3; i < l; i++ {
			group := record[i]
			if group != "" {
				groups = append(groups, group)
			}
		}
	}

	target = &Target{
		DocId:     docId,
		Title:     title,
		PositionY: positionY,
		Groups:    groups,
	}
	return target, nil
}
