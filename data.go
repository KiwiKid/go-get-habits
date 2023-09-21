package main

import (
	"fmt"
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


func (d *Database) CreateHabit(h *Habit) error {
	return d.db.Create(h).Error
}

func (d *Database) GetAllHabits(isActive ...bool) ([]Habit, error) {
    var habits []Habit
    db := d.db
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
