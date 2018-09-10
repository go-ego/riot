Key words close distance（Token Proximity）
===

The closeness of keywords is used to measure whether multiple keywords are in the same document or not. For example, if the user searches for the phrase "International football," including the keywords "international" and "football", the two adjacent keywords appear in a document immediately before and after the same order, Two words in the middle of a lot of words are close to the larger distance. Close proximity is a way of measuring the relevance of a document and multiple keywords. Close proximity should not be the only indicator of document ordering, but in some cases a fair amount of extraneous results can be filtered out by setting thresholds.

The nearest neighbor distance of N keyword is calculated as follows:

Assuming that the first byte of the i-th keyword appears in the text as P_i and length L_i, the immediate distance is:

  ArgMin(Sum(Abs(P_(i+1) - P_i - L_i)))

The specific calculation process is to take a P_1 first, calculate the smallest value of Abs (P_2 - P_1 - L1) among the possible values of all P_2, and then select P3, P4, etc. according to the same method after fixing P2. Traverse all possible P_1 to get the minimum value.

See the computeTokenProximity function in [core / indexer.go] (/ core / indexer.go) for implementation.

Close distance calculation need to save each word position in the indexer, which requires additional memory consumption, it is off by default, open this function, please set EngineOpts.IndexerOpts.IndexType LocsIndex initialization engine.
