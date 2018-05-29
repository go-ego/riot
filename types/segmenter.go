package types

import "github.com/go-ego/gse"

type Segmenter interface {
	// Segment 对文本分词
	//
	// 输入参数：
	//	bytes	UTF8 文本的字节数组
	//
	// 输出：
	//	[]Segment	划分的分词
	Segment(bytes []byte) []gse.Segment

	// ModeSegment segment using search mode if searchMode is true
	ModeSegment(bytes []byte, searchMode ...bool) []gse.Segment
}
