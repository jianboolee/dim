package service

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestResizeCoverReturnsRequestedSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 80, 40))

	resized := resizeCover(src, 24, 24)
	if resized.Bounds().Dx() != 24 || resized.Bounds().Dy() != 24 {
		t.Fatalf("expected 24x24 image, got %dx%d", resized.Bounds().Dx(), resized.Bounds().Dy())
	}
}

func TestFallbackAvatarIsDeterministic(t *testing.T) {
	first := fallbackAvatar("user_a", 16)
	second := fallbackAvatar("user_a", 16)

	if first.At(0, 0) != second.At(0, 0) {
		t.Fatal("expected fallback avatar color to be deterministic")
	}
}

func TestGroupAvatarPublicURL(t *testing.T) {
	groupID := primitive.NewObjectID()
	service := NewGroupAvatarService(nil, nil, nil, GroupAvatarOptions{
		PublicBaseURL: "https://im.example.com/",
	})

	got := service.publicURL(groupID, "123.png")
	want := "https://im.example.com/im/api/public/group-avatars/" + groupID.Hex() + "/123.png"
	if got != want {
		t.Fatalf("unexpected public url: got %q, want %q", got, want)
	}
}

func TestCleanupOldVersionsKeepsNewestFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"100.png", "200.png", "300.png", "note.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	service := NewGroupAvatarService(nil, nil, nil, GroupAvatarOptions{KeepVersions: 2})
	service.cleanupOldVersions(dir)

	if _, err := os.Stat(filepath.Join(dir, "100.png")); !os.IsNotExist(err) {
		t.Fatalf("expected oldest png to be removed, stat err: %v", err)
	}
	for _, name := range []string{"200.png", "300.png", "note.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected %s to remain: %v", name, err)
		}
	}
}
