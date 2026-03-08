# Prompt 72 — Derivative worker pool: types and queue

Create the foundation for background image processing — a bounded worker pool with a job queue.

## Implement

1. Create `backend/internal/derive/pool.go`:

```go
type JobKind int
const (
    JobThumbnail JobKind = iota
    JobPreview
)

type Job struct {
    Kind      JobKind
    AssetID   string
    Source    string   // path to original image
    Size      int
}

type Result struct {
    Job     Job
    Path    string // output path (empty on error)
    Err     error
}

type Pool struct {
    layout   *cache.Layout
    workers  int
    jobs     chan Job
    results  chan Result
}

func NewPool(layout *cache.Layout, workers int) *Pool
func (p *Pool) Start(ctx context.Context)     // Launch worker goroutines
func (p *Pool) Submit(job Job)                // Non-blocking enqueue (drops if full)
func (p *Pool) Results() <-chan Result         // Read completed jobs
func (p *Pool) Pending() int                  // Number of jobs in queue
```

2. Implement workers that call the existing `GenerateThumbnail` / `GeneratePreview` functions.

3. Create `backend/internal/derive/pool_test.go`:
   - Test that submitting a job produces a result.
   - Test that context cancellation stops workers.
   - Test that submitting to a full queue drops the job (returns immediately).
   - Test concurrent job processing (submit N jobs, collect N results).

## Verify

```bash
cd backend && go build ./... && go vet ./... && go test ./...
```

## Do not

- Change the API layer yet (next prompt)
- Change `GenerateThumbnail` or `GeneratePreview` signatures
- Add HTTP 202 responses yet
