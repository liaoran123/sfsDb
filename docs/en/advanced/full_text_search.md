# Full-Text Search

## 5.1 Full-Text Index Principles

sfsDb implements native research-grade full-text indexing functionality, using the **sliding window tokenization algorithm** (also known as n-gram tokenization algorithm) for text segmentation.

### Sliding Window Tokenization Algorithm

The principle of the sliding window tokenization algorithm is as follows:
1. Use a sliding window of length `ftlen`
2. The window starts from the beginning of the string and moves one character backward each time
3. Each time the window moves, extract the substring within the window as a token
4. This process continues until the window slides to the end of the string

### Custom Tokenization Algorithm

If you need to use other tokenization algorithms, simply override the `DefaultFullTextIndex.Tokenize` method:

```go
func (dfi *DefaultFullTextIndex) Tokenize(nr string, ftlen int) (tokens []string) {
    // Custom tokenization implementation
    // ...
    return tokens
}
```

### Full-Text Index Length

The length of the full-text index is controlled by the `ftlen` field, which is a property in the `DefaultFullTextIndex` struct:

```go
type DefaultFullTextIndex struct {
    BaseIndex // Embed base index
    // Full-text index split field
    ftsplit string
    // Split length
    ftlen int
}
```

The `ftlen` parameter determines the size of the sliding window, which is the maximum length of each token.

## 5.2 Creating Full-Text Index

```go
// Create full-text index
fullTextIndex, err := engine.DefaultFullTextIndexNew("fulltext_desc")
if err != nil {
    panic(err)
}
// Add fields, the last field must be the primary key
fullTextIndex.AddFields("description", "id")
// Set full-text index field and length
fullTextIndex.SetFullField("description", 5) // 5 represents the length of the full-text index
err = table.CreateIndex(fullTextIndex)
if err != nil {
    panic(err)
}
```

## 5.2 Inserting Records with Full-Text Index

```go
// Insert records with full-text index
contentRecords := []map[string]any{
    {"id": 1, "name": "Product 1", "description": "This is a high-performance laptop suitable for programming and gaming"},
    {"id": 2, "name": "Product 2", "description": "Smartphone with powerful camera and long-lasting battery"},
    {"id": 3, "name": "Product 3", "description": "Wireless headphones providing immersive audio experience"},
    {"id": 4, "name": "Product 4", "description": "Smartwatch that can monitor health data and receive notifications"},
}

for _, record := range contentRecords {
    _, err = table.Insert(&record)
    if err != nil {
        panic(err)
    }
}
```

## 5.3 Performing Full-Text Search

```go
// Full-text search example
fmt.Println("=== Full-Text Search Example ===")

// Search for records containing "laptop"
fmt.Println("\n1. Search for 'laptop':")
search1 := map[string]any{"description": "laptop"}
iter1, _ := table.Search(&search1)   
defer iter1.Release()
records1 := iter1.GetRecords(true)
defer records1.Release()   
for _, record := range records1 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// Search for records containing "smart"
fmt.Println("\n2. Search for 'smart':")
search2 := map[string]any{"description": "smart"}
iter2, _ := table.Search(&search2)   
defer iter2.Release()
records2 := iter2.GetRecords(true)
defer records2.Release()   
for _, record := range records2 {
    fmt.Printf("   - %s: %s\n", record["name"], record["description"])
}

// 3. Search result field selection example
fmt.Println("\n3. Search result field selection:")
search3 := map[string]any{"description": "smart"}
iter3, _ := table.Search(&search3)   
defer iter3.Release()
records3 := iter3.GetRecords(true)
defer records3.Release()   
// Use Select method to only select name field
selectedNames := records3.Select("name")
fmt.Println("Only display names of matching records:")
for _, record := range selectedNames {
    fmt.Printf("   - %s\n", record)
}
```