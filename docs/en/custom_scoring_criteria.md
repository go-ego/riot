Custom rating fields and rating rules
===

The riot engine supports saving some rating fields in the sequencer memory and scoring documents using custom rating rules. Example:

```go
// Custom rating field
type MyScoringFields struct {
	// Add some sort of document data, can be any type, such as:
	label string
	counter int32
	someScore float32
}

// MyScoringCriteria implements the types.ScoringCriteria interface, which is the Score function below
type MyScoringCriteria struct {
}
func (criteria MyScoringCriteria) Score(
	doc types.IndexedDoc, fields interface{}) []float32 {
	// First check if the scoring field is of type MyScoringFields, if not, an empty slice is returned and the document is removed from the result
	if reflect.TypeOf(fields) != reflect.TypeOf(MySearchFields{}) {
		return []float32{}
	}
	
	// Match is type conversion
	myFields := fields.(MySearchFields)
	
	// The following uses the data in myFields to rate the document and return the score
}
```

Document MyScoringFields data through riot.Engine Index function passed to the sequencer stored in memory. Then through the Search function parameters call MyScoringCriteria query.

Of course, the Score function of MyScoringCriteria can also read more document data from the hard disk or database through docId for scoring, but the speed is much slower than directly reading from memory. Please choose between memory and speed.

[examples/weibo/custom_scoring_criteria.go](/examples/weibo/custom_scoring_criteria.go)   contains an example of using custom rules to query Weibo data.
