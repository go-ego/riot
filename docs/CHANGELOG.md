# CHANGELOG

## riot v0.10.0, Danube River

### Add  

- [NEW] Add heartbeat
- [NEW] Add go dependency package to vendor
- [NEW] Add codecov
- [NEW] Add context cancel
- [NEW] Update gse and support Japanese
- [NEW] Update  gse 

    Load dictionary and add default dict
    
    Add different dict and adapt to different languages

- [NEW] Update gse, segmentation rules and docs
- [NEW] Add get the engine tokens func
- [NEW] Add more test: TestSearchJp, TestSearchLogic, TestSearchGse
- [NEW] English docs, more docs and examples 

    Add English docs and fix docs

    Add word segmentation rules docs

    [Look at docs](https://github.com/go-ego/riot/tree/master/docs)

    Store example

    Logic search example

    Pinyin weibo search example

    Pinyin search example

    Different dict and language search example
    
    [Look at more Examples](https://github.com/go-ego/riot/tree/master/examples)

### Update

- [NEW] Update README.md
- [NEW] Update examples
- [NEW] Move engine to riot
- [NEW] Update circle.yml and .travis.yml
- [NEW] Update badger to 1.10 and store name
- [NEW] Rename some flie
- [NEW] Update engine and gse log print
- [NEW] Update data riot log print
- [NEW] Remove old and unused code

### Fix

- [FIX] Format some code
- [FIX] Fix heartbeat port
- [FIX] Fix docs error, such as: link
- [FIX] Update badger to 1.10 fix api 
- [FIX] Add godoc, fix golint and remove something
- [FIX] Fix examples segmentation error
- [FIX] Fix pinyin segmenter slice and example
- [FIX] Fix and update .github template
- [FIX] Adjust frequency and update load dictionary, fix gse segmentater and other


See Commits for more details, after Oct 23. Before Oct 23 used to be the base code.