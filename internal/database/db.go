package database

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var DB *gorm.DB

func InitDB(dbPath string) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	runMigrations()

	// Seed data
	seedAdmin()
	seedConfigs()
}

func runMigrations() {
	// 1. Create migration tracking table if not exists
	DB.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		name TEXT PRIMARY KEY,
		executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Failed to read migration files: %v", err)
	}

	var filenames []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		// 2. Check if migration already ran
		var count int64
		DB.Table("migrations").Where("name = ?", filename).Count(&count)
		if count > 0 {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + filename)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", filename, err)
		}

		log.Printf("Executing migration: %s", filename)
		if err := DB.Exec(string(content)).Error; err != nil {
			log.Fatalf("Migration failed (%s): %v", filename, err)
		}

		// 3. Record migration
		DB.Exec("INSERT INTO migrations (name) VALUES (?)", filename)
	}
}

func seedAdmin() {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count == 0 {
		password := generateRandomString(12)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		admin := User{
			Username: "admin",
			Password: string(hashedPassword),
		}
		DB.Create(&admin)

		fmt.Println("=========================================================")
		fmt.Println("FIRST RUN: Admin user created")
		fmt.Printf("Username: admin\n")
		fmt.Printf("Password: %s\n", password)
		fmt.Println("PLEASE SAVE THESE CREDENTIALS!")
		fmt.Println("=========================================================")
	}

	var jwtSecret string
	err := DB.Model(&ConfigItem{}).Where("key = ?", "jwt_secret").Select("value").Scan(&jwtSecret).Error
	if err != nil || jwtSecret == "" {
		secret := generateRandomString(32)
		DB.Create(&ConfigItem{Key: "jwt_secret", Value: secret})
		log.Println("Generated new JWT secret")
	}
}

func seedConfigs() {
	defaults := map[string]string{
		"enable_openai":    "true",
		"enable_anthropic": "true",
		"enable_gemini":    "true",
	}

	for k, v := range defaults {
		var count int64
		DB.Model(&ConfigItem{}).Where("key = ?", k).Count(&count)
		if count == 0 {
			DB.Create(&ConfigItem{Key: k, Value: v})
			log.Printf("Seeded config: %s = %s", k, v)
		}
	}
}

func GetConfig(key string) string {
	var item ConfigItem
	if err := DB.Where("key = ?", key).First(&item).Error; err != nil {
		log.Printf("[DB] Config key '%s' not found: %v", key, err)
		return ""
	}
	// log.Printf("[DB] GetConfig '%s' = '%s'", key, item.Value)
	return item.Value
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "secret"
	}
	return hex.EncodeToString(b)[:n]
}
