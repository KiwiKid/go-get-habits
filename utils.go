package main

import (
	"log"
	"os"
	"strings"
)

func lowerAndReplaceSpaces(s string) string {
	lowercased := strings.ToLower(s)
	return strings.Replace(lowercased, " ", "_", -1)
}

func checkAndPublishAll() {
	db, closeDB, err := NewDatabase()
	if err != nil {
		panic(err)
	}
	defer closeDB()
	checkErr := db.checkAndUpdateHabits()

	if checkErr != nil {
		panic(checkErr)
	}

	publisher := NewHabitPublisher()

	// Connect to the MQTT broker.
	publisher.Connect()
	defer publisher.Disconnect()

	rows, err := db.GetAllHabits(true)
	if err != nil {
		panic(err)
	}
	// Publish the habits.
	publisher.PublishHabits(rows)
}

func GetEnvWithDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("Environment variable %s not found. Using default value: %s", key, defaultValue)
		return defaultValue
	}
	log.Printf("Found environment variable %s with value: %s", key, value)
	return value
}
