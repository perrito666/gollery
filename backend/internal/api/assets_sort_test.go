package api

import (
	"testing"
	"time"

	"github.com/perrito666/gollery/backend/internal/domain"
)

func TestSortAssets_Filename(t *testing.T) {
	assets := []domain.Asset{
		{ID: "b", Filename: "b.jpg"},
		{ID: "a", Filename: "a.jpg"},
		{ID: "c", Filename: "c.jpg"},
	}
	sortAssets(assets, "filename")
	want := []string{"a", "b", "c"}
	for i, w := range want {
		if assets[i].ID != w {
			t.Errorf("[%d] id = %q, want %q", i, assets[i].ID, w)
		}
	}
}

func TestSortAssets_Date_UsesModTime(t *testing.T) {
	early := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mid := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	late := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	assets := []domain.Asset{
		{ID: "mid", Filename: "b.jpg", ModTime: mid},
		{ID: "late", Filename: "a.jpg", ModTime: late},
		{ID: "early", Filename: "c.jpg", ModTime: early},
	}
	sortAssets(assets, "date")
	want := []string{"early", "mid", "late"}
	for i, w := range want {
		if assets[i].ID != w {
			t.Errorf("[%d] id = %q, want %q", i, assets[i].ID, w)
		}
	}
}

func TestSortAssets_DateTaken_PrefersEXIFOverModTime(t *testing.T) {
	// Two assets with inverted ModTime vs EXIF DateTaken order: sort
	// must follow EXIF DateTaken, ignoring ModTime.
	shotEarly := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	shotLate := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	mtimeSwapped := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	assets := []domain.Asset{
		{
			ID:       "shot-late",
			Filename: "a.jpg",
			ModTime:  time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			Metadata: &domain.ImageMetadata{DateTaken: &shotLate},
		},
		{
			ID:       "shot-early",
			Filename: "b.jpg",
			ModTime:  mtimeSwapped, // mtime later than shot-late's shot time
			Metadata: &domain.ImageMetadata{DateTaken: &shotEarly},
		},
	}
	sortAssets(assets, "date_taken")
	if assets[0].ID != "shot-early" {
		t.Fatalf("expected shot-early first, got %q", assets[0].ID)
	}
	if assets[1].ID != "shot-late" {
		t.Fatalf("expected shot-late second, got %q", assets[1].ID)
	}
}

func TestSortAssets_DateTaken_FallsBackToModTime(t *testing.T) {
	// One asset has EXIF DateTaken, one does not. The one without falls
	// back to its ModTime and mixes into the ordering correctly.
	shot := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	earlyMod := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	lateMod := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)

	assets := []domain.Asset{
		{ID: "with-exif", Filename: "b.jpg", ModTime: lateMod,
			Metadata: &domain.ImageMetadata{DateTaken: &shot}},
		{ID: "no-exif-early", Filename: "a.jpg", ModTime: earlyMod},
		{ID: "no-exif-late", Filename: "c.jpg", ModTime: lateMod},
	}
	sortAssets(assets, "date_taken")
	want := []string{"no-exif-early", "with-exif", "no-exif-late"}
	for i, w := range want {
		if assets[i].ID != w {
			t.Errorf("[%d] id = %q, want %q", i, assets[i].ID, w)
		}
	}
}

func TestSortAssets_DateTaken_TieBreakerByFilename(t *testing.T) {
	// Same DateTaken should sort by filename for determinism.
	shot := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	assets := []domain.Asset{
		{ID: "b", Filename: "b.jpg", Metadata: &domain.ImageMetadata{DateTaken: &shot}},
		{ID: "a", Filename: "a.jpg", Metadata: &domain.ImageMetadata{DateTaken: &shot}},
	}
	sortAssets(assets, "date_taken")
	if assets[0].ID != "a" || assets[1].ID != "b" {
		t.Errorf("tie-broken order = [%s,%s], want [a,b]", assets[0].ID, assets[1].ID)
	}
}
