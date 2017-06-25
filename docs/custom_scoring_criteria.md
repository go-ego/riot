自定义评分字段和评分规则
===

悟空引擎支持在排序器内存中保存一些评分字段，并利用自定义评分规则给文档打分。例子:

```go
// 自定义评分字段
type MyScoringFields struct {
	// 加入一些文档排序用的数据，可以是任意类型，比如：
	label string
	counter int32
	someScore float32
}

// MyScoringCriteria实现了types.ScoringCriteria接口，也就是下面的Score函数
type MyScoringCriteria struct {
}
func (criteria MyScoringCriteria) Score(
	doc types.IndexedDocument, fields interface{}) []float32 {
	// 首先检查评分字段是否为MyScoringFields类型的，如果不是则返回空切片，此文档将从结果中剔除
	if reflect.TypeOf(fields) != reflect.TypeOf(MySearchFields{}) {
		return []float32{}
	}
	
	// 匹配则进行类型转换
	myFields := fields.(MySearchFields)
	
	// 下面利用myFields中的数据给文档评分并返回分值
}
```

文档的MyScoringFields数据通过engine.Engine的IndexDocument函数传给排序器保存在内存中。然后通过Search函数的参数调用MyScoringCriteria进行查询。

当然，MyScoringCriteria的Score函数也可以通过docId从硬盘或数据库读取更多文档数据用于打分，但速度要比从内存中直接读慢许多，请在内存和速度之间合适取舍。

[examples/custom_scoring_criteria.go](/examples/custom_scoring_criteria.go)中包含了一个利用自定义规则查询微博数据的例子。
