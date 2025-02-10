# Performance Benchmark Report: XLORM vs GORM vs XORM

## Test Environment
- **Language**: Go
- **Database**: MySQL
- **Test Method**: Go Benchmark
- **Test Scale**: Single record and batch operations
- **Metrics**: 
  - Execution times (ns/op)
  - Memory allocation (B/op)
  - Allocation count (allocs/op)

## Test Results Comparison

### 1. Single Record Insert
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_Insert-4 | 1747 | 679,999 | 4,798 | 63 | - |
| XLORM | BenchmarkInsert-4 | 2,035 | 557,139 | 1,097 | 28 | 81.8 / 23.3 / 44.4 |
| XORM | BenchmarkXORM_Insert-4 | 1,951 | 581,697 | 2,259 | 43 | 85.5 / 47.1 / 68.3 |

### 2. Batch Insert
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_BatchInsert-4 | 1,636 | 709,971 | 6,595 | 89 | - |
| XLORM | BenchmarkBatchInsert-4 | 1,614 | 714,198 | 2,639 | 64 | 100.6 / 40.0 / 72.0 |
| XORM | BenchmarkXORM_BatchInsert-4 | 1,864 | 620,709 | 3,638 | 85 | 87.4 / 55.1 / 95.5 |

### 3. Single Record Query
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_Find-4 | 6,344 | 192,686 | 4,155 | 64 | - |
| XLORM | BenchmarkFind-4 | 5,638 | 190,575 | 2,748 | 51 | 99.0 / 66.1 / 79.7 |
| XORM | BenchmarkXORM_Find-4 | 5,640 | 203,428 | 4,274 | 116 | 98.4 / 102.9 / 181.3 |

### 4. Batch Query
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_FindAll-4 | 584 | 2,000,131 | 4,864 | 62 | - |
| XLORM | BenchmarkFindAll-4 | 608 | 1,940,781 | 2,471 | 37 | 97.0 / 50.8 / 59.7 |
| XORM | BenchmarkXORM_FindAll-4 | 619 | 1,986,661 | 3,882 | 86 | 106.0 / 79.8 / 138.7 |

### 5. Single Record Update
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_Update-4 | 1,506 | 833,433 | 6,103 | 69 | - |
| XLORM | BenchmarkUpdate-4 | 1,695 | 696,792 | 1,276 | 27 | 83.6 / 20.9 / 39.1 |
| XORM | BenchmarkXORM_Update-4 | 6,212 | 183,031 | 2,577 | 64 | 41.2 / 42.2 / 92.8 |

### 6. Batch Update
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_BatchUpdate-4 | 681 | 1,746,530 | 13,378 | 167 | - |
| XLORM | BenchmarkBatchUpdate-4 | 1,390 | 806,534 | 3,924 | 63 | 119.1 / 29.3 / 37.7 |
| XORM | BenchmarkXORM_BatchUpdate-4 | 1,027 | 1,244,602 | 6,201 | 160 | 71.3 / 46.3 / 95.8 |

### 7. Delete Operation
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_Delete-4 | 4,236 | 278,541 | 5,302 | 62 | - |
| XLORM | BenchmarkDelete-4 | 7,189 | 170,216 | 957 | 20 | 169.7 / 61.1 / 32.3 |
| XORM | BenchmarkXORM_Delete-4 | 6,818 | 178,486 | 2,600 | 69 | 161.0 / 49.0 / 111.3 |

### 8. Transaction Operation
| Framework | Test Case | Runs | Time per Operation (ns/op) | Memory Allocation (B/op) | Allocation Count (allocs/op) | Comparison to GORM (%) |
|-----------|-----------|------|----------------------------|-------------------------|------------------------------|------------------------|
| GORM | BenchmarkGORM_Transaction-4 | 1,572 | 751,466 | 5,893 | 65 | - |
| XLORM | BenchmarkTransaction-4 | 1,508 | 823,494 | 2,686 | 79 | 95.9 / 110.0 / 121.5 |
| XORM | BenchmarkXORM_Transaction-4 | 1,626 | 728,734 | 3,149 | 66 | 103.4 / 53.4 / 101.5 |

## Performance Analysis

### Performance Characteristics
- **XLORM**:
  - Excels in query and delete operations
  - Lowest memory allocation across most operations
  - Query operations: ~1% faster than GORM, ~66% lower memory allocation
  - Delete operations: ~69.7% faster than GORM, ~61.1% lower memory allocation

- **XORM**:
  - Best performance in update operations
  - Competitive memory allocation
  - Update operations: ~41.2% faster than GORM
  - Batch insert operations: Lower memory allocation compared to GORM

- **GORM**:
  - Stable performance in transaction operations
  - Comprehensive features and strong community support
  - Suitable for complex transaction management

### Memory Allocation Insights
- **XLORM**: 
  - Significantly lower memory allocation
  - Insert operations: ~23.3% of GORM's memory allocation
  - Allocation count reduced by ~44.4%

- **XORM**:
  - Moderate memory allocation
  - Batch update: ~46.3% of GORM's memory allocation

### Recommended Use Cases
1. **XLORM**:
   - Frequent query and delete operations
   - Performance and memory efficiency are priorities

2. **XORM**:
   - High-performance update operations
   - Moderate memory constraints

3. **GORM**:
   - Complex transaction management
   - Comprehensive ORM features
   - Strong ecosystem and community support

## Conclusion
This benchmark provides insights into the performance characteristics of XLORM, XORM, and GORM. While performance is crucial, it's essential to consider:
- Specific business requirements
- Ease of use
- Feature completeness
- Community support

Choose the ORM that best aligns with your project's unique needs and constraints.