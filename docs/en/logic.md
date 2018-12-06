# logic search

```Go
docs := searcher.Search(types.SearchReq{
		Text: "google testing",
		// Search "google testing" segmentation `or relation` 
		// and not the result of "accidentally"
        // 搜索 "google testing" 分词 `或关系` 并且不是 "accidentally" 的结果
		Logic: types.Logic{
			Should: true,
			// LogicExpr: types.LogicExpr{
			// 	NotInLabels: "accidentally",
			// },
			Expr: types.Expr{
				NotIn: "accidentally",
			},
        },
        // 0 to 100 of the results
		RankOpts: &types.RankOpts{
			OutputOffset: 0,
			MaxOutputs:   100,
		}})
```