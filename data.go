package main

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ResetFrequency string

const (
	Daily   ResetFrequency = "daily"
	Weekly  ResetFrequency = "weekly"
	Monthly ResetFrequency = "monthly"
	// Add more frequencies as needed
)

type Habit struct {
	ID             uint           `gorm:"primarykey"`
	Name           string         `gorm:"size:255"`
	ResetFrequency ResetFrequency `gorm:"type:varchar(10)"`
	IsActive       bool           `gorm:"type:boolean"`
	LastComplete   time.Time      `gorm:"type:datetime"`
}
type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(sqlite.Open("habits.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Habit{}); err != nil {
		return nil, err
	}

	return &Database{
		db: db,
	}, nil
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
