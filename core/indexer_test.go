package core

import (
	"testing"

	"github.com/go-ego/riot/types"
	"github.com/vcaesar/tt"
)

func TestAddKeywords(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    1,
		Keywords: []types.KeywordIndex{{"token1", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    2,
		Keywords: []types.KeywordIndex{{"token2", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    3,
		Keywords: []types.KeywordIndex{{"token3", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    7,
		Keywords: []types.KeywordIndex{{"token7", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    1,
		Keywords: []types.KeywordIndex{{"token2", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    7,
		Keywords: []types.KeywordIndex{{"token77", 0, []int{}}},
	}, false)
	indexer.AddDocToCache(nil, true)

	tt.Expect(t, "", indicesToString(&indexer, "token1"))
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "3 ", indicesToString(&indexer, "token3"))
	tt.Expect(t, "7 ", indicesToString(&indexer, "token77"))
}

func TestRemoveDocToCache(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})

	// doc1 = "token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, true)
	tt.Expect(t, "2 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 ", indicesToString(&indexer, "token3"))

	indexer.RemoveDocToCache(2, false)
	// doc1 = "token1 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, true)
	tt.Expect(t, "1 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 ", indicesToString(&indexer, "token3"))

	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
			{"token3", 0, []int{14}},
		},
	}, true)
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "2 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token3"))

	// doc3 = "token1 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 3,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, true)
	indexer.RemoveDocToCache(3, true)
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "2 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token3"))

	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, true)
	// doc3 = "token1 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 3,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, true)
	tt.Expect(t, "1 3 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "2 3 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 2 ", indicesToString(&indexer, "token3"))
}

func TestLookupLocsIndex(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})
	// doc1 = "token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
			{"token3", 0, []int{14}},
		},
	}, false)
	// doc3 = "token1 token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 3,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, false)
	// doc4 = "token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 4,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
		},
	}, false)
	// doc7 = "token1 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 7,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc9 = "token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 9,
		Keywords: []types.KeywordIndex{
			{"token3", 0, []int{0}},
		},
	}, true)

	tt.Expect(t, "2 3 7 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "1 2 3 4 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 2 7 9 ", indicesToString(&indexer, "token3"))

	tt.Expect(t, "", indexedDocsToString(indexer.Lookup(
		[]string{"token4"}, []string{}, nil, false)))

	tt.Expect(t, "[7 0 [0]] [3 0 [0]] [2 0 [0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1"}, []string{}, nil, false)))
	tt.Expect(t, "", indexedDocsToString(indexer.Lookup(
		[]string{"token1", "token4"}, []string{}, nil, false)))

	tt.Expect(t, "[3 1 [0 7]] [2 1 [0 7]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token2"}, []string{}, nil, false)))
	tt.Expect(t, "[3 13 [7 0]] [2 13 [7 0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token1"}, []string{}, nil, false)))
	tt.Expect(t, "[7 1 [0 7]] [2 8 [0 14]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[7 13 [7 0]] [2 20 [14 0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token1"}, []string{}, nil, false)))
	tt.Expect(t, "[2 1 [7 14]] [1 1 [0 7]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[2 13 [14 7]] [1 13 [7 0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token2"}, []string{}, nil, false)))

	tt.Expect(t, "[2 2 [0 7 14]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token2", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[2 26 [14 7 0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token2", "token1"}, []string{}, nil, false)))
}

func TestLookupDocIdsIndex(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.DocIdsIndex})
	// doc1 = "token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
			{"token3", 0, []int{14}},
		},
	}, false)
	// doc3 = "token1 token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 3,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, false)
	// doc4 = "token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 4,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
		},
	}, false)
	// doc7 = "token1 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 7,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc9 = "token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 9,
		Keywords: []types.KeywordIndex{
			{"token3", 0, []int{0}},
		},
	}, true)

	tt.Expect(t, "2 3 7 ", indicesToString(&indexer, "token1"))
	tt.Expect(t, "1 2 3 4 ", indicesToString(&indexer, "token2"))
	tt.Expect(t, "1 2 7 9 ", indicesToString(&indexer, "token3"))

	tt.Expect(t, "", indexedDocsToString(indexer.Lookup(
		[]string{"token4"}, []string{}, nil, false)))

	tt.Expect(t, "[7 0 []] [3 0 []] [2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1"}, []string{}, nil, false)))
	tt.Expect(t, "", indexedDocsToString(indexer.Lookup(
		[]string{"token1", "token4"}, []string{}, nil, false)))

	tt.Expect(t, "[3 0 []] [2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token2"}, []string{}, nil, false)))
	tt.Expect(t, "[3 0 []] [2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token1"}, []string{}, nil, false)))
	tt.Expect(t, "[7 0 []] [2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[7 0 []] [2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token1"}, []string{}, nil, false)))
	tt.Expect(t, "[2 0 []] [1 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[2 0 []] [1 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token2"}, []string{}, nil, false)))

	tt.Expect(t, "[2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token1", "token2", "token3"}, []string{}, nil, false)))
	tt.Expect(t, "[2 0 []] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token3", "token2", "token1"}, []string{}, nil, false)))
}

func TestLookupWithProximity(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})

	// doc1 = "token2 token4 token4 token2 token3 token4"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0, 21}},
			{"token3", 0, []int{28}},
			{"token4", 0, []int{7, 14, 35}},
		},
	}, true)
	tt.Expect(t, "[1 1 [21 28]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token3"}, []string{}, nil, false)))

	// doc1 = "t2 t1 . . . t2 t3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"t1", 0, []int{3}},
			{"t2", 0, []int{0, 12}},
			{"t3", 0, []int{15}},
		},
	}, true)
	tt.Expect(t, "[1 8 [3 12 15]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"t1", "t2", "t3"}, []string{}, nil, false)))

	// doc1 = "t3 t2 t1 . . . . . t2 t3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"t1", 0, []int{6}},
			{"t2", 0, []int{3, 19}},
			{"t3", 0, []int{0, 22}},
		},
	}, true)
	tt.Expect(t, "[1 10 [6 3 0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"t1", "t2", "t3"}, []string{}, nil, false)))
}

func TestLookupWithPartialLocations(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})
	// doc1 = "token2 token4 token4 token2 token3 token4" + "label1"(不在文本中)
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0, 21}},
			{"token3", 0, []int{28}},
			{"label1", 0, []int{}},
			{"token4", 0, []int{7, 14, 35}},
		},
	}, false)
	// doc2 = "token2 token4 token4 token2 token3 token4"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0, 21}},
			{"token3", 0, []int{28}},
			{"token4", 0, []int{7, 14, 35}},
		},
	}, true)

	tt.Expect(t, "1 ", indicesToString(&indexer, "label1"))

	tt.Expect(t, "[1 1 [21 28]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2", "token3"}, []string{"label1"}, nil, false)))
}

func TestLookupWithBM25(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{
		IndexType: types.FrequenciesIndex,
		BM25Parameters: &types.BM25Parameters{
			K1: 1,
			B:  1,
		},
	})
	// doc1 = "token2 token4 token4 token2 token3 token4"
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    1,
		TokenLen: 6,
		Keywords: []types.KeywordIndex{
			{"token2", 3, []int{0, 21}},
			{"token3", 7, []int{28}},
			{"token4", 15, []int{7, 14, 35}},
		},
	}, false)
	// doc2 = "token6 token7"
	indexer.AddDocToCache(&types.DocIndex{
		DocId:    2,
		TokenLen: 2,
		Keywords: []types.KeywordIndex{
			{"token6", 3, []int{0}},
			{"token7", 15, []int{7}},
		},
	}, true)

	outputs, _ := indexer.Lookup([]string{"token2", "token3", "token4"},
		[]string{}, nil, false)

	// BM25 = log2(3) * (12/9 + 28/17 + 60/33) = 6.3433
	tt.Expect(t, "76055", int(outputs[0].BM25*10000))
}

func TestLookupWithinDocIds(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})
	// doc1 = "token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
			{"token3", 0, []int{7}},
		},
	}, false)
	// doc2 = "token1 token2 token3"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
			{"token3", 0, []int{14}},
		},
	}, false)
	// doc3 = "token1 token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 3,
		Keywords: []types.KeywordIndex{
			{"token1", 0, []int{0}},
			{"token2", 0, []int{7}},
		},
	}, false)
	// doc4 = "token2"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 4,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0}},
		},
	}, true)

	docIds := make(map[uint64]bool)
	docIds[1] = true
	docIds[3] = true
	tt.Expect(t, "[3 0 [7]] [1 0 [0]] ",
		indexedDocsToString(indexer.Lookup(
			[]string{"token2"}, []string{}, docIds, false)))
}

func TestLookupWithLocations(t *testing.T) {
	var indexer Indexer
	indexer.Init(types.IndexerOpts{IndexType: types.LocsIndex})
	// doc1 = "token2 token4 token4 token2 token3 token4"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 1,
		Keywords: []types.KeywordIndex{
			{"token2", 0, []int{0, 21}},
			{"token3", 0, []int{28}},
			{"token4", 0, []int{7, 14, 35}},
		},
	}, true)

	// doc2 = "token2 token4 token4 token2 token3 token4"
	indexer.AddDocToCache(&types.DocIndex{
		DocId: 2,
		Keywords: []types.KeywordIndex{
			{"token3", 0, []int{0, 21}},
			{"token5", 0, []int{28}},
			{"token2", 0, []int{7, 14, 35}},
		},
	}, true)

	indexer.RemoveDocToCache(2, true)
	docs, _ := indexer.Lookup(
		[]string{"token2", "token3"}, []string{}, nil, false)
	tt.Expect(t, "[[0 21] [28]]", docs[0].TokenLocs)
}
