package rfctree

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
)

func ListKeyword(rfcsDir string, keywords []string) error {
	rfcIndex, err := NewRfcIndex(rfcsDir)
	if err != nil {
		return err
	}

	keywordMap := make(map[string]int, len(keywords))
	for _, rfc := range rfcIndex.FindAllByKeywords(keywords) {
		for _, keyword := range rfc.Keywords {
			if keyword == "" {
				continue
			}
			if strings.HasPrefix(keyword, "[--------") {
				continue
			}
			n, ok := keywordMap[keyword]
			if ok {
				keywordMap[keyword] = n + 1
			} else {
				keywordMap[keyword] = 1
			}
		}
	}

	sortedKeyword := slices.Sorted(maps.Keys(keywordMap))
	slices.Reverse(sortedKeyword)
	slices.SortStableFunc(sortedKeyword, func(a, b string) int {
		return cmp.Compare(keywordMap[a], keywordMap[b])
	})
	slices.Reverse(sortedKeyword)
	for _, keyword := range sortedKeyword {
		fmt.Printf("%7d %s\n", keywordMap[keyword], keyword)
	}
	return nil
}
