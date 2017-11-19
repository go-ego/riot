## Dynamically modify the index table

The riot engine supports searching while adding an index (engine.IndexDocument function), but as indexes are indexed and write-locked when added, search performance decreases as indexes are added. Please control the frequency of additions or move a large number of add-ons to engines when they are idle. Deleting a document (engine.RemoveDocument function) has the same problem.

The riot engine supports cache insert and delete index operations, bulk insert and delete documents to improve performance. The delete operation also supports custom rating fields that delete the document from the sorter.
