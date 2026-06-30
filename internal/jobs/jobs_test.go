package jobs

import "testing"

func TestRegistryLifecycle(t *testing.T) {
	r := NewRegistry()

	job := r.Add("podcast", "Mitochondria")
	if job.Status != StatusQueued {
		t.Fatalf("new job status = %q, want %q", job.Status, StatusQueued)
	}
	if job.ID == "" {
		t.Fatal("job ID should not be empty")
	}

	r.SetProcessing(job.ID)
	if got, _ := r.Get(job.ID); got.Status != StatusProcessing {
		t.Errorf("status = %q, want %q", got.Status, StatusProcessing)
	}

	r.Complete(job.ID, "/out/a.mp3", "~5 min", 120, 800)
	got, ok := r.Get(job.ID)
	if !ok {
		t.Fatal("job missing after Complete")
	}
	if got.Status != StatusCompleted || got.OutputPath != "/out/a.mp3" {
		t.Errorf("completed job = %+v", got)
	}
	if got.WordCount != 800 || got.CharacterCount != 120 {
		t.Errorf("counts = %d words / %d chars", got.WordCount, got.CharacterCount)
	}
}

func TestRegistryFail(t *testing.T) {
	r := NewRegistry()
	job := r.Add("synthesize", "hi")
	r.Fail(job.ID, "boom")
	got, _ := r.Get(job.ID)
	if got.Status != StatusFailed || got.Error != "boom" {
		t.Errorf("failed job = %+v", got)
	}
}

func TestRegistryListNewestFirst(t *testing.T) {
	r := NewRegistry()
	first := r.Add("podcast", "one")
	second := r.Add("podcast", "two")

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	// Newest first: second was added last.
	if list[0].ID != second.ID || list[1].ID != first.ID {
		t.Errorf("order = [%s, %s], want [%s, %s]", list[0].ID, list[1].ID, second.ID, first.ID)
	}
}

func TestGetUnknown(t *testing.T) {
	r := NewRegistry()
	if _, ok := r.Get("nope"); ok {
		t.Error("Get returned ok for unknown job")
	}
}
