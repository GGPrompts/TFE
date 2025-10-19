# Analyze TFE Performance

You are analyzing the performance of TFE to identify bottlenecks and optimization opportunities.

## Your Task

Use Desktop Commander and analysis tools to:

1. **Profile the codebase**
   - Search for potential performance issues:
     - Unnecessary file reads in loops
     - Inefficient string operations
     - Missing caching
     - Redundant calculations

2. **Analyze specific modules**
   - `file_operations.go`: File loading efficiency
   - `render_*.go`: Rendering performance
   - `update_*.go`: Event handling efficiency
   - `helpers.go`: Utility function optimization

3. **Look for common issues**
   - O(n²) operations
   - Repeated file I/O
   - Unnecessary allocations
   - Missing early returns
   - Inefficient string concatenation

4. **Run benchmarks (if they exist)**
   - Execute `go test -bench=.`
   - Analyze benchmark results
   - Compare with previous runs

5. **Memory profiling**
   - Look for potential memory leaks
   - Check for large allocations
   - Identify unnecessary copies

6. **Suggest optimizations**
   - Prioritize by impact (high/medium/low)
   - Provide code examples for fixes
   - Estimate performance improvement
   - List any trade-offs

## Report Format

```
🔍 TFE Performance Analysis

📊 High Priority Issues:
  1. loadFiles() reads directory twice - file_operations.go:123
     • Impact: 2x slower directory loading
     • Fix: Cache results in model
     • Estimated improvement: 50% faster

  2. renderPreview() rebuilds styles every frame - render_preview.go:89
     • Impact: Unnecessary CPU usage
     • Fix: Move styles to global constants
     • Estimated improvement: 20% faster rendering

📊 Medium Priority Issues:
  ...

💡 Quick Wins:
  - Add early return in getCurrentFile() if cursor out of bounds
  - Cache getFileIcon() results
  - Use strings.Builder instead of += in formatFileSize()
```

Start the performance analysis now and provide a detailed report.
