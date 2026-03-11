package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	v1 "backend/api/fenzvideo/v1"
	"backend/internal/conf"
	"backend/internal/data"
	pjwt "backend/internal/pkg/jwt"
	"backend/internal/pkg/upload"
	"backend/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/websocket"
)

const realtimeReadDeadline = 10 * time.Minute

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	ac *conf.Auth,
	logger log.Logger,
	authSvc *service.AuthService,
	categorySvc *service.CategoryService,
	tagSvc *service.TagService,
	videoSvc *service.VideoService,
	searchSvc *service.SearchService,
	channelSvc *service.ChannelService,
	adminSvc *service.AdminService,
	uploader *upload.MinIOUploader,
	hub *data.RealtimeHub,
) *kratoshttp.Server {
	var opts = []kratoshttp.ServerOption{
		kratoshttp.Middleware(
			recovery.Recovery(),
			JWTAuthMiddleware(ac.JwtSecret),
			AdminGuardMiddleware(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratoshttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratoshttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratoshttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratoshttp.NewServer(opts...)

	// Register proto-generated HTTP routes for all services
	v1.RegisterAuthServiceHTTPServer(srv, authSvc)
	v1.RegisterCategoryServiceHTTPServer(srv, categorySvc)
	v1.RegisterTagServiceHTTPServer(srv, tagSvc)
	v1.RegisterVideoServiceHTTPServer(srv, videoSvc)
	v1.RegisterSearchServiceHTTPServer(srv, searchSvc)
	v1.RegisterChannelServiceHTTPServer(srv, channelSvc)
	v1.RegisterAdminServiceHTTPServer(srv, adminSvc)

	// Two-step file upload endpoints (not proto-generated, since gRPC doesn't support multipart)
	route := srv.Route("/")
	route.POST("/api/v1/upload/video", handleUpload(uploader, "videos", []string{
		"video/mp4", "video/webm", "video/quicktime",
	}, 500<<20, logger)) // 500MB max
	route.POST("/api/v1/upload/thumbnail", handleUpload(uploader, "thumbnails", []string{
		"image/jpeg", "image/png", "image/webp",
	}, 10<<20, logger)) // 10MB max
	route.POST("/api/v1/videos/{id}/likes", handleLike(videoSvc, ac.JwtSecret, logger))
	route.GET("/api/v1/realtime/ws", handleRealtimeWS(ac.JwtSecret, hub, logger))

	return srv
}

func handleLike(videoSvc *service.VideoService, jwtSecret string, logger log.Logger) kratoshttp.HandlerFunc {
	return func(ctx kratoshttp.Context) error {
		claims, err := authenticateRequest(ctx.Request(), jwtSecret)
		if err != nil {
			writeJSONError(ctx.Response(), http.StatusUnauthorized, err.Error())
			return nil
		}

		videoID, err := parseVideoIDFromPath(ctx.Request().URL.Path)
		if err != nil {
			writeJSONError(ctx.Response(), http.StatusBadRequest, "invalid video id")
			return nil
		}

		if err := videoSvc.LikeVideo(ctx.Request().Context(), claims.UserID, videoID); err != nil {
			writeJSONError(ctx.Response(), http.StatusBadRequest, err.Error())
			return nil
		}

		ctx.Response().Header().Set("Content-Type", "application/json")
		fmt.Fprint(ctx.Response(), `{"success":true}`)
		return nil
	}
}

func handleRealtimeWS(jwtSecret string, hub *data.RealtimeHub, logger log.Logger) kratoshttp.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return func(ctx kratoshttp.Context) error {
		if hub == nil {
			writeJSONError(ctx.Response(), http.StatusServiceUnavailable, "realtime unavailable")
			return nil
		}

		claims, err := authenticateRequest(ctx.Request(), jwtSecret)
		if err != nil {
			writeJSONError(ctx.Response(), http.StatusUnauthorized, err.Error())
			return nil
		}

		conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
		if err != nil {
			log.NewHelper(logger).Warnf("websocket upgrade failed: %v", err)
			return nil
		}

		connectionID := hub.Register(context.Background(), claims.UserID, conn)
		defer hub.Unregister(context.Background(), claims.UserID, connectionID)

		_ = conn.SetReadDeadline(time.Now().Add(realtimeReadDeadline))
		conn.SetPongHandler(func(string) error {
			hub.RefreshPresence(context.Background(), claims.UserID, connectionID)
			return conn.SetReadDeadline(time.Now().Add(realtimeReadDeadline))
		})

		if err := conn.WriteJSON(map[string]any{"type": "connected"}); err != nil {
			return nil
		}

		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
			if messageType == websocket.PingMessage || messageType == websocket.TextMessage {
				hub.RefreshPresence(context.Background(), claims.UserID, connectionID)
			}
		}
	}
}

func authenticateRequest(r *http.Request, jwtSecret string) (*pjwt.Claims, error) {
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		token = strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			token = ""
		}
	}
	if token == "" {
		return nil, fmt.Errorf("missing token")
	}
	return pjwt.ParseToken(jwtSecret, token)
}

func parseVideoIDFromPath(path string) (uint64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		return 0, fmt.Errorf("invalid path")
	}
	return strconv.ParseUint(parts[3], 10, 64)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error":%q}`, message)
}

// handleUpload creates an HTTP handler for file uploads to MinIO.
// dir: MinIO subdirectory (e.g. "videos", "thumbnails")
// allowedTypes: permitted Content-Types
// maxSize: maximum file size in bytes
func handleUpload(uploader *upload.MinIOUploader, dir string, allowedTypes []string, maxSize int64, logger log.Logger) kratoshttp.HandlerFunc {
	return func(ctx kratoshttp.Context) error {
		l := log.NewHelper(logger)

		r := ctx.Request()
		w := ctx.Response()

		// Enforce max upload size
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"invalid file: %s"}`, err.Error())
			return nil
		}
		defer file.Close()

		ct := header.Header.Get("Content-Type")
		allowed := false
		for _, t := range allowedTypes {
			if ct == t {
				allowed = true
				break
			}
		}
		if !allowed {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"unsupported content type: %s"}`, ct)
			return nil
		}

		ext := upload.ExtFromContentType(ct)
		objectPath, err := uploader.Upload(context.Background(), io.Reader(file), header.Size, ct, dir, ext)
		if err != nil {
			l.Errorf("upload failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"upload failed"}`)
			return nil
		}

		url := uploader.GetURL(objectPath)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"url":"%s","path":"%s"}`, url, objectPath)
		return nil
	}
}
