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
	rfcMap           map[string]*RfcLabel
	excludeMap       map[string]bool
	relationScoreMap map[string]map[string]*RelationScore
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
	relationMap := make(map[string]map[string]*RelationScore)

	rfcIndex := RfcIndex{
		rfcMap:           rfcMap,
		excludeMap:       excludeMap,
		relationScoreMap: relationMap,
	}

	rfcIndex.buildRelationScore()

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
		if rfc == nil {
			continue
		}
		rfc.IsTarget = true
		rfc.Position.Y = target.PositionY
		rfc.Groups = target.Groups
	}

	for _, rfc := range r.rfcMap {
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

			r.followRelation(rfc.DocId)
		}
	}

	for _, rfc := range r.GetTargets() {
		rfc.DescendantYRange = r.getDescendantYRange(rfc.DocId)
	}

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

func (r *RfcIndex) followRelation(rfcId string) {
	targetRfc := r.Find(rfcId)
	if targetRfc == nil {
		return
	}

	for _, id := range targetRfc.Obsoletes {
		if r.IsExcluded(id) {
			continue
		}
		rfc := r.Find(id)
		if rfc == nil {
			continue
		}
		if rfc.IsTarget {
			continue
		}
		rfc.IsTarget = true
		r.followRelation(id)
	}

	for _, id := range targetRfc.ObsoletedBy {
		if r.IsExcluded(id) {
			continue
		}
		rfc := r.Find(id)
		if rfc == nil {
			continue
		}
		if rfc.IsTarget {
			continue
		}
		rfc.IsTarget = true
		r.followRelation(id)
	}

	for _, id := range targetRfc.Updates {
		if r.IsExcluded(id) {
			continue
		}
		rfc := r.Find(id)
		if rfc == nil {
			continue
		}
		if rfc.IsTarget {
			continue
		}
		rfc.IsTarget = true
		r.followRelation(id)
	}

	for _, id := range targetRfc.UpdatedBy {
		if r.IsExcluded(id) {
			continue
		}
		rfc := r.Find(id)
		if rfc == nil {
			continue
		}
		if rfc.IsTarget {
			continue
		}
		rfc.IsTarget = true
		r.followRelation(id)
	}
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

func (r *RfcIndex) Clear() {
	for _, rfc := range r.rfcMap {
		if rfc.IsTarget {
			rfc.IsTarget = false
		}
	}
	r.excludeMap = make(map[string]bool)
	r.relationScoreMap = make(map[string]map[string]*RelationScore)
}

func (r *RfcIndex) Find(docId string) *RfcLabel {
	rfc, ok := r.rfcMap[docId]
	if !ok {
		return nil
	}
	return rfc
}

func (r *RfcIndex) IsExcluded(docId string) bool {
	isExcluded, ok := r.excludeMap[docId]
	return ok && isExcluded
}

func (r *RfcIndex) buildRelationScore() {
	const (
		obsoletesScore                = 100
		updatesBaseScore              = 20
		updatesSharedScore            = 50
		obsoletesSiblingBaseScore     = 10
		obsoletesSiblingSharedScore   = 20
		obsoletedBySiblingBaseScore   = 10
		obsoletedBySiblingSharedScore = 20
		updatesSiblingBaseScore       = 10
		updatesSiblingSharedScore     = 20
		updatedBySiblingBaseScore     = 10
		updatedBySiblingSharedScore   = 20
	)

	for _, rfc := range r.Values() {
		for _, rfcId := range rfc.Updates {
			relationScore := r.getRelationScore(rfcId, rfc.DocId, true)
			if relationScore == nil {
				continue
			}
			relationScore.updates = updatesBaseScore + updatesSharedScore/len(r.Find(rfcId).UpdatedBy)
		}

		for _, rfcId := range rfc.Obsoletes {
			relationScore := r.getRelationScore(rfcId, rfc.DocId, true)
			if relationScore == nil {
				continue
			}
			relationScore.obsoletes = obsoletesScore
		}

		if len(rfc.Obsoletes) > 1 {
			for _, rfcIdA := range rfc.Obsoletes {
				for _, rfcIdB := range rfc.Obsoletes {
					if rfcIdB <= rfcIdA {
						continue
					}
					relationScore := r.getRelationScore(rfcIdA, rfcIdB, true)
					if relationScore == nil {
						continue
					}
					relationScore.obsoletesSibling = obsoletesSiblingBaseScore + obsoletesSiblingSharedScore/len(rfc.Obsoletes)
				}
			}
		}

		if len(rfc.ObsoletedBy) > 1 {
			for _, rfcIdA := range rfc.ObsoletedBy {
				for _, rfcIdB := range rfc.ObsoletedBy {
					if rfcIdB <= rfcIdA {
						continue
					}
					relationScore := r.getRelationScore(rfcIdA, rfcIdB, true)
					if relationScore == nil {
						continue
					}
					relationScore.obsoletedBySibling = obsoletedBySiblingBaseScore + obsoletedBySiblingSharedScore/len(rfc.ObsoletedBy)
				}
			}
		}

		if len(rfc.Updates) > 1 {
			for _, rfcIdA := range rfc.Updates {
				for _, rfcIdB := range rfc.Updates {
					if rfcIdB <= rfcIdA {
						continue
					}
					relationScore := r.getRelationScore(rfcIdA, rfcIdB, true)
					if relationScore == nil {
						continue
					}
					relationScore.updatesSibling = updatesSiblingBaseScore + updatesSiblingSharedScore/len(rfc.Updates)
				}
			}
		}

		if len(rfc.UpdatedBy) > 1 {
			for _, rfcIdA := range rfc.UpdatedBy {
				for _, rfcIdB := range rfc.UpdatedBy {
					if rfcIdB <= rfcIdA {
						continue
					}
					relationScore := r.getRelationScore(rfcIdA, rfcIdB, true)
					if relationScore == nil {
						continue
					}
					relationScore.updatedBySibling = updatedBySiblingBaseScore + updatedBySiblingSharedScore/len(rfc.UpdatedBy)
				}
			}
		}
	}
}

func (r *RfcIndex) getRelationScore(rfcIdA, rfcIdB string, createIfNotExists bool) *RelationScore {
	rfcA := r.Find(rfcIdA)
	rfcB := r.Find(rfcIdB)
	if rfcA == nil || rfcB == nil {
		return nil
	}
	var elderRfcId, youngerRfcId string
	if rfcIdA < rfcIdB {
		elderRfcId = rfcIdA
		youngerRfcId = rfcIdB
	} else {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}
	if rfcB.PubDateInMonth < rfcA.PubDateInMonth {
		elderRfcId = rfcIdB
		youngerRfcId = rfcIdA
	}

	relationScoreMap, ok := r.relationScoreMap[elderRfcId]
	if !ok {
		if createIfNotExists {
			relationScoreMap = make(map[string]*RelationScore)
			r.relationScoreMap[elderRfcId] = relationScoreMap
		} else {
			return nil
		}
	}
	relationScore, ok := relationScoreMap[youngerRfcId]
	if !ok {
		if createIfNotExists {
			relationScore = &RelationScore{}
			relationScoreMap[youngerRfcId] = relationScore
		} else {
			return nil
		}
	}
	return relationScore
}

func (r *RfcIndex) GetRelationScore(rfcIdA, rfcIdB string) *RelationScore {
	return r.getRelationScore(rfcIdA, rfcIdB, false)
}

func (r *RfcIndex) getDescendantYRange(rfcId string) int {
	relationScoreMap, ok := r.relationScoreMap[rfcId]
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

type RelationScore struct {
	obsoletes          int
	updates            int
	obsoletesSibling   int
	obsoletedBySibling int
	updatesSibling     int
	updatedBySibling   int
}

func (r *RelationScore) Value() int {
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
	)

	l := len(record)
	groups := make([]string, 0)

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
