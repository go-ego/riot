package riot

import (
	"log"

	"github.com/go-ego/riot/types"
)

type ScoringFields struct {
	A, B, C float32
}

var (
	text1 = "Hello world, 你好世界!"
	text2 = "在路上, in the way"

	textJP  = "こんにちは世界"
	textJP1 = "こんにちは世界, こんにちは"
	reqText = "World人口"

	reqG = types.SearchReq{Text: "Google"}
	Req1 = types.SearchReq{Text: reqText}

	score1   = ScoringFields{1, 2, 3}
	score091 = ScoringFields{0, 9, 1}

	inxOpts = &types.IndexerOpts{
		IndexType: types.LocsIndex,
	}
)

var (
	rankOptsMax1 = rankOptsMax(0, 1)

	rankOptsMax10      = rankOptsOrder(false)
	rankOptsMax10Order = rankOptsOrder(true)

	rankOptsMax3 = rankOptsMax(1, 3)
)

func makeDocIds() map[string]bool {
	docIds := make(map[string]bool)
	docIds["5"] = true
	docIds["3"] = true
	docIds["1"] = true
	docIds["2"] = true

	return docIds
}

type RankByTokenProximity struct {
}

func (rule RankByTokenProximity) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	if doc.TokenProximity < 0 {
		return []float32{}
	}
	return []float32{1.0 / (float32(doc.TokenProximity) + 1)}
}

func OrderlessOpts(idOnly bool) types.EngineOpts {
	return types.EngineOpts{
		Using:   1,
		IDOnly:  idOnly,
		GseDict: "./testdata/test_dict.txt",
	}
}

func rankEngineOpts(rankOpts types.RankOpts) types.EngineOpts {
	return types.EngineOpts{
		Using:       1,
		GseDict:     "./testdata/test_dict.txt",
		DefRankOpts: &rankOpts,
		IndexerOpts: inxOpts,
	}
}

func rankOptsOrder(order bool) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    order,
		OutputOffset:    0,
		MaxOutputs:      10,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

func rankOptsMax(output, max int) types.RankOpts {
	return types.RankOpts{
		ReverseOrder:    true,
		OutputOffset:    output,
		MaxOutputs:      max,
		ScoringCriteria: &RankByTokenProximity{},
	}
}

var (
	TestIndexOpts = rankEngineOpts(rankOptsMax10)

	orderOpts = rankEngineOpts(rankOptsMax10Order)
)

func AddDocs(engine *Engine) {
	// docId := uint64(1)
	engine.Index("1", types.DocData{
		Content: "The world, 有七十亿人口人口",
		Fields:  score1,
	})

	// docId++
	engine.IndexDoc("2", types.DocIndexData{
		Content: "The world, 人口",
		Fields:  nil,
	})

	engine.Index("3", types.DocData{
		Content: "The world",
		Fields:  nil,
	})

	engine.Index("4", types.DocData{
		Content: "有人口",
		Fields:  ScoringFields{2, 3, 1},
	})

	engine.Index("5", types.DocData{
		Content: "The world, 七十亿人口",
		Fields:  score091,
	})

	engine.Index("6", types.DocData{
		Content: "有七十亿人口",
		Fields:  ScoringFields{2, 3, 3},
	})

	engine.Flush()
}

func AddDocsWithLabels(engine *Engine) {
	// docId := uint64(1)
	engine.Index("1", types.DocData{
		Content: "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄",
		Labels:  []string{"复仇者", "战争"},
	})
	log.Println("engine.Segment(): ",
		engine.Segment("《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"))

	// docId++
	engine.Index("2", types.DocData{
		Content: "在IMAX影院放映时",
		Labels:  []string{"影院"},
	})

	engine.Index("3", types.DocData{
		Content: " Google 是世界最大搜索引擎, baidu 是最大中文的搜索引擎",
		Labels:  []string{"Google"},
	})

	engine.Index("4", types.DocData{
		Content: "Google 在研制无人汽车",
		Labels:  []string{"Google"},
	})

	engine.Index("5", types.DocData{
		Content: " GAMAF 世界五大互联网巨头, BAT 是中国互联网三巨头",
		Labels:  []string{"互联网"},
	})

	engine.Flush()
}
