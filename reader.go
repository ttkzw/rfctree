package rfctree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReadRfcJsonFiles(rfcsDir string) (map[string]*Rfc, error) {
	files, err := os.ReadDir(rfcsDir)
	if err != nil {
		return nil, err
	}

	rfcMap := make(map[string]*Rfc, 10000)
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

		rfc, err := NewRfc(data)
		if err == ErrRfcNotIssued {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %v", file.Name(), err.Error())
		}

		rfcMap[rfc.DocId] = rfc
	}
	return rfcMap, nil
}
