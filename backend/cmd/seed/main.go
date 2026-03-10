package main

import (
	"backend/internal/data/model"
	"backend/internal/pkg/hash"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

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

	// Step 5: Seed videos with pre-defined content
	seedVideos(db, userModels, categoryModels, tagModels)

	// Step 6: Upload sample video files to MinIO
	seedMinIOVideos()

	log.Println("Seed completed successfully!")
}

// sampleVideoURL is a public-domain 1080p H.264 sample video (≈2.7 MB, ~6 s).
const sampleVideoURL = "https://download.samplelib.com/mp4/sample-5s.mp4"

// downloadSampleVideo fetches a real sample video from the internet.
// Falls back to a tiny valid MP4 (ftyp atom only) if the download fails.
func downloadSampleVideo() []byte {
	log.Println("Downloading sample video from", sampleVideoURL)
	resp, err := http.Get(sampleVideoURL)
	if err != nil {
		log.Printf("Warning: failed to download sample video: %v (using minimal fallback)", err)
		return minimalMP4
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("Warning: sample video download returned %d (using minimal fallback)", resp.StatusCode)
		return minimalMP4
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Warning: failed to read sample video body: %v (using minimal fallback)", err)
		return minimalMP4
	}
	log.Printf("Downloaded sample video: %d bytes", len(data))
	return data
}

// minimalMP4 is a tiny valid ftyp box, used as a fallback if the download fails.
var minimalMP4 = []byte{
	0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70,
	0x69, 0x73, 0x6f, 0x6d, 0x00, 0x00, 0x02, 0x00,
	0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
}

func seedMinIOVideos() {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "127.0.0.1:9100"
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}
	bucket := "fenzvideo"

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Printf("Warning: Could not connect to MinIO: %v (skipping video upload)", err)
		return
	}

	ctx := context.Background()

	// Ensure bucket exists
	exists, err := mc.BucketExists(ctx, bucket)
	if err != nil {
		log.Printf("Warning: MinIO bucket check failed: %v", err)
		return
	}
	if !exists {
		if err := mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			log.Printf("Warning: Could not create MinIO bucket: %v", err)
			return
		}
	}

	// Set public-read policy
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/*"]
		}]
	}`, bucket)
	_ = mc.SetBucketPolicy(ctx, bucket, policy)

	// Download a real sample video (or fall back to minimal MP4)
	videoData := downloadSampleVideo()

	// Upload a sample video file for each seeded video (20 videos)
	uploaded := 0
	for i := 1; i <= 20; i++ {
		objectName := fmt.Sprintf("videos/sample_%d.mp4", i)

		// Skip if already exists
		_, err := mc.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
		if err == nil {
			continue // already uploaded
		}

		_, err = mc.PutObject(ctx, bucket, objectName,
			bytes.NewReader(videoData), int64(len(videoData)),
			minio.PutObjectOptions{ContentType: "video/mp4"},
		)
		if err != nil {
			log.Printf("Warning: Failed to upload %s: %v", objectName, err)
			continue
		}
		uploaded++
	}
	if uploaded > 0 {
		log.Printf("Uploaded %d sample video files to MinIO", uploaded)
	} else {
		log.Println("Sample video files already exist in MinIO, skipping")
	}

	// Upload placeholder thumbnails (simple colored JPEG images)
	thumbUploaded := 0
	colors := []string{"4A90D9", "E74C3C", "2ECC71", "F39C12", "9B59B6",
		"1ABC9C", "E67E22", "3498DB", "E91E63", "00BCD4",
		"8BC34A", "FF5722", "607D8B", "795548", "CDDC39",
		"FF9800", "673AB7", "009688", "F44336", "2196F3"}
	for i := 1; i <= 20; i++ {
		objectName := fmt.Sprintf("thumbnails/thumb_%d.jpg", i)

		_, err := mc.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
		if err == nil {
			continue
		}

		// Generate a simple colored placeholder thumbnail (1x1 pixel JPEG, browsers will stretch it)
		thumbData := generatePlaceholderJPEG(colors[(i-1)%len(colors)])
		_, err = mc.PutObject(ctx, bucket, objectName,
			bytes.NewReader(thumbData), int64(len(thumbData)),
			minio.PutObjectOptions{ContentType: "image/jpeg"},
		)
		if err != nil {
			log.Printf("Warning: Failed to upload %s: %v", objectName, err)
			continue
		}
		thumbUploaded++
	}
	if thumbUploaded > 0 {
		log.Printf("Uploaded %d placeholder thumbnails to MinIO", thumbUploaded)
	} else {
		log.Println("Thumbnails already exist in MinIO, skipping")
	}
}

// generatePlaceholderJPEG creates a minimal colored JPEG image.
func generatePlaceholderJPEG(hexColor string) []byte {
	r, g, b := parseHexColor(hexColor)
	img := image.NewRGBA(image.Rect(0, 0, 640, 360))
	for y := 0; y < 360; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return buf.Bytes()
}

func parseHexColor(hex string) (r, g, b uint8) {
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return
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

func seedVideos(db *gorm.DB, users []model.User, cats []model.Category, tagModels []model.Tag) {
	var videoCount int64
	db.Model(&model.Video{}).Count(&videoCount)
	if videoCount > 0 {
		log.Printf("Videos already exist (%d), skipping", videoCount)
		return
	}

	// Build tag lookup by slug
	tagBySlug := make(map[string]model.Tag)
	for _, t := range tagModels {
		tagBySlug[t.Slug] = t
	}
	// Build category lookup by slug
	catBySlug := make(map[string]model.Category)
	for _, c := range cats {
		catBySlug[c.Slug] = c
	}

	// 57 pre-defined videos with diverse titles, descriptions, tags, and categories
	type videoSeed struct {
		Title       string
		Description string
		Category    string   // slug
		Tags        []string // slugs
		Duration    uint32
	}

	videos := []videoSeed{
		// --- Music ---
		{"Lofi Beats to Study To 📚🎵", "A curated playlist of lofi hip hop beats perfect for late night study sessions.", "music", []string{"music-video", "tutorial"}, 3600},
		{"街頭藝人現場演奏 Jazz 經典名曲", "在台北街頭偶遇超強爵士樂手，薩克斯風演奏令人陶醉。", "music", []string{"music-video", "vlog"}, 485},
		{"吉他新手必學！10 首超簡單彈唱歌曲", "零基礎也能學會的吉他入門歌曲教學，附完整和弦譜。", "music", []string{"music-video", "tutorial", "beginner"}, 720},
		{"EDM Live Mix｜電音派對精華", "週末電音派對現場 DJ 混音精華，高能節拍嗨翻全場！", "music", []string{"music-video", "stream-highlight"}, 1800},

		// --- Gaming ---
		{"Minecraft 生存模式 Day 1 到鑽石裝全攻略", "從零開始的 Minecraft 生存之旅，手把手帶你挖到第一顆鑽石。", "gaming", []string{"gameplay", "tutorial", "beginner"}, 1250},
		{"英雄聯盟 S15 上路坦克位攻略", "新賽季上路坦克英雄全解析，出裝順序與團戰技巧大公開。", "gaming", []string{"gameplay", "tutorial"}, 890},
		{"直播精華｜史詩級五殺瞬間合集", "本週各大實況主最精彩的五殺集錦，每一幕都讓人起雞皮疙瘩！", "gaming", []string{"gameplay", "stream-highlight", "funny"}, 640},
		{"電競選手的一天 Vlog", "跟著電競職業選手體驗高強度訓練日常，看看 Pro 的真實生活。", "gaming", []string{"gameplay", "vlog"}, 920},
		{"2024 Steam 夏日特賣必買清單", "精選 20 款打骨折的高評價遊戲，錯過再等一年！", "gaming", []string{"gameplay", "unboxing", "tech-review"}, 780},

		// --- Education ---
		{"30 分鐘搞懂微積分基礎概念", "大一必修微積分不再是噩夢！用動畫圖解讓你秒懂極限與導數。", "education", []string{"tutorial", "animation", "beginner"}, 1800},
		{"英文面試必備 50 句萬用句型", "外商面試不緊張！精選 50 個高頻英文面試句型，附發音示範。", "education", []string{"tutorial", "beginner"}, 960},
		{"Python 爬蟲實戰：自動抓取股票數據", "用 Python + BeautifulSoup 寫一個股票即時爬蟲，完整程式碼教學。", "education", []string{"tutorial", "tech-review"}, 1400},
		{"動畫解說：相對論到底在說什麼？", "用簡單的動畫和比喻，帶你理解愛因斯坦最偉大的理論。", "education", []string{"animation", "tutorial"}, 620},
		{"台灣歷史 10 分鐘速成", "從荷蘭時期到現代民主化，用動畫帶你快速回顧台灣 400 年歷史。", "education", []string{"animation", "beginner"}, 600},

		// --- Entertainment ---
		{"爆笑街訪：路人的奇葩才藝大挑戰", "在西門町隨機挑戰路人秀才藝，結果笑到停不下來！", "entertainment", []string{"funny", "vlog"}, 540},
		{"密室逃脫挑戰！能在 60 分鐘內逃出嗎？", "帶團隊挑戰全台最難密室逃脫，全程高能尖叫不斷。", "entertainment", []string{"funny", "vlog", "stream-highlight"}, 1320},
		{"整人計畫：假蟑螂嚇同事的反應太爆笑", "辦公室惡作劇系列第 3 集，這次用超逼真假蟑螂整同事！", "entertainment", []string{"funny", "vlog"}, 380},
		{"深夜談話：聊聊 YouTuber 的煩惱", "凌晨三點的真心話大冒險，分享做影片背後不為人知的辛苦。", "entertainment", []string{"interview", "vlog"}, 2400},
		{"年度十大搞笑影片回顧", "2024 網路瘋傳的十大搞笑影片合集，保證看一次笑一次！", "entertainment", []string{"funny", "stream-highlight"}, 900},

		// --- Technology ---
		{"M4 MacBook Pro 深度評測：值得升級嗎？", "完整效能測試、續航實測、螢幕比較，幫你判斷該不該換。", "technology", []string{"tech-review", "unboxing"}, 1100},
		{"2024 旗艦手機大比拼：iPhone vs Samsung vs Pixel", "三大旗艦手機全方位對比評測：拍照、效能、電量一次看。", "technology", []string{"tech-review", "unboxing"}, 1500},
		{"自製智慧家庭系統 DIY 全教學", "用 Raspberry Pi + Home Assistant 打造你的智慧家庭，零基礎也能上手。", "technology", []string{"diy", "tutorial", "tech-review"}, 2100},
		{"AI 繪圖工具大對決：Midjourney vs DALL-E vs Stable Diffusion", "三大 AI 繪圖工具完整比較，同一組 prompt 生成結果差超多！", "technology", []string{"tech-review", "tutorial"}, 840},
		{"開箱最新 VR 頭盔！沉浸感超乎想像", "Meta Quest 3 開箱實測，混合實境 MR 功能真的太驚豔了。", "technology", []string{"unboxing", "tech-review"}, 720},

		// --- Sports ---
		{"居家徒手健身 30 分鐘全身訓練", "不需任何器材！跟著做完就能燃燒 300 大卡的高效居家健身。", "sports", []string{"fitness", "tutorial", "beginner"}, 1800},
		{"馬拉松備賽攻略：從 5K 到全馬訓練課表", "跑步教練分享 16 週全馬訓練計畫，新手也能完跑 42K！", "sports", []string{"fitness", "tutorial"}, 1200},
		{"極限運動合集：滑板、衝浪、跑酷精華", "2024 年度最刺激極限運動精華片段，腎上腺素飆升！", "sports", []string{"fitness", "stream-highlight"}, 480},
		{"瑜珈初學者 15 分鐘晨間伸展", "每天早上跟著做，改善體態、紓解壓力的簡單瑜珈流程。", "sports", []string{"fitness", "beginner"}, 900},
		{"NBA 本週十大好球精華", "本週 NBA 最精彩的十大好球，最後一球太不可思議了！", "sports", []string{"stream-highlight", "funny"}, 360},

		// --- News ---
		{"一週科技新聞懶人包", "本週最重要的科技新聞整理：AI 新突破、手機發表會、隱私爭議一次看。", "news", []string{"tech-review", "interview"}, 600},
		{"獨家專訪：台灣新創 CEO 談 AI 未來", "深度對談台灣最具潛力 AI 新創公司創辦人，聊產業趨勢與挑戰。", "news", []string{"interview", "tech-review"}, 1800},
		{"直播回顧：總統大選即時開票", "選舉之夜完整開票直播精華，見證歷史性的一刻。", "news", []string{"stream-highlight", "interview"}, 3600},
		{"遊戲產業年度回顧座談會", "邀請三位資深遊戲媒體人，回顧今年遊戲圈大事件。", "news", []string{"interview", "gameplay"}, 2700},

		// --- Food ---
		{"挑戰 $1000 吃遍夜市美食", "帶著一千塊走遍饒河夜市，看能吃到多少攤！", "food", []string{"cooking", "vlog", "funny"}, 960},
		{"日本拉麵之旅：東京五大名店實吃評比", "親自走訪東京最有名的五間拉麵店，從豚骨到味噌全制霸。", "food", []string{"cooking", "travel-log", "vlog"}, 1200},
		{"零失敗甜點：免烤巴斯克乳酪蛋糕", "超簡單食譜！只要攪拌放進烤箱就能做出餐廳級甜點。", "food", []string{"cooking", "tutorial", "beginner"}, 480},
		{"韓式炸雞自己做！比外賣還好吃", "在家復刻橋村炸雞的秘密醬料，酥脆多汁的完美配方。", "food", []string{"cooking", "tutorial", "diy"}, 720},
		{"美食 Youtuber 的冰箱大公開", "偷看知名美食 Youtuber 冰箱裡到底有什麼？滿滿的食材太驚人。", "food", []string{"cooking", "vlog", "funny"}, 540},
		{"台南在地人帶路：隱藏版小吃地圖", "跟著台南人吃真正的在地美食，觀光客絕對找不到的巷弄小店。", "food", []string{"cooking", "travel-log", "vlog"}, 840},

		// --- Travel ---
		{"一個人的日本自由行 Vlog｜京都篇", "獨旅京都七天六夜，嵐山竹林、伏見稻荷、錦市場全記錄。", "travel", []string{"travel-log", "vlog"}, 1500},
		{"花蓮三天兩夜行程攻略", "太魯閣、七星潭、東大門夜市⋯⋯花蓮必去景點完整規劃。", "travel", []string{"travel-log", "tutorial"}, 780},
		{"背包客挑戰｜$5000 玩遍東南亞一週", "極限省錢旅行！看我如何用五千塊玩遍泰國清邁。", "travel", []string{"travel-log", "vlog", "funny"}, 1320},
		{"露營開箱：完整裝備清單與搭帳教學", "新手露營必看！從帳篷到炊具，完整裝備開箱與搭設教學。", "travel", []string{"travel-log", "unboxing", "beginner"}, 960},
		{"空拍秘境：台灣最美 10 個無人海灘", "用空拍機帶你看台灣隱藏版絕美海灘，每一個都像仙境。", "travel", []string{"travel-log", "diy"}, 600},
		{"冰島極光之旅｜環島公路自駕遊記", "追極光、泡溫泉、冰川健行⋯⋯冰島環島 14 天全紀錄。", "travel", []string{"travel-log", "vlog"}, 2400},

		// --- Lifestyle ---
		{"極簡生活挑戰：30 天只留 100 件物品", "丟掉多餘的東西，體驗極簡生活帶來的自由與平靜。", "lifestyle", []string{"vlog", "diy"}, 840},
		{"在家打造夢想工作區 Room Tour", "居家辦公空間大改造！從 IKEA 收納到氣氛燈光全攻略。", "lifestyle", []string{"vlog", "diy", "unboxing"}, 720},
		{"一日店員體驗：咖啡廳打工實錄", "假裝自己是咖啡師！體驗一天咖啡廳店員的真實日常。", "lifestyle", []string{"vlog", "funny"}, 640},
		{"開箱人生第一台車！新手買車攻略", "分享買車心路歷程、選車比較、交車開箱全記錄。", "lifestyle", []string{"unboxing", "vlog"}, 1100},
		{"手帳控必看！超療癒 Bullet Journal 教學", "手帳排版靈感分享，從月計畫到每日記錄的完整教學。", "lifestyle", []string{"diy", "tutorial"}, 540},
		{"居家 DIY：用不到 $500 改造書架", "便宜又有質感的居家改造計畫，新手也能輕鬆完成。", "lifestyle", []string{"diy", "tutorial", "beginner"}, 660},

		// --- Cross-category (more variety) ---
		{"科技 x 美食：用 AI 設計一週菜單", "讓 ChatGPT 幫你規劃一整週的健康餐，實際照著吃看看結果⋯⋯", "technology", []string{"cooking", "tech-review", "funny"}, 780},
		{"動畫解說：為什麼你總是拖延？", "從心理學角度分析拖延症的成因，以及三個實用的改善方法。", "education", []string{"animation", "tutorial"}, 540},
		{"挑戰用 DIY 材料做出電動滑板", "用廢棄零件自己組裝一台電動滑板，成本不到三千塊！", "technology", []string{"diy", "fitness", "funny"}, 1400},
		{"遊戲配樂幕後：作曲家深度訪談", "專訪知名遊戲配樂作曲家，揭密那些經典配樂的創作故事。", "entertainment", []string{"interview", "gameplay", "music-video"}, 1800},
		{"新手入門攝影：手機拍出電影感", "不需要昂貴器材！用手機就能拍出專業級影片的 5 個技巧。", "education", []string{"beginner", "tutorial", "diy"}, 660},
		{"健身 x 搞笑：情侶雙人運動挑戰", "跟另一半挑戰超爆笑的雙人健身動作，笑到腹肌要練出來了！", "sports", []string{"fitness", "funny", "vlog"}, 720},
	}

	totalVideos := len(videos)
	log.Printf("Seeding %d videos with pre-defined diverse content...", totalVideos)

	for i, vs := range videos {
		creator := users[i%len(users)]
		cat := catBySlug[vs.Category]

		desc := vs.Description
		thumbURL := fmt.Sprintf("/fenzvideo/thumbnails/thumb_%d.jpg", (i%20)+1)
		viewsMember := uint64(rand.Intn(8000) + 200)
		viewsNonMember := uint64(rand.Intn(15000) + 500)

		video := model.Video{
			UserID:         creator.ID,
			CategoryID:     cat.ID,
			Title:          vs.Title,
			Description:    &desc,
			VideoURL:       fmt.Sprintf("/fenzvideo/videos/sample_%d.mp4", (i%20)+1),
			ThumbnailURL:   &thumbURL,
			Duration:       vs.Duration,
			ViewsMember:    viewsMember,
			ViewsNonMember: viewsNonMember,
			AccessTier:     0,
			IsPublished:    true,
			IsHidden:       false,
		}

		if err := db.Create(&video).Error; err != nil {
			log.Printf("Failed to create video %d: %v", i+1, err)
			continue
		}

		// Associate tags via GORM many2many
		var videoTags []model.Tag
		for _, slug := range vs.Tags {
			if tag, ok := tagBySlug[slug]; ok {
				videoTags = append(videoTags, tag)
			} else {
				log.Printf("  Warning: tag slug %q not found, skipping", slug)
			}
		}
		if len(videoTags) > 0 {
			if err := db.Model(&video).Association("Tags").Append(&videoTags); err != nil {
				log.Printf("  Warning: failed to associate tags for video %d: %v", video.ID, err)
			}
		}

		log.Printf("[%d/%d] %s (tags: %v)", i+1, totalVideos, vs.Title, vs.Tags)
	}

	log.Printf("Seeded %d videos", totalVideos)

}
