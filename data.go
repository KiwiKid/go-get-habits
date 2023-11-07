package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ResetFrequency string

const (
	Minutes ResetFrequency = "minutes"
	Hourly  ResetFrequency = "hourly"
	Daily   ResetFrequency = "daily"
	Weekly  ResetFrequency = "weekly"
	Monthly ResetFrequency = "monthly"

	// Add more frequencies as needed
)

type Habit struct {
	ID             uint           `gorm:"primarykey"`
	Name           string         `gorm:"size:255"`
	ResetFrequency ResetFrequency `gorm:"type:varchar(10)"`
	ResetValue     int            `gorm:"type:int"`

	Group           string    `gorm:"size:255"`
	IsActive        bool      `gorm:"type:boolean"`
	LastComplete    time.Time `gorm:"type:datetime"`
	NeedsCompletion bool      `gorm:"type:boolean"`
	StartHour       int       `gorm:"type:int"` // Value between 0-23
	StartMinute     int       `gorm:"type:int"` // Value between 0-23
	EndHour         int       `gorm:"type:int"`
	EndMinute       int       `gorm:"type:int"` // Value between 0-23
}

func (h Habit) getUniqueId() string {
	isDevStr := GetEnvWithDefault("IS_DEV", "false")

	return fmt.Sprintf("%d-%s", h.ID, isDevStr)
}

func (h Habit) getMQTTTopic(topic string, msgType string) string {

	if msgType != "state" && msgType != "config" {
		log.Fatalf("only 'state' and 'config' are supported")
	}

	fullTopic := "homeassistant/binary_sensor/" + topic + "/" + h.Name + "/" + msgType
	publishTopic := lowerAndReplaceSpaces(fullTopic)

	println(publishTopic)
	return publishTopic
}

func (h *Habit) getDeviceName() (name string, id string) {

	var deviceName string
	isDevStr := GetEnvWithDefault("IS_DEV", "false")
	if isDevStr == "true" {
		deviceName = h.Group + "_dev"
	} else {
		deviceName = h.Group
	}

	deviceId := lowerAndReplaceSpaces(deviceName)

	return deviceName, deviceId
}

type HabitUpdates struct {
	ID             *uint
	Name           *string
	ResetFrequency *ResetFrequency
	ResetValue     *int
	Group          *string
	IsActive       *bool
	StartHour      *int
	StartMinute    *int
	EndHour        *int
	EndMinute      *int
}

type Note struct {
	ID                uint   `gorm:"primarykey"`
	Title             string `gorm:"type:varchar(1024)"`
	Content           string `gorm:"type:varchar(1024)"`
	OnlyRelevantOnDay string `gorm:"type:varchar(255)"`
}

func (h *Note) getDeviceName() (name string, id string) {

	var deviceName string
	isDevStr := GetEnvWithDefault("IS_DEV", "false")
	if isDevStr == "true" {
		deviceName = "Notes_DEV"
	} else {
		deviceName = "Notes"
	}

	deviceId := lowerAndReplaceSpaces(deviceName)

	return deviceName, deviceId
}

func (n *Note) getMQTTTopic(postfix string) string {
	noteTitle := lowerAndReplaceSpaces(n.Title)
	//publishTopic := strings.ToLower(thisTopic)

	fullTopic := "homeassistant/sensor/notes/" + noteTitle + "/" + postfix

	return fullTopic
}

type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, func(), error) {

	dbpath := GetEnvWithDefault("DB_PATH", "db/habits.db")
	if _, err := os.Stat("db"); os.IsNotExist(err) {
		log.Print("making db dir")
		if err := os.Mkdir("db", 0755); err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	} else {
		log.Print("db dir exists")
	}

	if _, err := os.Stat(dbpath); os.IsNotExist(err) {
		log.Print("Database file does not exist")
	}

	db, err := gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	if err := db.AutoMigrate(&Habit{}); err != nil {
		return nil, nil, err
	}

	if err := db.AutoMigrate(&Note{}); err != nil {
		return nil, nil, err
	}

	closeFunc := func() {
		dbSession, err := db.DB()
		if err != nil {
			fmt.Printf("Failed to get database session: %s", err)
			return
		}
		if err := dbSession.Close(); err != nil {
			fmt.Printf("Error closing database: %s", err)
		}
	}

	return &Database{
		db: db,
	}, closeFunc, nil
}

func (d *Database) checkAndUpdateHabits() error {
	log.Printf("checking")

	db, closeDB, err := NewDatabase()
	if err != nil {
		log.Printf("Error initializing database: %s", err)
	}

	defer closeDB()

	rows, err := db.GetAllHabits(true)
	if err != nil {
		log.Printf("Error fetching habits: %s", err)
	}

	for _, habit := range rows {
		fmt.Println("Checking :" + habit.Name)

		if needsCompletion(habit) {
			fmt.Println("ACTION_NEEDED")

			habit.NeedsCompletion = true

			err := db.SetHabitNeedCompletion(habit.ID, true)

			if err != nil {
				log.Printf("ERROR ERROR saving check habit: %s", err)
				return err
			}

			// Handle what to do if habit needs completion. For instance, notify the user.
		} else {
			// This extended update might not be needed always(just after config update)
			habit.NeedsCompletion = false

			err := db.SetHabitNeedCompletion(habit.ID, false)

			if err != nil {
				log.Printf("ERROR ERROR saving check habit: %s", err)
				return err
			}
			fmt.Println("ALL GOOD" + habit.Name)
		}
	}
	return nil
}

func (d *Database) CreateHabit(h *Habit) error {
	return d.db.Create(h).Error
}

func (d *Database) GetAllHabits(isActive ...bool) ([]Habit, error) {
	var habits []Habit

	if d.db == nil {
		return nil, errors.New("database is not initialized")
	}

	db := d.db // Define and initialize the db variable
	if len(isActive) > 0 && isActive[0] {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&habits).Error; err != nil {
		return nil, err
	}
	return habits, nil
}

func (d *Database) GetHabitByID(id uint) (*Habit, error) {
	fmt.Printf(`GetHabitByID`)
	var habit Habit
	if err := d.db.First(&habit, id).Error; err != nil {
		return nil, err
	}
	return &habit, nil
}

func (d *Database) DeleteHabitByID(id uint) error {
	return d.db.Delete(&Habit{}, id).Error
}

func (d *Database) DeleteNoteByID(id uint) error {
	return d.db.Delete(&Note{}, id).Error
}

func (d *Database) CreateNote(n *Note) error {
	return d.db.Create(n).Error
}

func (d *Database) getAllNotes() ([]Note, error) {
	var notes []Note

	if d.db == nil {
		return nil, errors.New("database is not initialized")
	}

	if err := d.db.Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}

func (d *Database) GetNoteByID(id uint) (*Note, error) {
	if d.db == nil {
		return nil, errors.New("database is not initialized")
	}

	var note Note
	result := d.db.First(&note, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // or return nil, result.Error if you want to return the not found error
		}
		return nil, result.Error
	}
	return &note, nil
}

func (d *Database) EditHabit(id uint, updatedHabit *HabitUpdates) error {
	return d.db.Model(&Habit{}).Where("id = ?", id).Updates(updatedHabit).Error
}

func (d *Database) SetHabitNeedCompletion(id uint, NeedsCompletion bool) error {
	return d.db.Model(&Habit{}).Where("id = ?", id).Updates(map[string]interface{}{"NeedsCompletion": NeedsCompletion}).Error
}

func (d *Database) CompleteHabit(id uint) error {

	preRow, err := d.GetHabitByID(id)
	if err != nil {
		panic(err)
	}

	preRow.LastComplete = time.Now()

	return d.db.Model(&Habit{}).Where("id = ?", id).Updates(preRow).Error
}

func (d *Database) SetGroup(id uint, group string) error {

	preRow, err := d.GetHabitByID(id)
	if err != nil {
		panic(err)
	}

	preRow.Group = group

	return d.db.Model(&Habit{}).Where("id = ?", id).Updates(preRow).Error
}

func (d *Database) GetAllGroups() ([]string, error) {
	var groupNames []string
	if err := d.db.Raw("SELECT DISTINCT IFNULL(`group`, '') FROM habits").Scan(&groupNames).Error; err != nil {
		return nil, err
	}
	return groupNames, nil
}
