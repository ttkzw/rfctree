package rfctree

import (
	"encoding/json"
	"strings"
)

type RfcDoc struct {
	Draft       string   `json:"draft"`
	DocId       string   `json:"doc_id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Format      []string `json:"format"`
	PageCount   string   `json:"page_count"`
	PubStatus   string   `json:"pub_status"`
	Status      string   `json:"status"`
	Source      string   `json:"source"`
	Abstract    string   `json:"abstract"`
	PubDate     string   `json:"pub_date"`
	Keywords    []string `json:"keywords"`
	Obsoletes   []string `json:"obsoletes"`
	ObsoletedBy []string `json:"obsoleted_by"`
	Updates     []string `json:"updates"`
	UpdatedBy   []string `json:"updated_by"`
	SeeAlso     []string `json:"see_also"`
	Doi         string   `json:"doi"`
	ErrataUrl   string   `json:"errata_url"`
}

func NewRfcDoc(data []byte) (*RfcDoc, error) {
	var rfc RfcDoc
	err := json.Unmarshal(data, &rfc)
	if err != nil {
		return nil, err
	}

	if rfc.PubStatus == "NOT ISSUED" {
		return nil, ErrRfcNotIssued
	}

	rfc.Draft = strings.TrimSpace(rfc.Draft)
	rfc.DocId = strings.TrimSpace(rfc.DocId)
	rfc.Title = strings.TrimSpace(rfc.Title)
	rfc.Abstract = strings.TrimSpace(rfc.Abstract)
	rfc.Source = strings.TrimSpace(rfc.Source)
	rfc.Keywords = trimSpaceFromSlice(rfc.Keywords)
	rfc.Obsoletes = trimSpaceFromSlice(rfc.Obsoletes)
	rfc.ObsoletedBy = trimSpaceFromSlice(rfc.ObsoletedBy)
	rfc.Updates = trimSpaceFromSlice(rfc.Updates)
	rfc.UpdatedBy = trimSpaceFromSlice(rfc.UpdatedBy)
	rfc.SeeAlso = trimSpaceFromSlice(rfc.SeeAlso)

	return &rfc, nil
}

func trimSpaceFromSlice(s []string) []string {
	trimmed := make([]string, 0, len(s))
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		trimmed = append(trimmed, v)
	}
	return trimmed
}
