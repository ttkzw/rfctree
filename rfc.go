package rfctree

import (
	"cmp"
	"encoding/json"
	"errors"
	"image/color"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Status is a status of RFC.
type Status int

const (
	NOT_ISSUED = Status(iota)
	PROPOSED_STANDARD
	DRAFT_STANDARD
	INTERNET_STANDARD
	INFORMATIONAL
	EXPERIMENTAL
	HISTORIC
	BEST_CURRENT_PRACTICE
	UNKNOWN
)

var ErrRfcNotIssued = errors.New("rfc is not issued")

var statusString = [...]string{
	NOT_ISSUED:            "NOT ISSUED",
	PROPOSED_STANDARD:     "PROPOSED STANDARD",
	DRAFT_STANDARD:        "DRAFT STANDARD",
	INTERNET_STANDARD:     "INTERNET STANDARD",
	INFORMATIONAL:         "INFORMATIONAL",
	EXPERIMENTAL:          "EXPERIMENTAL",
	HISTORIC:              "HISTORIC",
	BEST_CURRENT_PRACTICE: "BEST CURRENT PRACTICE",
	UNKNOWN:               "UNKNOWN",
}

var statusDisplay = [...]string{
	NOT_ISSUED:            "NI",
	PROPOSED_STANDARD:     "PS",
	DRAFT_STANDARD:        "DS",
	INTERNET_STANDARD:     "STD",
	INFORMATIONAL:         "Informational",
	EXPERIMENTAL:          "Experimental",
	HISTORIC:              "Historic",
	BEST_CURRENT_PRACTICE: "BCP",
	UNKNOWN:               "Unknown",
}

var statusColor = [...]color.Color{
	NOT_ISSUED:            color.Black,
	PROPOSED_STANDARD:     color.RGBA{0, 0, 191, 223},
	DRAFT_STANDARD:        color.RGBA{63, 191, 191, 223},
	INTERNET_STANDARD:     color.RGBA{63, 127, 63, 223},
	INFORMATIONAL:         color.RGBA{191, 127, 63, 223},
	EXPERIMENTAL:          color.RGBA{191, 191, 63, 223},
	HISTORIC:              color.RGBA{127, 127, 127, 223},
	BEST_CURRENT_PRACTICE: color.RGBA{127, 63, 127, 223},
	UNKNOWN:               color.RGBA{255, 255, 255, 223},
}

func NewStatus(status string) Status {
	for i, s := range statusString {
		if status == s {
			return Status(i)
		}
	}
	return NOT_ISSUED
}

func (s Status) String() string {
	i := int(s)
	if 0 <= i && i < len(statusString) {
		return statusString[i]
	}
	return statusString[0]
}

func (s Status) Display() string {
	i := int(s)
	if 0 <= i && i < len(statusDisplay) {
		return statusDisplay[i]
	}
	return statusDisplay[0]
}

func (s Status) Color() color.Color {
	i := int(s)
	if 0 <= i && i < len(statusColor) {
		return statusColor[i]
	}
	return statusColor[0]
}

type Rfc struct {
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
	Metadata    Metadata `json:"-,omitempty"`
}

type Metadata struct {
	// DocIs is a document ID for display
	DocId string

	// Status is a status of RFC.
	Status Status

	// SubSeries is a sub-series of RFC.
	SubSeries string

	// PubDateInMonth is the number of months since January of the earliest publication year.
	PubDateInMonth int

	// PubYear is a publication year.
	PubYear int

	// Keywords are keywords that are normalized to uppercase.
	Keywords []string

	// IsTarget
	IsTarget bool

	// Position
	Position Position
}

type Position struct {
	X int
	Y int
}

func NewRfc(data []byte) (*Rfc, error) {
	var rfc Rfc
	err := json.Unmarshal(data, &rfc)
	if err != nil {
		return nil, err
	}

	if rfc.PubStatus == NOT_ISSUED.String() {
		return nil, ErrRfcNotIssued
	}

	rfc.Draft = strings.TrimSpace(rfc.Draft)
	rfc.DocId = strings.TrimSpace(rfc.DocId)
	rfc.Title = strings.TrimSpace(rfc.Title)
	rfc.Abstract = strings.TrimSpace(rfc.Abstract)
	rfc.Source = strings.TrimSpace(rfc.Source)
	rfc.Keywords = *trimSpaceFromSlice(&rfc.Keywords)
	rfc.Obsoletes = *trimSpaceFromSlice(&rfc.Obsoletes)
	rfc.ObsoletedBy = *trimSpaceFromSlice(&rfc.ObsoletedBy)
	rfc.Updates = *trimSpaceFromSlice(&rfc.Updates)
	rfc.UpdatedBy = *trimSpaceFromSlice(&rfc.UpdatedBy)
	rfc.SeeAlso = *trimSpaceFromSlice(&rfc.SeeAlso)

	t := toTime(rfc.PubDate)

	rfc.Metadata = Metadata{
		DocId:          toDisplayDocId(rfc.DocId),
		Status:         NewStatus(rfc.Status),
		SubSeries:      toSubSeries(rfc.SeeAlso),
		PubDateInMonth: (t.Year()-yearOfOrigin)*12 + int(t.Month()) - 1,
		PubYear:        t.Year(),
		Keywords:       normalizeKeywords(rfc.Keywords),
	}

	return &rfc, nil
}

func trimSpaceFromSlice(s *[]string) *[]string {
	trimmed := make([]string, len(*s))
	for _, v := range *s {
		trimmed = append(trimmed, strings.TrimSpace(v))
	}
	return &trimmed
}

var docIdRe = regexp.MustCompile(`^([A-Z]+)0+([0-9]+?)$`)

// toDisplayDocId converts a string like "RFC0123" to a string like "RFC 123" for display purposes.
func toDisplayDocId(docId string) string {
	return docIdRe.ReplaceAllString(docId, "$1 $2")
}

func toSubSeries(seeAlso []string) string {
	for _, ss := range []string{"STD", "BCP", "FYI"} {
		for _, sa := range seeAlso {
			if strings.HasPrefix(sa, ss) {
				return toDisplayDocId(sa)
			}
		}
	}
	return ""
}

const (
	// the year of the oldest publish date, RFC 31.
	yearOfOrigin int = 1968
)

func toTime(pubDate string) time.Time {
	var t time.Time
	var err error
	layout := "2 January 2006"
	if len(strings.Split(pubDate, " ")) == 2 {
		pubDate = "1 " + pubDate
	}

	t, err = time.Parse(layout, pubDate)
	if err != nil {
		panic(err)
	}
	return t
}

func normalizeKeywords(keywords []string) []string {
	normalized := make([]string, len(keywords))
	for _, keyword := range keywords {
		normalized = append(normalized, strings.ToUpper(keyword))
	}
	return normalized
}

func getTargetDocIds(rfcMap map[string]*Rfc, keywords []string) []string {
	var targetDocIds []string
	rfcs := slices.SortedFunc(maps.Values(rfcMap), func(a, b *Rfc) int {
		return cmp.Compare(a.DocId, b.DocId)
	})
	for _, rfc := range rfcs {
		for _, keyword := range keywords {
			if slices.Contains(rfc.Metadata.Keywords, keyword) {
				rfc.Metadata.IsTarget = true
				break
			}
		}
		if rfc.Metadata.IsTarget {
			targetDocIds = append(targetDocIds, rfc.DocId)
		}
	}
	return targetDocIds
}
