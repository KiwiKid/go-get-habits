package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jritsema/gotoolbox"
	"github.com/jritsema/gotoolbox/web"
)

var (
	//go:embed templates/*
	templateFS embed.FS

	//go:embed css/output.css
	css embed.FS

	//parsed templates
	html *template.Template
)

/*
func keepCheckingHabitStatus() {
	fmt.Println("keepCheckingHabitStatus")
	//ticker := time.NewTicker(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			log.Printf("checking")

			// Initialize database
			db, closeDB, err := NewDatabase()
			if err != nil {
				log.Printf("Error initializing database: %s", err)
				continue
			}

			defer closeDB()

			// Fetch all active habits
			rows, err := db.GetAllHabits(true)
			if err != nil {
				log.Printf("Error fetching habits: %s", err)
				continue
			}

			// Check each habit's status
			for _, habit := range rows {
				fmt.Println("keepCheckingHabitStatus:"+habit.Name)

				if needsCompletion(habit) {
					fmt.Println("keepCheckingHabitStatus:ACTION_NEEDED"+habit.Name)

					habit.NeedsCompletion = true;

					err := db.EditHabit(habit.ID, &habit)

					if(err != nil){
						log.Printf("Error Editing habit: %s", err)

					}

					// Handle what to do if habit needs completion. For instance, notify the user.
				}else{
					fmt.Println("keepCheckingHabitStatus:GOOD"+habit.Name)
				}
			}
		}
	}
}*/

func needsResetBasedCompletion(h Habit) bool {
	switch h.ResetFrequency {
	case Daily:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24
	case Weekly:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24*7
	case Monthly:
		// Note: Using 30 days for a month is a simplification, some months have 28, 29, or 31 days.
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24*30
	case Minutes:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Minute
	case Hourly:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour
	default:
		log.Printf("Unknown reset frequency: %s", h.ResetFrequency)
		return false
	}
}

func needsCompletion(h Habit) bool {
	log.Printf("Checking completion for habit: %+v\n", h)

	if !needsResetBasedCompletion(h) {
		log.Printf("Habit does not need reset based completion.")
		return false
	}
	var isAfterStartTime bool
	var isBeforeEndTime bool
	// Ensure the current time is within the defined start and end times for the habit.
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), h.StartHour, h.StartMinute, 0, 0, currentTime.Location())
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), h.EndHour, h.EndMinute, 0, 0, currentTime.Location())

	log.Printf("Current time: %v, Start time: %v, End time: %v", currentTime, startTime, endTime)

	isAfterStartTime = currentTime.After(startTime)
	isBeforeEndTime = currentTime.Before(endTime)
	log.Printf("Is current time after start time? %v. Is current time before end time? %v.", isAfterStartTime, isBeforeEndTime)

	noStartTime := h.StartHour == 0 && h.StartMinute == 0
	noEndTime := h.EndHour == 0 && h.EndMinute == 0
	log.Printf("Is there a start and end time set? noStartTime: %v noEndTime: %v", noStartTime, noEndTime)
	return (noStartTime || isAfterStartTime) && (isBeforeEndTime || noEndTime)
}

func main() {

	//exit process immediately upon sigterm
	handleSigTerms()

	//parse templates
	var err error
	html, err = ParseTemplates(templateFS)
	if err != nil {
		panic(err)
	}

	//add routes
	router := http.NewServeMux()

	router.Handle("/habit/complete", web.Action(habitCompleted))
	router.Handle("/habit/complete/", web.Action(habitCompleted))

	router.Handle("/only-check", web.Action(check))

	router.Handle("/publish", web.Action(publish))
	router.Handle("/check", web.Action(checkAndPublish))

	router.Handle("/notes", web.Action(notes))
	router.Handle("/notes/", web.Action(notes))

	router.Handle("/habit/add", web.Action(habitAdd))
	router.Handle("/habit/add/", web.Action(habitAdd))

	router.Handle("/habit/edit", web.Action(habitEdit))
	router.Handle("/habit/edit/", web.Action(habitEdit))
	router.Handle("/habit/edit/group", web.Action(habitGroup))

	router.Handle("/habit", web.Action(habits))
	router.Handle("/habit/", web.Action(habits))

	router.Handle("/css/output.css", http.FileServer(http.FS(css)))
	//router.Handle("/company/edit", web.Action(companyEdit))
	//router.Handle("/company/edit/", web.Action(companyEdit))

	//router.Handle("/company", web.Action(companies))
	//router.Handle("/company/", web.Action(companies))

	router.Handle("/", web.Action(index))
	router.Handle("/index.html", web.Action(index))

	router.Handle("/preview", web.Action(preview))

	//logging/tracing
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	middleware := tracing(nextRequestID)(logging(logger)(router))

	port := gotoolbox.GetEnvWithDefault("PORT", "8122")
	logger.Println("listening on http://localhost:" + port)

	if err := http.ListenAndServe(":"+port, middleware); err != nil {
		logger.Println("http.ListenAndServe():", err)
		os.Exit(1)
	}
	logger.Println("Down here?")

	//keepCheckingHabitStatus()

}

func handleSigTerms() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("received SIGTERM, exiting")
		os.Exit(1)
	}()
}
