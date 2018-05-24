## Word segmentation rules:

```Go
types.EngineOpts{
	Using:         4,
	GseMode: true,
}
```
- When Using is 0 and the content participle is not empty, the keyword is obtained from the content through the gse participle and Tokens.

- When Using is 1 and the Content participle is not empty, the keyword is obtained from the content by the gse participle.

- When Using is 2, Using is 1 or 3 and the Content participle is empty; the key word is obtained from Tokens.

- When Using is 3 and the Content participle is not empty, the keyword is obtained from the content through the gse participle and through the `""` participle.

- When Using is 4 and the content participle is not empty, the keyword is obtained from the content through the space participle.

- When GseMode is true, the Gse participle uses a search pattern to give search engines as many keywords as possible.