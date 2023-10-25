package main

import (
	"fmt"
	"html/template"
	"io/fs"
	"time"

	"github.com/jritsema/gotoolbox/web" // Assuming this is where web.TemplateParseFSRecursive is located
)

func formatDate(t time.Time) string {
	day := t.Day()
	suffix := "th"

	switch day {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	}

	monthName := t.Month().String()
	hourMinute := t.Format("3:04 pm")

	return fmt.Sprintf("%s %d%s @ %s", monthName, day, suffix, hourMinute)
}

func relativeTime(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("now - %d secs ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else if duration < 48*time.Hour {
		return "yesterday"
	} else if duration < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (7 * 24))
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 18*30*24*time.Hour {
		months := int(duration.Hours() / (30 * 24))
		return fmt.Sprintf("%d months ago", months)
	} else {
		return "ages ago"
	}
}

/*

// Custom function to format date
func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 -0700", dateStr)
	if err != nil {
		return dateStr // return the original string in case of an error
	}
	return t.Format("2006-01-02 15:04:05")
}*/
// ParseTemplates parses the templates from the provided file system and attaches custom functions
func ParseTemplates(templateFS fs.FS) (*template.Template, error) {
	funcs := template.FuncMap{
		"formatDate":   formatDate,
		"relativeTime": relativeTime,
	}

	// Parse the templates and attach the custom functions
	return web.TemplateParseFSRecursive(templateFS, ".html", true, funcs)
}
