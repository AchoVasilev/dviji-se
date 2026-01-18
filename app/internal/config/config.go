package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	c    *cfg
	once sync.Once
)

type cfg struct {
	// Server
	port        string
	environment string

	// Database
	dbHost        string
	dbPort        string
	dbUser        string
	dbPassword    string
	dbName        string
	dbSSLMode     string
	dbMaxConns    int

	// JWT
	jwtAccessKey  string
	jwtRefreshKey string

	// Security
	xsrfKey           string
	corsOrigins       []string
	allowRegistration bool

	// SMTP
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	smtpFrom     string

	// Cloudinary
	cloudinaryCloudName string
	cloudinaryAPIKey    string
	cloudinaryAPISecret string
	cloudinaryFolder    string

	// App
	baseURL string
}

func load() {
	once.Do(func() {
		c = &cfg{
			// Server
			port:        getEnv("PORT", "8080"),
			environment: getEnv("ENVIRONMENT", "development"),

			// Database
			dbHost:     getEnv("DBHOST", "localhost"),
			dbPort:     getEnv("DBPORT", "5432"),
			dbUser:     getEnv("DBUSER", "postgres"),
			dbPassword: getEnv("DBPASSWORD", ""),
			dbName:     getEnv("DBNAME", "dviji_se"),
			dbSSLMode:  getEnv("DBSSLMODE", "disable"),
			dbMaxConns: getEnvInt("DBCONNECTIONS", 10),

			// JWT
			jwtAccessKey:  getEnvRequired("JWT_KEY"),
			jwtRefreshKey: getEnvRequired("JWT_REFRESH_KEY"),

			// Security
			xsrfKey:           getEnvRequired("XSRF"),
			corsOrigins:       getEnvSlice("CORS_ORIGINS", ",", []string{"http://localhost:3000"}),
			allowRegistration: getEnvBool("ALLOW_REGISTRATION", false),

			// SMTP
			smtpHost:     getEnv("SMTP_HOST", ""),
			smtpPort:     getEnv("SMTP_PORT", "587"),
			smtpUsername: getEnv("SMTP_USERNAME", ""),
			smtpPassword: getEnv("SMTP_PASSWORD", ""),
			smtpFrom:     getEnv("SMTP_FROM", "noreply@example.com"),

			// Cloudinary
			cloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
			cloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
			cloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
			cloudinaryFolder:    getEnv("CLOUDINARY_FOLDER", "uploads"),

			// App
			baseURL: getEnv("APP_BASE_URL", "http://localhost:8080"),
		}
	})
}

func get() *cfg {
	if c == nil {
		load()
	}
	return c
}

// --- Server ---

func Port() string        { return get().port }
func Environment() string { return get().environment }
func IsDevelopment() bool { return get().environment == "development" }
func IsProduction() bool  { return get().environment == "production" }
func IsTest() bool        { return get().environment == "test" }

// --- Database ---

func DBHost() string     { return get().dbHost }
func DBPort() string     { return get().dbPort }
func DBUser() string     { return get().dbUser }
func DBPassword() string { return get().dbPassword }
func DBName() string     { return get().dbName }
func DBSSLMode() string  { return get().dbSSLMode }
func DBMaxConns() int    { return get().dbMaxConns }

// --- JWT ---

func JWTAccessKey() string  { return get().jwtAccessKey }
func JWTRefreshKey() string { return get().jwtRefreshKey }

// --- Security ---

func XSRFKey() string           { return get().xsrfKey }
func CORSOrigins() []string     { return get().corsOrigins }
func AllowRegistration() bool   { return get().allowRegistration }

// --- SMTP ---

func SMTPHost() string     { return get().smtpHost }
func SMTPPort() string     { return get().smtpPort }
func SMTPUsername() string { return get().smtpUsername }
func SMTPPassword() string { return get().smtpPassword }
func SMTPFrom() string     { return get().smtpFrom }
func SMTPConfigured() bool { return get().smtpHost != "" && get().smtpUsername != "" }

// --- Cloudinary ---

func CloudinaryCloudName() string { return get().cloudinaryCloudName }
func CloudinaryAPIKey() string    { return get().cloudinaryAPIKey }
func CloudinaryAPISecret() string { return get().cloudinaryAPISecret }
func CloudinaryFolder() string    { return get().cloudinaryFolder }
func CloudinaryConfigured() bool {
	return get().cloudinaryCloudName != "" && get().cloudinaryAPIKey != "" && get().cloudinaryAPISecret != ""
}

// --- App ---

func BaseURL() string { return get().baseURL }

// --- Helpers ---

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

func getEnvSlice(key, separator string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, separator)
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}
