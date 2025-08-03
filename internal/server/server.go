package server

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/cncde/songs-backend/internal/config"
	"github.com/cncde/songs-backend/internal/middleware"
	"github.com/cncde/songs-backend/internal/repository"
	"github.com/cncde/songs-backend/internal/service"
	"github.com/cncde/songs-backend/internal/storage"
)

// Dependencies holds all external dependencies
type Dependencies struct {
	Config       *config.Config
	DynamoClient *dynamodb.Client
	S3Client     *s3.Client
}

// Server represents the HTTP server with all dependencies
type Server struct {
	config   *config.Config
	router   *gin.Engine
	services *Services
}

// Services holds all business logic services
type Services struct {
	SongService    *service.SongService
	CommentService *service.CommentService
	MediaService   *service.MediaService
}

// New creates a new server instance with all dependencies
func New(deps *Dependencies) (*Server, error) {
	// Initialize repositories
	dynamoClient := repository.NewDynamoDBClient(deps.DynamoClient)
	songRepo := repository.NewSongRepository(dynamoClient, deps.Config.SongsTable)
	commentRepo := repository.NewCommentRepository(dynamoClient, deps.Config.CommentsTable)

	// Initialize storage services
	s3Service := storage.NewS3Service(deps.S3Client, deps.Config.S3Bucket)

	// Initialize business services
	songService := service.NewSongService(songRepo, commentRepo)
	commentService := service.NewCommentService(commentRepo, songRepo)
	mediaService := service.NewMediaService(s3Service)

	services := &Services{
		SongService:    songService,
		CommentService: commentService,
		MediaService:   mediaService,
	}

	// Create server
	server := &Server{
		config:   deps.Config,
		services: services,
	}

	// Initialize router
	router, err := server.setupRouter()
	if err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}
	server.router = router

	return server, nil
}

// Handler returns the HTTP handler (for Lambda and local development)
func (s *Server) Handler() *gin.Engine {
	return s.router
}

// setupRouter initializes the Gin router with middleware and routes
func (s *Server) setupRouter() (*gin.Engine, error) {
	// Set Gin mode based on environment
	if s.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	r := gin.New()

	// Add global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Initialize JWT middleware
	jwtMiddleware, err := middleware.NewJWTMiddleware(
		s.config.JWKSUrl,
		s.config.CognitoUserPoolID,
		"", // Expected audience - could be set from config
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT middleware: %w", err)
	}

	// Health check endpoint (no auth required)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "songs-api",
			"version": s.config.APIVersion,
		})
	})

	// API version group
	api := r.Group("/api/" + s.config.APIVersion)

	// Public routes (no auth required)
	public := api.Group("/public")
	{
		// Public song endpoints (only published songs)
		public.GET("/songs", s.handleListPublishedSongs)
		public.GET("/songs/:id", s.handleGetPublicSong)
		public.GET("/songs/search", s.handleSearchSongs)
	}

	// Protected routes (auth required)
	protected := api.Group("")
	protected.Use(jwtMiddleware.Middleware())
	{
		// Song endpoints
		songs := protected.Group("/songs")
		{
			songs.POST("", s.handleCreateSong)
			songs.GET("", s.handleListUserSongs)
			songs.GET("/:id", s.handleGetSong)
			songs.PUT("/:id", s.handleUpdateSong)
			songs.DELETE("/:id", s.handleDeleteSong)
			
			// Song media endpoints
			songs.POST("/:id/images", s.handleAddImageToSong)
			songs.DELETE("/:id/images", s.handleRemoveImageFromSong)
			songs.POST("/:id/audio", s.handleAddAudioToSong)
			songs.DELETE("/:id/audio", s.handleRemoveAudioFromSong)
			
			// Comment endpoints for songs
			songs.GET("/:id/comments", s.handleListSongComments)
			songs.POST("/:id/comments", s.handleCreateComment)
		}

		// Comment endpoints
		comments := protected.Group("/comments")
		{
			comments.GET("/:id", s.handleGetComment)
			comments.PUT("/:id", s.handleUpdateComment)
			comments.DELETE("/:id", s.handleDeleteComment)
			comments.GET("", s.handleListUserComments)
		}

		// Media/file endpoints
		media := protected.Group("/media")
		{
			media.POST("/upload-url", s.handleGenerateUploadURL)
			media.GET("/download-url/:fileKey", s.handleGenerateDownloadURL)
			media.DELETE("/:fileKey", s.handleDeleteFile)
			media.GET("/files", s.handleListUserFiles)
		}

		// User stats endpoints
		stats := protected.Group("/stats")
		{
			stats.GET("/songs", s.handleGetSongStats)
			stats.GET("/comments", s.handleGetCommentStats)
			stats.GET("/media", s.handleGetMediaStats)
		}
	}

	return r, nil
}

// Placeholder handler methods - these would be implemented in separate handler files
// For now, return basic responses to make the server compilable

func (s *Server) handleListPublishedSongs(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List published songs endpoint"})
}

func (s *Server) handleGetPublicSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get public song endpoint"})
}

func (s *Server) handleSearchSongs(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Search songs endpoint"})
}

func (s *Server) handleCreateSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create song endpoint"})
}

func (s *Server) handleListUserSongs(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List user songs endpoint"})
}

func (s *Server) handleGetSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get song endpoint"})
}

func (s *Server) handleUpdateSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update song endpoint"})
}

func (s *Server) handleDeleteSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete song endpoint"})
}

func (s *Server) handleAddImageToSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Add image to song endpoint"})
}

func (s *Server) handleRemoveImageFromSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Remove image from song endpoint"})
}

func (s *Server) handleAddAudioToSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Add audio to song endpoint"})
}

func (s *Server) handleRemoveAudioFromSong(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Remove audio from song endpoint"})
}

func (s *Server) handleListSongComments(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List song comments endpoint"})
}

func (s *Server) handleCreateComment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create comment endpoint"})
}

func (s *Server) handleGetComment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get comment endpoint"})
}

func (s *Server) handleUpdateComment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update comment endpoint"})
}

func (s *Server) handleDeleteComment(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete comment endpoint"})
}

func (s *Server) handleListUserComments(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List user comments endpoint"})
}

func (s *Server) handleGenerateUploadURL(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Generate upload URL endpoint"})
}

func (s *Server) handleGenerateDownloadURL(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Generate download URL endpoint"})
}

func (s *Server) handleDeleteFile(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete file endpoint"})
}

func (s *Server) handleListUserFiles(c *gin.Context) {
	c.JSON(200, gin.H{"message": "List user files endpoint"})
}

func (s *Server) handleGetSongStats(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get song stats endpoint"})
}

func (s *Server) handleGetCommentStats(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get comment stats endpoint"})
}

func (s *Server) handleGetMediaStats(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get media stats endpoint"})
}