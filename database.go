package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

func (a *App) InitialiseDatabase() {

	// due to postgres docker container not starting
	// up in time even with depends_on we have to keep
	// trying to connect. if after 60 secs we still
	// haven't connected we log fatal and stop
	timeout := 60 * time.Second
	start := time.Now()
	var err error
	x := 1
	for time.Since(start) < timeout {
		a.Log.Info().Msgf("Trying to connect to db...[%d]", x)
		a.DB, err = ConnectToDB()
		if err == nil {
			break
		}
		a.Log.Error().Err(err)
		time.Sleep(2 * time.Second)
		x++
	}

	if err != nil {
		a.Log.Fatal().Msgf("Failed to connect to the database after %s seconds", timeout)
	}

	a.Log.Info().Msg("Connected to db successfully")
	a.MigrateModels()
}

func ConnectToDB() (*gorm.DB, error) {

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (a *App) MigrateModels() {

	a.Log.Info().Msg("Migrating models")
	err := a.DB.AutoMigrate(&Review{})
	if err != nil {
		a.Log.Fatal().Msg(err.Error())
	}
	a.Log.Info().Msg("Models migrated successfully")
}