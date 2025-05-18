package cliutil

import (
	"encoding/csv"
	"os"
	"slices"
	"strings"

	"github.com/ttkzw/rfctree"
)

func GetTargets(filename string, docIds []string) ([]*rfctree.Target, error) {
	targets := make([]*rfctree.Target, 0, len(docIds))
	for _, docId := range docIds {
		target, err := rfctree.NewTarget([]string{docId})
		if err == rfctree.ErrInvalidCSVRecord {
			continue
		}
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}

	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}

		c := csv.NewReader(f)
		records, err := c.ReadAll()
		if err != nil {
			return nil, err
		}

		targetsFromFile := make([]*rfctree.Target, 0, len(records))
		for _, record := range records {
			target, err := rfctree.NewTarget(record)
			if err == rfctree.ErrInvalidCSVRecord {
				continue
			}
			if err != nil {
				return nil, err
			}
			targetsFromFile = append(targetsFromFile, target)
		}
		targets = append(targets, targetsFromFile...)
	}
	return targets, nil
}

func GetKeywords(filename string, keywords []string) ([]string, error) {
	if filename != "" {
		keywordsFromFile, err := readLineFromFile(filename)
		if err != nil {
			return nil, err
		}
		keywords = append(keywords, keywordsFromFile...)
	}
	keywords = slices.DeleteFunc(keywords, func(s string) bool {
		return s == "" || strings.HasPrefix(s, "#")
	})
	return keywords, nil
}

func GetExcludes(filename string, docIds []string) ([]string, error) {
	excludes := docIds
	if filename != "" {
		excludesFromFile, err := readLineFromFile(filename)
		if err != nil {
			return nil, err
		}
		excludes = append(excludes, excludesFromFile...)
	}
	excludes = slices.DeleteFunc(excludes, func(s string) bool {
		return !strings.HasPrefix(s, "RFC")
	})
	return excludes, nil
}

func readLineFromFile(name string) ([]string, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(data), "\n"), nil
}
