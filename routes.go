package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jritsema/gotoolbox/web"
)

// GET /company/add
func habitAdd(r *http.Request) *web.Response {
	fmt.Println("habitAdd:")

	db, closeDB, err := NewDatabase()
	if err != nil {
		panic(err)
	}
	defer closeDB()
	habit := &Habit{
		Name:           "Exercise",
		ResetFrequency: Daily,
	}
	if err := db.CreateHabit(habit); err != nil {
		panic(err)
	}
	habits, err := db.GetAllHabits()
	return web.HTML(http.StatusOK, html, "habit-add.html", habits, nil)
}

func habitCompleted(r *http.Request) *web.Response {

	switch r.Method {
	case http.MethodPost:
		fmt.Println("habitCompleted:")
		db, closeDB, err := NewDatabase()
		if err != nil {
			panic(err)
		}
		defer closeDB()
		id, _ := web.PathLast(r)
		var idInt uint
		_, idError := fmt.Sscanf(id, "%d", &idInt)
		if idError != nil {
			fmt.Println("Error:", err)

			return web.DataJSON(http.StatusNotFound, nil, map[string]string{"Content-Type": "application/json"})
		}


		db.CompleteHabit(idInt)

		row, err := db.GetHabitByID(idInt)

		if err != nil {
			fmt.Println("Error:", err)
			if err != nil {
				panic(err)
			}
		}
		return web.DataJSON(http.StatusOK, row, map[string]string{"Content-Type": "application/json"})
	}
	return web.Empty(http.StatusNotImplemented)
}

func publish(r *http.Request) *web.Response {
	fmt.Println("publish start")

	switch r.Method {
	case http.MethodPost:
		r.ParseForm()
		
		topic := r.Form.Get("Topic")
		println("Saving "+topic)

		db, closeDB, err := NewDatabase()
		if err != nil {
			panic(err)
		}
		defer closeDB()
		rows, err := db.GetAllHabits(true)
		if err != nil {
			panic(err)
		}

		publisher := NewHabitPublisher(topic)

		// Connect to the MQTT broker.
		publisher.Connect()
		defer publisher.Disconnect()

		fmt.Println("Publishing, Topic:")
		fmt.Println(topic)
		fmt.Println("Publishing, Message:")
		fmt.Println(rows)
		// Publish the habits.
		publisher.PublishHabits(rows)

		data := map[string]interface{}{
			"topic": topic,
		}		

		return  web.HTML(http.StatusOK, html, "publish.html", data, nil)
	case http.MethodGet:
		data := map[string]interface{}{
			"topic": "go_habits",
		}

		return  web.HTML(http.StatusOK, html, "publish.html", data,  nil)
	}
	return web.Empty(http.StatusNotImplemented)
}



// GET /company
// GET /company/{id}
// DELETE /company/{id}
// PUT /company/{id}
// POST /company
func habits(r *http.Request) *web.Response {
	fmt.Println("habits start")

	id, _ := web.PathLast(r)
	db, closeDB, err := NewDatabase()
	var idInt uint
	_, idError := fmt.Sscanf(id, "%d", &idInt)
	
	if err != nil {
		fmt.Println("Error:", err)
		return web.Empty(http.StatusInternalServerError)
	}
	defer closeDB()

	switch r.Method {

	case http.MethodDelete:
		fmt.Println("Delete start")


		if idError != nil {
			fmt.Println("Error:", err)
			rows, err := db.GetAllHabits()
			if err != nil {
				panic(err)
			}
			return  web.HTML(http.StatusNotFound, html, "habits.html", rows, nil)
		}
		err = db.DeleteHabitByID(idInt)
		if err != nil {
			panic(err)
		}
		publisher := NewHabitPublisher("go_habits")
		publisher.Connect()
		defer publisher.Disconnect()
		rows, err := db.GetAllHabits()
		if err != nil {
			panic(err)
		}
		fmt.Println("Delete complete")
		return web.HTML(http.StatusOK, html, "habits.html", rows, nil)

	//cancel
	case http.MethodGet:
		fmt.Println("get start")
		if idError != nil {
			row, err := db.GetHabitByID(idInt)
			if err != nil {
				panic(err)
			}
			fmt.Println("returning row")
			return web.HTML(http.StatusOK, html, "row.html", row, nil)
		} else {
			//cancel add
			row, err := db.GetHabitByID(idInt)
			if err != nil {
				panic(err)
			}
			fmt.Println("returning list")
			return web.HTML(http.StatusOK, html, "row.html", row, nil)
		}

	//save edit
	case http.MethodPut:
		println("http.MethodPut")
		row, err := db.GetHabitByID(idInt)
		if err != nil {
			panic(err)
		}
		r.ParseForm()
		
		row.Name = r.Form.Get("Name")
		row.ResetFrequency = ResetFrequency(r.Form.Get("ResetFrequency"))
		resetValue, err := strconv.Atoi(r.Form.Get("ResetValue"))
		if err != nil {
			// handle error
		}
		row.ResetValue = resetValue
		row.Group = r.Form.Get("Group")
		row.IsActive = len(r.Form.Get("IsActive")) > 0
		println("Saving")
		db.EditHabit(idInt, row)

		return web.HTML(http.StatusOK, html, "row.html", row, nil)

	case http.MethodPost:
		r.ParseForm()
	
		name := r.Form.Get("Name")
		resetFrequency := ResetFrequency(r.Form.Get("ResetFrequency"))
		resetValue, err := strconv.Atoi(r.Form.Get("ResetValue"))
		if err != nil {
			// handle error
		}
		group := r.Form.Get("Group")
		isActive := len(r.Form.Get("IsActive")) > 0
	
		newHabit := Habit{
			Name:           name,
			ResetFrequency: resetFrequency,
			ResetValue:     resetValue,
			Group:          group,
			IsActive:       isActive,
			// Add any other required fields if they exist.
		}
	
		db.CreateHabit(&newHabit)  // I'm assuming the method should be called with newHabit instead of 'h'
		rows, err := db.GetAllHabits()
		if err != nil {
			panic(err)
		}
		fmt.Println("Create complete")
		return web.HTML(http.StatusOK, html, "habits.html", rows, nil)
	}
	

	return web.Empty(http.StatusNotImplemented)
}


// Delete -> DELETE /company/{id} -> delete, companys.html

// Edit   -> GET /company/edit/{id} -> row-edit.html
// Save   ->   PUT /company/{id} -> update, row.html
// Cancel ->	 GET /company/{id} -> nothing, row.html

// Add    -> GET /company/add/ -> companys-add.html (target body with row-add.html and row.html)
// Save   ->   POST /company -> add, companys.html (target body without row-add.html)
// Cancel ->	 GET /company -> nothing, companys.html

func index(r *http.Request) *web.Response {
	fmt.Println("index:")

	db, closeDB, err := NewDatabase()
	rows, err := db.GetAllHabits()
	if err != nil {
		panic(err)
	}
	defer closeDB()
	return web.HTML(http.StatusOK, html, "index.html", rows, nil)
}

// GET /company/add
//func companyAdd(r *http.Request) *web.Response {
//	return web.HTML(http.StatusOK, html, "company-add.html", data, nil)
//}

 // /GET habit/edit/{id}
func habitEdit(r *http.Request) *web.Response {
	fmt.Println("index:")

	db, closeDB, err := NewDatabase()
	if err != nil {
		panic(err)
	}
	defer closeDB()
	id, _ := web.PathLast(r)
	var idInt uint
	_, idError := fmt.Sscanf(id, "%d", &idInt)
	
	if idError != nil {
		fmt.Println("Error:", err)
	}
	row, err := db.GetHabitByID(idInt)
	if err != nil {
		panic(err)
	}
	return web.HTML(http.StatusOK, html, "row-edit.html", row, nil)
}

// GET /company
// GET /company/{id}
// DELETE /company/{id}
// PUT /company/{id}
// POST /company
//func companies(r *http.Request) *web.Response {
//	id, segments := web.PathLast(r)
//	switch r.Method {
//
//	case http.MethodDelete:
//		deleteCompany(id)
//		return web.HTML(http.StatusOK, html, "companies.html", data, nil)
//
//	//cancel
//	case http.MethodGet:
//		if segments > 1 {
			//cancel edit
	//		row := getCompanyByID(id)
	//		return web.HTML(http.StatusOK, html, "row.html", row, nil)
//		} else {
			//cancel add
	//		return web.HTML(http.StatusOK, html, "companies.html", data, nil)
//		}
//
//	//save edit
//	case http.MethodPut:
//		row := getCompanyByID(id)
//		r.ParseForm()
//		row.Company = r.Form.Get("company")
//		row.Contact = r.Form.Get("contact")
//		row.Country = r.Form.Get("country")
//		updateCompany(row)
//		return web.HTML(http.StatusOK, html, "row.html", row, nil)
//
//	//save add
//	case http.MethodPost:
//		row := Company{}
//		r.ParseForm()
//		row.Company = r.Form.Get("company")
//		row.Contact = r.Form.Get("contact")
//		row.Country = r.Form.Get("country")
//		addCompany(row)
//		return web.HTML(http.StatusOK, html, "companies.html", data, nil)
//	}
//
//	return web.Empty(http.StatusNotImplemented)
//}
