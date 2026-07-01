package service

import (
	"context"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"d-im/internal/models"
	"d-im/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	groupAvatarRoutePrefix = "/im/api/public/group-avatars"
	maxGroupAvatarMembers  = 9
	maxAvatarDownloadBytes = 5 << 20
)

type GroupAvatarOptions struct {
	Directory     string
	PublicBaseURL string
	Size          int
	KeepVersions  int
}

type GroupAvatarService struct {
	groupRepo     *repository.GroupRepository
	memberRepo    *repository.GroupMemberRepository
	userRepo      *repository.UserRepository
	directory     string
	publicBaseURL string
	size          int
	keepVersions  int
	httpClient    *http.Client
}

func NewGroupAvatarService(
	groupRepo *repository.GroupRepository,
	memberRepo *repository.GroupMemberRepository,
	userRepo *repository.UserRepository,
	options GroupAvatarOptions,
) *GroupAvatarService {
	size := options.Size
	if size <= 0 {
		size = 240
	}
	keepVersions := options.KeepVersions
	if keepVersions <= 0 {
		keepVersions = 3
	}
	directory := strings.TrimSpace(options.Directory)
	if directory == "" {
		directory = "./storage/group-avatars"
	}
	return &GroupAvatarService{
		groupRepo:     groupRepo,
		memberRepo:    memberRepo,
		userRepo:      userRepo,
		directory:     directory,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(options.PublicBaseURL), "/"),
		size:          size,
		keepVersions:  keepVersions,
		httpClient:    &http.Client{Timeout: 4 * time.Second},
	}
}

func (s *GroupAvatarService) Directory() string {
	return s.directory
}

func (s *GroupAvatarService) Regenerate(ctx context.Context, groupID primitive.ObjectID) error {
	if s == nil {
		return nil
	}
	members, err := s.memberRepo.ListActiveByGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}

	limit := len(members)
	if limit > maxGroupAvatarMembers {
		limit = maxGroupAvatarMembers
	}
	members = members[:limit]

	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}
	usersByID := map[string]*models.User{}
	if s.userRepo != nil {
		if users, err := s.userRepo.FindByIDs(ctx, userIDs); err == nil {
			usersByID = users
		}
	}

	img := s.render(ctx, members, usersByID)
	version := time.Now().UnixNano()
	groupDir := filepath.Join(s.directory, groupID.Hex())
	if err := os.MkdirAll(groupDir, 0o755); err != nil {
		return fmt.Errorf("create group avatar directory: %w", err)
	}

	filename := strconv.FormatInt(version, 10) + ".png"
	filePath := filepath.Join(groupDir, filename)
	tempPath := filePath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("create group avatar file: %w", err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		_ = os.Remove(tempPath)
		return fmt.Errorf("encode group avatar: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("close group avatar file: %w", err)
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("save group avatar file: %w", err)
	}

	avatarURL := s.publicURL(groupID, filename)
	now := time.Now()
	if err := s.groupRepo.SetAvatar(ctx, groupID, avatarURL, filePath, version, now); err != nil {
		return err
	}
	s.cleanupOldVersions(groupDir)
	return nil
}

func (s *GroupAvatarService) render(ctx context.Context, members []*models.GroupMember, usersByID map[string]*models.User) *image.RGBA {
	size := s.size
	canvas := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.RGBA{R: 238, G: 242, B: 247, A: 255}), image.Point{}, draw.Src)

	grid := 2
	if len(members) > 4 {
		grid = 3
	}
	gap := size / 60
	if gap < 3 {
		gap = 3
	}
	cellSize := (size - gap*(grid+1)) / grid

	for i, member := range members {
		if i >= grid*grid {
			break
		}
		user := usersByID[member.UserID]
		tile := s.loadUserAvatar(ctx, user, member.UserID, cellSize)
		x := gap + (i%grid)*(cellSize+gap)
		y := gap + (i/grid)*(cellSize+gap)
		draw.Draw(canvas, image.Rect(x, y, x+cellSize, y+cellSize), tile, image.Point{}, draw.Over)
	}
	return canvas
}

func (s *GroupAvatarService) loadUserAvatar(ctx context.Context, user *models.User, userID string, size int) image.Image {
	if user != nil && strings.TrimSpace(user.Avatar) != "" {
		if img, err := s.downloadAvatar(ctx, user.Avatar); err == nil && img != nil {
			return resizeCover(img, size, size)
		}
	}
	key := userID
	if user != nil && user.Nickname != "" {
		key = user.Nickname
	}
	return fallbackAvatar(key, size)
}

func (s *GroupAvatarService) downloadAvatar(ctx context.Context, url string) (image.Image, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("unsupported avatar url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("avatar response status %d", resp.StatusCode)
	}
	img, _, err := image.Decode(io.LimitReader(resp.Body, maxAvatarDownloadBytes))
	return img, err
}

func (s *GroupAvatarService) publicURL(groupID primitive.ObjectID, filename string) string {
	path := groupAvatarRoutePrefix + "/" + groupID.Hex() + "/" + filename
	if s.publicBaseURL == "" {
		return path
	}
	return s.publicBaseURL + path
}

func (s *GroupAvatarService) cleanupOldVersions(groupDir string) {
	entries, err := os.ReadDir(groupDir)
	if err != nil {
		return
	}
	type versionFile struct {
		name    string
		version int64
	}
	files := make([]versionFile, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".png" {
			continue
		}
		version, err := strconv.ParseInt(strings.TrimSuffix(entry.Name(), ".png"), 10, 64)
		if err != nil {
			continue
		}
		files = append(files, versionFile{name: entry.Name(), version: version})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].version > files[j].version
	})
	for _, file := range files[safeKeepCount(s.keepVersions, len(files)):] {
		_ = os.Remove(filepath.Join(groupDir, file.name))
	}
}

func safeKeepCount(keepVersions int, total int) int {
	if keepVersions < 0 {
		keepVersions = 0
	}
	if keepVersions > total {
		return total
	}
	return keepVersions
}

func resizeCover(src image.Image, width int, height int) *image.RGBA {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	if srcW <= 0 || srcH <= 0 || width <= 0 || height <= 0 {
		return dst
	}

	scaleX := float64(width) / float64(srcW)
	scaleY := float64(height) / float64(srcH)
	scale := scaleX
	if scaleY > scale {
		scale = scaleY
	}
	cropW := float64(width) / scale
	cropH := float64(height) / scale
	offsetX := float64(bounds.Min.X) + (float64(srcW)-cropW)/2
	offsetY := float64(bounds.Min.Y) + (float64(srcH)-cropH)/2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(offsetX + float64(x)/scale)
			srcY := int(offsetY + float64(y)/scale)
			if srcX < bounds.Min.X {
				srcX = bounds.Min.X
			}
			if srcY < bounds.Min.Y {
				srcY = bounds.Min.Y
			}
			if srcX >= bounds.Max.X {
				srcX = bounds.Max.X - 1
			}
			if srcY >= bounds.Max.Y {
				srcY = bounds.Max.Y - 1
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func fallbackAvatar(key string, size int) *image.RGBA {
	palette := []color.RGBA{
		{R: 85, G: 125, B: 235, A: 255},
		{R: 63, G: 171, B: 134, A: 255},
		{R: 240, G: 150, B: 73, A: 255},
		{R: 218, G: 91, B: 116, A: 255},
		{R: 113, G: 93, B: 196, A: 255},
		{R: 78, G: 156, B: 197, A: 255},
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	bg := palette[int(hash.Sum32())%len(palette)]
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	return img
}
