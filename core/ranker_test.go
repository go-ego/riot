package core

import (
	"reflect"
	"testing"

	"github.com/go-ego/riot/types"
	"github.com/go-ego/riot/utils"
)

type DummyScoringFields struct {
	label   string
	counter int
	amount  float32
}

type DummyScoringCriteria struct {
	Threshold float32
}

func (criteria DummyScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	if reflect.TypeOf(fields) == reflect.TypeOf(DummyScoringFields{}) {
		dsf := fields.(DummyScoringFields)
		value := float32(dsf.counter) + dsf.amount
		if value < criteria.Threshold {
			return []float32{}
		}
		return []float32{value}
	}
	return []float32{}
}

type Attri struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

func TestRankDocument(t *testing.T) {
	var ranker Ranker
	attri := Attri{Title: "title", Author: "who"}

	ranker.Init()
	ranker.AddDoc(1, DummyScoringFields{}, "content", attri)
	ranker.AddDoc(3, DummyScoringFields{}, "content", attri)
	ranker.AddDoc(4, DummyScoringFields{}, "content", attri)

	scoredDocs, _ := ranker.Rank([]types.IndexedDocument{
		{DocId: 1, BM25: 6},
		{DocId: 3, BM25: 24},
		{DocId: 4, BM25: 18},
	}, types.RankOptions{ScoringCriteria: types.RankByBM25{}}, false)
	utils.Expect(t, "[3 [24000 ]] [4 [18000 ]] [1 [6000 ]] ", scoredDocsToString(scoredDocs))

	scoredDocs, _ = ranker.Rank([]types.IndexedDocument{
		{DocId: 1, BM25: 6},
		{DocId: 3, BM25: 24},
		{DocId: 2, BM25: 0},
		{DocId: 4, BM25: 18},
	}, types.RankOptions{ScoringCriteria: types.RankByBM25{}, ReverseOrder: true}, false)
	// doc0因为没有AddDoc所以没有添加进来
	utils.Expect(t, "[1 [6000 ]] [4 [18000 ]] [3 [24000 ]] ", scoredDocsToString(scoredDocs))
}

func TestRankWithCriteria(t *testing.T) {
	var ranker Ranker
	attri := Attri{Title: "title", Author: "who"}

	ranker.Init()
	ranker.AddDoc(1, DummyScoringFields{
		label:   "label3",
		counter: 3,
		amount:  22.3,
	}, "content", attri)
	ranker.AddDoc(2, DummyScoringFields{
		label:   "label4",
		counter: 1,
		amount:  2,
	}, "content", attri)
	ranker.AddDoc(3, DummyScoringFields{
		label:   "label1",
		counter: 7,
		amount:  10.3,
	}, "content", attri)
	ranker.AddDoc(4, DummyScoringFields{
		label:   "label1",
		counter: -1,
		amount:  2.3,
	}, "content", attri)

	criteria := DummyScoringCriteria{}
	scoredDocs, _ := ranker.Rank([]types.IndexedDocument{
		{DocId: 1, TokenProximity: 6},
		{DocId: 2, TokenProximity: -1},
		{DocId: 3, TokenProximity: 24},
		{DocId: 4, TokenProximity: 18},
	}, types.RankOptions{ScoringCriteria: criteria}, false)
	utils.Expect(t, "[1 [25300 ]] [3 [17300 ]] [2 [3000 ]] [4 [1300 ]] ", scoredDocsToString(scoredDocs))

	criteria.Threshold = 4
	scoredDocs, _ = ranker.Rank([]types.IndexedDocument{
		{DocId: 1, TokenProximity: 6},
		{DocId: 2, TokenProximity: -1},
		{DocId: 3, TokenProximity: 24},
		{DocId: 4, TokenProximity: 18},
	}, types.RankOptions{ScoringCriteria: criteria}, false)
	utils.Expect(t, "[1 [25300 ]] [3 [17300 ]] ", scoredDocsToString(scoredDocs))
}

func TestRemoveDoc(t *testing.T) {
	var ranker Ranker
	attri := Attri{Title: "title", Author: "who"}

	ranker.Init()
	ranker.AddDoc(1, DummyScoringFields{
		label:   "label3",
		counter: 3,
		amount:  22.3,
	}, "content", attri)
	ranker.AddDoc(2, DummyScoringFields{
		label:   "label4",
		counter: 1,
		amount:  2,
	}, "content", attri)
	ranker.AddDoc(3, DummyScoringFields{
		label:   "label1",
		counter: 7,
		amount:  10.3,
	}, "content", attri)
	ranker.RemoveDoc(3)

	criteria := DummyScoringCriteria{}
	scoredDocs, _ := ranker.Rank([]types.IndexedDocument{
		{DocId: 1, TokenProximity: 6},
		{DocId: 2, TokenProximity: -1},
		{DocId: 3, TokenProximity: 24},
		{DocId: 4, TokenProximity: 18},
	}, types.RankOptions{ScoringCriteria: criteria}, false)
	utils.Expect(t, "[1 [25300 ]] [2 [3000 ]] ", scoredDocsToString(scoredDocs))
}
