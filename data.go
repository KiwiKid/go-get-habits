package main

import (
	"errors"
	"fmt"
	"log"
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
	ResetValue 	   int			  `gorm:"type:int"`
	Group	 	   string	      `gorm:"size:255"`
	IsActive       bool           `gorm:"type:boolean"`
	LastComplete   time.Time      `gorm:"type:datetime"`
	NeedsCompletion bool	      `gorm:"type:boolean"`
}
type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, func(), error) {
	db, err := gorm.Open(sqlite.Open("habits.db"), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	if err := db.AutoMigrate(&Habit{}); err != nil {
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

	// Initialize database
	db, closeDB, err := NewDatabase()
	if err != nil {
		log.Printf("Error initializing database: %s", err)
	}

	defer closeDB()

	// Fetch all active habits
	rows, err := db.GetAllHabits(true)
	if err != nil {
		log.Printf("Error fetching habits: %s", err)
	}

	// Check each habit's status
	for _, habit := range rows {
		fmt.Println("Checking :"+habit.Name)

		if needsCompletion(habit) {
			fmt.Println("ACTION_NEEDED")

			habit.NeedsCompletion = true;

			err := db.SetHabitNeedCompletion(habit.ID, true)

			if(err != nil){
				log.Printf("ERROR ERROR saving check habit: %s", err)
				return err;
			}

			// Handle what to do if habit needs completion. For instance, notify the user.
		}else{
			// This extended update might not be needed always(just after config update)
			habit.NeedsCompletion = false;

			err := db.SetHabitNeedCompletion(habit.ID, false)

			if(err != nil){
				log.Printf("ERROR ERROR saving check habit: %s", err)
				return err;
			}
			fmt.Println("ALL GOOD"+habit.Name)
		}
	}
	return nil
}


func (d *Database) CreateHabit(h *Habit) error {
	return d.db.Create(h).Error
}

func (d *Database) GetAllHabits(isActive ...bool) ([]Habit, error) {
    var habits []Habit

    // Check if the database is initialized
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

func (d *Database) EditHabit(id uint, updatedHabit *Habit) error {
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