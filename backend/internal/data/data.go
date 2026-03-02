package data

import (
	"backend/internal/biz"
	"backend/internal/conf"
	"backend/internal/data/model"
	"backend/internal/pkg/hash"
	"backend/internal/pkg/upload"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var ProviderSet = wire.NewSet(
	NewData, NewDB, NewRedisClient, NewMinIOClient, NewNATSConn,
	NewAuthRepo,
	NewCategoryRepo,
	NewTagRepo,
	NewVideoRepo,
	NewSearchRepo,
	NewChannelRepo,
	NewAdminRepo,
	NewMembershipChecker,
	NewUploader,
	NewVideoCache,
)

// NewMembershipChecker adapts biz.ChannelRepo (which includes HasMembership)
// to the biz.MembershipChecker interface needed by VideoUsecase.
// Wire can't bind interface→interface directly, so this adapter bridges them.
func NewMembershipChecker(repo biz.ChannelRepo) biz.MembershipChecker {
	return repo
}

// NewUploader creates a MinIOUploader for the two-step file upload endpoints.
func NewUploader(mc *minio.Client, c *conf.Storage) *upload.MinIOUploader {
	return upload.NewMinIOUploader(mc, c.Bucket)
}

type Data struct {
	DB    *gorm.DB
	Redis *redis.Client
	MinIO *minio.Client
	NATS  *nats.Conn
}

func NewData(db *gorm.DB, rdb *redis.Client, mc *minio.Client, nc *nats.Conn, ac *conf.Admin, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		DB:    db,
		Redis: rdb,
		MinIO: mc,
		NATS:  nc,
	}

	// Ensure admin account exists (idempotent)
	if ac != nil && ac.Username != "" {
		ensureAdmin(db, ac, logger)
	}

	// Warm up recommendation cache before servers accept traffic.
	d.WarmUpCache(context.Background(), logger)

	// Start background workers (view flush, cleanup retry) with cancelable context.
	bgCtx, bgCancel := context.WithCancel(context.Background())
	StartBackgroundWorkers(bgCtx, d, logger)

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		bgCancel() // stop background workers
		if rdb != nil {
			rdb.Close()
		}
		if nc != nil {
			nc.Close()
		}
	}

	return d, cleanup, nil
}

func ensureAdmin(db *gorm.DB, ac *conf.Admin, logger log.Logger) {
	l := log.NewHelper(logger)

	var user model.User
	err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(gormLogger.Silent)}).Where("username = ?", ac.Username).First(&user).Error
	if err == nil {
		// Admin exists — update password if changed
		hashed, hashErr := hash.HashPassword(ac.Password)
		if hashErr != nil {
			l.Warnf("failed to hash admin password: %v", hashErr)
			return
		}
		if !hash.ComparePassword(user.Password, ac.Password) {
			db.Model(&user).Update("password", hashed)
			l.Info("admin password updated")
		}
		return
	}

	// Create admin user
	hashed, hashErr := hash.HashPassword(ac.Password)
	if hashErr != nil {
		l.Errorf("failed to hash admin password: %v", hashErr)
		return
	}

	admin := &model.User{
		Username:    ac.Username,
		DisplayName: "Admin",
		Password:    hashed,
		Role:        "admin",
	}
	if err := db.Create(admin).Error; err != nil {
		l.Errorf("failed to create admin user: %v", err)
		return
	}

	// Auto-create channel for admin
	ch := &model.Channel{
		UserID:     admin.ID,
		MonthlyFee: 0,
	}
	if err := db.Create(ch).Error; err != nil {
		l.Warnf("failed to create admin channel: %v", err)
	}

	l.Infof("admin account '%s' created", ac.Username)
}

func NewDB(c *conf.Data, logger log.Logger) *gorm.DB {
	l := log.NewHelper(logger)

	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		l.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		l.Fatalf("failed to get sql.DB: %v", err)
	}

	if c.Database.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(int(c.Database.MaxIdleConns))
	}
	if c.Database.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(int(c.Database.MaxOpenConns))
	}
	if c.Database.ConnMaxLifetime != nil {
		sqlDB.SetConnMaxLifetime(c.Database.ConnMaxLifetime.AsDuration())
	}

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
		l.Fatalf("failed to auto-migrate database: %v", err)
	}

	// Create FULLTEXT index for search (GORM AutoMigrate cannot create FULLTEXT indexes).
	// MySQL doesn't support IF NOT EXISTS for CREATE INDEX, so check first.
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'videos' AND index_name = 'idx_videos_title_fulltext'").Scan(&count)
	if count == 0 {
		db.Exec("CREATE FULLTEXT INDEX idx_videos_title_fulltext ON videos(title)")
	}

	l.Info("database connected and migrated")
	return db
}

func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(logger)

	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		l.Warnf("failed to connect redis: %v (continuing without redis)", err)
	} else {
		l.Info("redis connected")
	}

	return rdb
}

func NewMinIOClient(c *conf.Storage, logger log.Logger) *minio.Client {
	l := log.NewHelper(logger)

	mc, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
		Secure: c.UseSsl,
		Region: c.Region,
	})
	if err != nil {
		l.Warnf("failed to create MinIO client: %v (continuing without MinIO)", err)
		return nil
	}

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := mc.BucketExists(ctx, c.Bucket)
	if err != nil {
		l.Warnf("failed to check MinIO bucket: %v", err)
	} else if !exists {
		if err := mc.MakeBucket(ctx, c.Bucket, minio.MakeBucketOptions{Region: c.Region}); err != nil {
			l.Warnf("failed to create MinIO bucket: %v", err)
		} else {
			l.Infof("MinIO bucket '%s' created", c.Bucket)
		}
	}

	l.Info("MinIO client initialized")
	return mc
}

func NewNATSConn(c *conf.NATS, logger log.Logger) *nats.Conn {
	l := log.NewHelper(logger)

	nc, err := nats.Connect(c.Url)
	if err != nil {
		l.Warnf("failed to connect NATS: %v (continuing without NATS)", err)
		return nil
	}

	l.Info("NATS connected")
	return nc
}
