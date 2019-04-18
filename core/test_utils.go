package core

import (
	"fmt"

	"github.com/oGre222/tea/types"
)

func indicesToString(indexer *Indexer, token string) (output string) {
	if indices, ok := indexer.tableLock.table[token]; ok {
		for i := 0; i < indexer.getIndexLen(indices); i++ {
			output += fmt.Sprintf("%s ", indexer.getDocId(indices, i))
		}
	}
	return
}

func indexedDocsToString(docs []types.IndexedDoc, numDocs int) (output string) {
	for _, doc := range docs {
		output += fmt.Sprintf("[%s %d %v] ",
			doc.DocId, doc.TokenProximity, doc.TokenSnippetLocs)
	}
	return
}

func scoredDocsToString(docs []types.ScoredDoc) (output string) {
	for _, doc := range docs {
		output += fmt.Sprintf("[%s [", doc.DocId)
		for _, score := range doc.Scores {
			output += fmt.Sprintf("%d ", int(score*1000))
		}
		output += "]] "
	}
	return
}

func indexedDocIdsToString(docs []types.IndexedDoc, numDocs int) (output string) {
	for _, doc := range docs {
		output += fmt.Sprintf("[%s] ",
			doc.DocId)
	}
	return
}
