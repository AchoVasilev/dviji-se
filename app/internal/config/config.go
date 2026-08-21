package config

import (
	"log"
	"net/netip"
	"net/url"
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
	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbName     string
	dbSSLMode  string
	dbMaxConns int

	// JWT
	jwtAccessKey  string
	jwtRefreshKey string

	// Security
	xsrfKey           string
	corsOrigins       []string
	allowRegistration bool
	trustedProxies    []netip.Prefix

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

	// Feature flags
	workoutsEnabled  bool
	nutritionEnabled bool

	// Admin bootstrap
	adminEmail    string
	adminPassword string

	// App
	baseURL             string
	tinymceURL          string
	privacyContactEmail string
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
			xsrfKey:     getEnvRequired("XSRF"),
			corsOrigins: getEnvSlice("CORS_ORIGINS", ",", []string{"http://localhost:3000"}),
			// Off by default: this is a single author blog, so public
			// registration is opt in rather than something you must remember
			// to switch off.
			allowRegistration: getEnvBool("ALLOW_REGISTRATION", false),
			trustedProxies:    getEnvPrefixes("TRUSTED_PROXIES"),

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

			// Feature flags: the sections are hidden until their pages exist.
			workoutsEnabled:  getEnvBool("ENABLED_WORKOUTS", false),
			nutritionEnabled: getEnvBool("ENABLED_NUTRITION", false),

			// Admin bootstrap
			adminEmail:    getEnv("ADMIN_EMAIL", ""),
			adminPassword: getEnv("ADMIN_PASSWORD", ""),

			// App
			baseURL:    getEnv("APP_BASE_URL", "http://localhost:8080"),
			tinymceURL: getEnv("TINYMCE_URL", ""),
		}

		// The privacy policy has to name an address people can write to, so
		// this falls back to privacy@<site host> rather than leaving the page
		// telling visitors to contact nobody.
		c.privacyContactEmail = getEnv("PRIVACY_CONTACT_EMAIL", defaultContactEmail(c.baseURL))
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

func XSRFKey() string         { return get().xsrfKey }
func CORSOrigins() []string   { return get().corsOrigins }
func AllowRegistration() bool { return get().allowRegistration }

// TrustedProxies lists the networks whose X-Forwarded-For / X-Real-IP headers
// may be believed. Empty means trust none, which is the safe default: any
// client can set those headers, so believing them without a proxy in front
// lets a caller forge its own address.
func TrustedProxies() []netip.Prefix { return get().trustedProxies }

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

// --- Feature flags ---

// The flags below gate navigation entries for sections that are not built
// yet. They control visibility only: turning one on shows the link, but the
// route still has to exist or it will 404.

// WorkoutsEnabled reports whether the workouts section is advertised.
func WorkoutsEnabled() bool { return get().workoutsEnabled }

// NutritionEnabled reports whether the nutrition section is advertised.
func NutritionEnabled() bool { return get().nutritionEnabled }

// --- Admin bootstrap ---

// AdminEmail and AdminPassword seed the first administrator when the database
// has none. They are ignored once an administrator exists.
func AdminEmail() string    { return get().adminEmail }
func AdminPassword() string { return get().adminPassword }

// --- App ---

func BaseURL() string    { return get().baseURL }
func TinyMCEURL() string { return get().tinymceURL }

// PrivacyContactEmail is the address the privacy policy points data subjects
// at. Set PRIVACY_CONTACT_EMAIL to a mailbox that is actually read.
func PrivacyContactEmail() string { return get().privacyContactEmail }

// defaultContactEmail derives privacy@<host> from the site's base URL. A base
// URL that does not parse leaves the address empty, and the policy then omits
// the mail link rather than printing a broken one.
func defaultContactEmail(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	host := parsed.Hostname()
	if host == "" {
		return ""
	}

	return "privacy@" + host
}

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

// getEnvPrefixes reads a comma separated list of CIDRs or bare IP addresses.
// Entries that do not parse are dropped with a warning rather than failing
// startup, so one bad entry cannot take the service down.
func getEnvPrefixes(key string) []netip.Prefix {
	var prefixes []netip.Prefix

	for _, entry := range getEnvSlice(key, ",", nil) {
		if strings.Contains(entry, "/") {
			prefix, err := netip.ParsePrefix(entry)
			if err != nil {
				log.Printf("Ignoring invalid CIDR %q in %s: %v", entry, key, err)
				continue
			}

			prefixes = append(prefixes, prefix)
			continue
		}

		addr, err := netip.ParseAddr(entry)
		if err != nil {
			log.Printf("Ignoring invalid IP %q in %s: %v", entry, key, err)
			continue
		}

		prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
	}

	return prefixes
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
