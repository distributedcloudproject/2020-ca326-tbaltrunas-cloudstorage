package network

import (
	"cloud/utils"

	"os"
	"fmt"
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Note that the model must have the same name as the table.
// i.e. model 'Account' for table 'accounts'.
type Account struct {
	gorm.Model
	Credentials
}

// Adapted from: https://blog.usejournal.com/authentication-in-golang-c0677bcce1a8
func ConnectDB() (*gorm.DB, error) {
	databaseUser := os.Getenv("DB_USER")
	databasePassword := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")
	databaseHost := os.Getenv("DB_HOST")
	databaseType := os.Getenv("DB_TYPE")
	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", 
						 databaseHost, databaseUser, databaseName, databasePassword)
	
	db, err := gorm.Open(databaseType, dbURI)
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Account{})
	return db, nil
}

func DBGetAccountByUsername(username string) (Account, error) {
	utils.GetLogger().Printf("[DEBUG] Username: %s", username)
	var storedAccount Account
	db, err := ConnectDB()
	if err != nil {
		utils.GetLogger().Printf("[ERROR] %v", err)
		return storedAccount, errors.New("DB connection error")
	}
	defer db.Close()

	err = db.Where("username = ?", username).First(&storedAccount).Error
	return storedAccount, err
}
