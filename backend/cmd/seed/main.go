package main

import (
	"backend/internal/data/model"
	"backend/internal/pkg/hash"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// --- Gemini API types ---

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// --- Seed data ---

var categories = []struct {
	Name string
	Slug string
}{
	{"音樂", "music"},
	{"遊戲", "gaming"},
	{"教育", "education"},
	{"娛樂", "entertainment"},
	{"科技", "technology"},
	{"運動", "sports"},
	{"新聞", "news"},
	{"美食", "food"},
	{"旅遊", "travel"},
	{"生活", "lifestyle"},
}

var tags = []struct {
	Name string
	Slug string
}{
	{"搞笑", "funny"},
	{"教學", "tutorial"},
	{"Vlog", "vlog"},
	{"開箱", "unboxing"},
	{"直播精華", "stream-highlight"},
	{"音樂MV", "music-video"},
	{"遊戲實況", "gameplay"},
	{"美食料理", "cooking"},
	{"旅行紀錄", "travel-log"},
	{"科技評測", "tech-review"},
	{"新手入門", "beginner"},
	{"健身運動", "fitness"},
	{"動畫", "animation"},
	{"訪談", "interview"},
	{"DIY手作", "diy"},
}

var creators = []struct {
	Username    string
	DisplayName string
	Password    string
}{
	{"creator_alice", "Alice Chen", "password123"},
	{"creator_bob", "Bob Wang", "password123"},
	{"creator_cindy", "Cindy Liu", "password123"},
	{"creator_david", "David Lin", "password123"},
	{"creator_emma", "Emma Huang", "password123"},
}

func main() {
	// Load .env from project root
	if err := godotenv.Load("../../.env"); err != nil {
		// Try current directory
		if err := godotenv.Load(".env"); err != nil {
			log.Println("Warning: .env file not found, using environment variables")
		}
	}

	geminiKey := os.Getenv("GEMINI_KEY")
	if geminiKey == "" {
		log.Fatal("GEMINI_KEY not set in .env or environment")
	}

	// Connect to MySQL (same DSN as config.yaml)
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/fenzvideo?parseTime=True&loc=Local&charset=utf8mb4"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(
		&model.User{},
		&model.Channel{},
		&model.Category{},
		&model.Video{},
		&model.Tag{},
		&model.UserTagPreference{},
		&model.Membership{},
		&model.ViewRecord{},
		&model.Notification{},
		&model.Donation{},
	); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
	log.Println("Database migrated successfully")

	// Step 1: Seed admin user
	seedAdmin(db)

	// Step 2: Seed categories
	categoryModels := seedCategories(db)

	// Step 3: Seed tags
	tagModels := seedTags(db)

	// Step 4: Seed creator users + channels
	userModels := seedCreators(db)

	// Step 5: Generate videos using Gemini API (one per tag)
	seedVideos(db, geminiKey, userModels, categoryModels, tagModels)

	log.Println("Seed completed successfully!")
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&count)
	if count > 0 {
		log.Println("Admin user already exists, skipping")
		return
	}

	hashed, _ := hash.HashPassword("admin123")
	admin := model.User{
		Username:    "admin",
		DisplayName: "System Admin",
		Password:    hashed,
		Role:        "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		log.Printf("Failed to create admin: %v", err)
		return
	}

	// Create admin channel
	db.Create(&model.Channel{UserID: admin.ID})
	log.Println("Admin user created")
}

func seedCategories(db *gorm.DB) []model.Category {
	var existing []model.Category
	db.Find(&existing)
	if len(existing) > 0 {
		log.Printf("Categories already exist (%d), skipping", len(existing))
		return existing
	}

	var models []model.Category
	for _, c := range categories {
		m := model.Category{Name: c.Name, Slug: c.Slug}
		db.Create(&m)
		models = append(models, m)
	}
	log.Printf("Seeded %d categories", len(models))
	return models
}

func seedTags(db *gorm.DB) []model.Tag {
	var existing []model.Tag
	db.Find(&existing)
	if len(existing) > 0 {
		log.Printf("Tags already exist (%d), skipping", len(existing))
		return existing
	}

	var models []model.Tag
	for _, t := range tags {
		m := model.Tag{Name: t.Name, Slug: t.Slug}
		db.Create(&m)
		models = append(models, m)
	}
	log.Printf("Seeded %d tags", len(models))
	return models
}

func seedCreators(db *gorm.DB) []model.User {
	var existing []model.User
	db.Where("role = ? AND username LIKE ?", "user", "creator_%").Find(&existing)
	if len(existing) >= len(creators) {
		log.Printf("Creator users already exist (%d), skipping", len(existing))
		return existing
	}

	var models []model.User
	for _, c := range creators {
		var count int64
		db.Model(&model.User{}).Where("username = ?", c.Username).Count(&count)
		if count > 0 {
			var u model.User
			db.Where("username = ?", c.Username).First(&u)
			models = append(models, u)
			continue
		}

		hashed, _ := hash.HashPassword(c.Password)
		u := model.User{
			Username:    c.Username,
			DisplayName: c.DisplayName,
			Password:    hashed,
			Role:        "user",
		}
		db.Create(&u)

		// Create channel for creator
		db.Create(&model.Channel{
			UserID:     u.ID,
			MonthlyFee: float64(rand.Intn(10) + 1),
		})
		models = append(models, u)
	}
	log.Printf("Seeded %d creator users with channels", len(models))
	return models
}

func seedVideos(db *gorm.DB, geminiKey string, users []model.User, cats []model.Category, tagModels []model.Tag) {
	var videoCount int64
	db.Model(&model.Video{}).Count(&videoCount)
	if videoCount > 0 {
		log.Printf("Videos already exist (%d), skipping", videoCount)
		return
	}

	log.Printf("Generating %d videos via Gemini API (one per tag)...", len(tagModels))

	for i, tag := range tagModels {
		creator := users[i%len(users)]
		cat := cats[i%len(cats)]

		// Call Gemini API to generate video title and description
		title, desc := generateVideoContent(geminiKey, tag.Name, cat.Name)
		if title == "" {
			title = fmt.Sprintf("%s - Sample Video %d", tag.Name, i+1)
			d := fmt.Sprintf("This is a sample video about %s in category %s", tag.Name, cat.Name)
			desc = d
		}

		duration := uint32(rand.Intn(600) + 60) // 1-11 minutes
		viewsMember := uint64(rand.Intn(5000))
		viewsNonMember := uint64(rand.Intn(10000))

		video := model.Video{
			UserID:         creator.ID,
			CategoryID:     cat.ID,
			Title:          title,
			Description:    &desc,
			VideoURL:       fmt.Sprintf("/fenzvideo/videos/sample_%d.mp4", i+1),
			Duration:       duration,
			ViewsMember:    viewsMember,
			ViewsNonMember: viewsNonMember,
			AccessTier:     0, // public
			IsPublished:    true,
			IsHidden:       false,
		}

		if err := db.Create(&video).Error; err != nil {
			log.Printf("Failed to create video for tag %s: %v", tag.Name, err)
			continue
		}

		// Associate tag with video
		db.Exec("INSERT INTO video_tags (video_id, tag_id) VALUES (?, ?)", video.ID, tag.ID)

		log.Printf("[%d/%d] Created video: %s (tag: %s, creator: %s)",
			i+1, len(tagModels), title, tag.Name, creator.DisplayName)

		// Rate limit Gemini API calls
		time.Sleep(1 * time.Second)
	}

	log.Printf("Seeded %d videos", len(tagModels))
}

func generateVideoContent(apiKey, tagName, categoryName string) (title, description string) {
	prompt := fmt.Sprintf(
		`Generate a creative video title and description for a video streaming platform.
The video is tagged as "%s" and belongs to the "%s" category.
The content should be in Traditional Chinese (繁體中文).

Respond in exactly this JSON format (no markdown, no code blocks):
{"title": "video title here (max 50 chars)", "description": "video description here (2-3 sentences, max 200 chars)"}`,
		tagName, categoryName,
	)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s",
		apiKey,
	)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Gemini API request failed: %v", err)
		return "", ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
		return "", ""
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("Failed to parse Gemini response: %v", err)
		return "", ""
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		log.Println("Gemini returned empty response")
		return "", ""
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	// Strip markdown code block if present
	rawText = strings.TrimSpace(rawText)
	rawText = strings.TrimPrefix(rawText, "```json")
	rawText = strings.TrimPrefix(rawText, "```")
	rawText = strings.TrimSuffix(rawText, "```")
	rawText = strings.TrimSpace(rawText)

	var result struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		log.Printf("Failed to parse Gemini JSON output: %v (raw: %s)", err, rawText)
		return "", ""
	}

	return result.Title, result.Description
}
