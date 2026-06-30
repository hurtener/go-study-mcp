// Package jobs provides an in-memory, concurrency-safe registry of
// asynchronous audio-generation jobs.
//
// Long-running LLM + TTS work returns a job handle immediately and runs in a
// background goroutine; the UI polls the registry (via the list_jobs tool) to
// render progress and, once complete, plays the generated audio.
package jobs

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Status is the lifecycle state of a Job.
type Status string

const (
	// StatusQueued is the initial state before the worker goroutine starts.
	StatusQueued Status = "queued"
	// StatusProcessing means the worker is generating script and/or audio.
	StatusProcessing Status = "processing"
	// StatusCompleted means the audio file was written successfully.
	StatusCompleted Status = "completed"
	// StatusFailed means generation errored; Error holds the reason.
	StatusFailed Status = "failed"
)

// Job is a single asynchronous generation job.
type Job struct {
	ID               string
	Kind             string // podcast, study_guide, flashcards, synthesize
	Title            string
	Status           Status
	CreatedAt        time.Time
	UpdatedAt        time.Time
	OutputPath       string
	DurationEstimate string
	CharacterCount   int
	WordCount        int
	Error            string
}

// Registry is a concurrency-safe store of jobs.
type Registry struct {
	mu   sync.RWMutex
	jobs map[string]*Job
	seq  int
}

// NewRegistry returns an empty registry ready for use.
func NewRegistry() *Registry {
	return &Registry{jobs: make(map[string]*Job)}
}

// Add creates a new queued job and returns a snapshot copy.
func (r *Registry) Add(kind, title string) Job {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	now := time.Now()
	id := fmt.Sprintf("job-%d-%03d", now.UnixNano(), r.seq)
	j := &Job{
		ID:        id,
		Kind:      kind,
		Title:     title,
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.jobs[id] = j
	return *j
}

func (r *Registry) update(id string, fn func(*Job)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	j, ok := r.jobs[id]
	if !ok {
		return
	}
	fn(j)
	j.UpdatedAt = time.Now()
}

// SetProcessing transitions a job to the processing state.
func (r *Registry) SetProcessing(id string) {
	r.update(id, func(j *Job) { j.Status = StatusProcessing })
}

// Complete records a successful result on the job.
func (r *Registry) Complete(id, outputPath, durationEstimate string, charCount, wordCount int) {
	r.update(id, func(j *Job) {
		j.Status = StatusCompleted
		j.OutputPath = outputPath
		j.DurationEstimate = durationEstimate
		j.CharacterCount = charCount
		j.WordCount = wordCount
	})
}

// Fail records a terminal failure with a human-readable message.
func (r *Registry) Fail(id, msg string) {
	r.update(id, func(j *Job) {
		j.Status = StatusFailed
		j.Error = msg
	})
}

// Get returns a snapshot copy of a job by ID.
func (r *Registry) Get(id string) (Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	j, ok := r.jobs[id]
	if !ok {
		return Job{}, false
	}
	return *j, true
}

// List returns all jobs, newest first.
func (r *Registry) List() []Job {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Job, 0, len(r.jobs))
	for _, j := range r.jobs {
		out = append(out, *j)
	}
	sort.Slice(out, func(i, k int) bool {
		return out[i].CreatedAt.After(out[k].CreatedAt)
	})
	return out
}
