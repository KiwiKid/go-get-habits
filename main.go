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

func keepCheckingHabitStatus() {
	ticker := time.NewTicker(5 * time.Minute)

	for {
		select {
		case <-ticker.C:
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
				if needsCompletion(habit) {
					
					habit.NeedsCompletion = true;

					err := db.EditHabit(habit.ID, &habit)

					if(err != nil){
						log.Printf("Error Editing habit: %s", err)

					}

					// Handle what to do if habit needs completion. For instance, notify the user.
				}
			}
		}
	}
}

func needsCompletion(h Habit) bool {
	// Based on your habit's reset frequency and reset value,
	// determine if the habit needs to be completed again
	switch h.ResetFrequency {
	case Daily:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24
	case Weekly:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24*7
	case Monthly:
		return time.Since(h.LastComplete) > time.Duration(h.ResetValue)*time.Hour*24*30
	default:
		log.Printf("Unknown reset frequency: %s", h.ResetFrequency)
		return false
	}
}

func main() {

	//exit process immediately upon sigterm
	handleSigTerms()

	//parse templates
	var err error
	html, err = web.TemplateParseFSRecursive(templateFS, ".html", true, nil)
	if err != nil {
		panic(err)
	}

	//add routes
	router := http.NewServeMux()

	router.Handle("/habit/complete", web.Action(habitCompleted))
	router.Handle("/habit/complete/", web.Action(habitCompleted))

	router.Handle("/habit/add", web.Action(habitAdd))
	router.Handle("/habit/add/", web.Action(habitAdd))

	router.Handle("/habit/edit", web.Action(habitEdit))
	router.Handle("/habit/edit/", web.Action(habitEdit))

	router.Handle("/habit", web.Action(habits))
	router.Handle("/habit/", web.Action(habits))


	router.Handle("/publish", web.Action(publish))

	router.Handle("/css/output.css", http.FileServer(http.FS(css)))
	//router.Handle("/company/edit", web.Action(companyEdit))
	//router.Handle("/company/edit/", web.Action(companyEdit))

	//router.Handle("/company", web.Action(companies))
	//router.Handle("/company/", web.Action(companies))

	router.Handle("/", web.Action(index))
	router.Handle("/index.html", web.Action(index))

	//logging/tracing
	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	middleware := tracing(nextRequestID)(logging(logger)(router))

	port := gotoolbox.GetEnvWithDefault("PORT", "8080")
	logger.Println("listening on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, middleware); err != nil {
		logger.Println("http.ListenAndServe():", err)
		os.Exit(1)
	}

	keepCheckingHabitStatus()
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
