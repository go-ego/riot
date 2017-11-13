Word segmentation rules:

```Go
types.EngineInitOptions{
		Using:         4,
}
```

When Using is 1 and the content participle is not empty, the keywords are preferentially obtained from the content participle.

When Using is 2, Using is 1 or 3 and the Content participle is empty; the key word is obtained from Tokens.

When Using is 3 and the Content clause is not empty, the keyword is obtained from the content token and Tokens.

When Using is 5 and the content participle is not empty, the keyword is obtained from the content through the space participle.