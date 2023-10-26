package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type HabitPublisher struct {
	Client mqtt.Client
	Topic  string
}

type HabitMessage struct {
	ObjectId     string `json:"object_id"`
	Name         string `json:"name"`
	CommandTopic string `json:"command_topic"`
	FriendlyName string `json:"friendly_name"`
	Payload      string `json:"payload"`
	Schema       string `json:"schema"`
}

type Device struct {
	Identifiers string `json:"identifiers"`
	Name        string `json:"name"`
}

type MQTTConfig struct {
	Name         string `json:"name"`
	DeviceClass  string `json:"device_name"`
	UniqueId     string `json:"unique_id"`
	StateTopic   string `json:"state_topic"`
	CommandTopic string `json:"command_topic"`
	FriendlyName string `json:"friendly_name"`
	Device       Device `json:"device"`
	Schema       string `json:"schema"`
}

func (p *HabitPublisher) getDeviceName(class string, modifier string) (string, string) {
	var deviceName string = class + modifier
	isDevStr := GetEnvWithDefault("IS_DEV", "false")
	if isDevStr == "true" {
		deviceName = class + "-" + modifier + "_dev"
	} else {
		deviceName = class + "-" + modifier
	}

	deviceId := toSnakeCase(deviceName)

	return deviceName, deviceId
}

func (p *HabitPublisher) getMqttTopic(rawTopic string, posfix string) string {
	publishTopic := toSnakeCase(rawTopic)
	//publishTopic := strings.ToLower(thisTopic)

	fullTopic := "homeassistant/binary_sensor/" + p.Topic + "/" + publishTopic + "/" + posfix

	println(fullTopic)
	return fullTopic
}

func (p *HabitPublisher) getNoteMqttTopic(rawTopic string, posfix string) string {
	publishTopic := toSnakeCase(rawTopic)
	//publishTopic := strings.ToLower(thisTopic)

	fullTopic := "homeassistant/sensor/notes/" + publishTopic + "/" + posfix

	println(fullTopic)
	return fullTopic
}

func NewHabitPublisher() *HabitPublisher {
	broker := GetEnvWithDefault("MQTT_URL", "localhost")
	portStr := GetEnvWithDefault("MQTT_PORT", "1883")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Failed to convert port string to integer: %v", err)
	}
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s:%d", broker, port))
	client := mqtt.NewClient(opts)
	return &HabitPublisher{Client: client, Topic: "go_habits"}
}

func (p *HabitPublisher) Connect() {
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (p *HabitPublisher) DeleteHabit(habit Habit) {
	configTopic := p.getMqttTopic(habit.Name, "config")
	if token := p.Client.Publish(configTopic, 0, false, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

func (p *HabitPublisher) PublishHabits(habits []Habit) {
	for _, habit := range habits {

		stateTopic := p.getMqttTopic(habit.Name, "state")
		//	setTopic := p.getMqttTopic(habit.Name, "set")

		deviceName, deviceId := p.getDeviceName("habit", habit.Group)

		configMessage := MQTTConfig{
			Name:         habit.Name,
			StateTopic:   stateTopic,
			DeviceClass:  "binary_sensor",
			FriendlyName: habit.Name,
			UniqueId:     habit.Name,
			//		CommandTopic: setTopic,
			Device: Device{
				Identifiers: deviceId,
				Name:        deviceName,
			},
			Schema: "json",
		}

		configTopic := p.getMqttTopic(habit.Name, "config")

		fmt.Println("HABITSHabits\n\n\npublishing:")
		fmt.Println(configTopic)

		habitConfigJson, err := json.Marshal(configMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		data, err := json.MarshalIndent(configMessage, "", "    ")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if token := p.Client.Publish(configTopic, 0, true, habitConfigJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		payload := "OFF"
		payloadMap := map[string]string{}
		if habit.NeedsCompletion {
			payloadMap["state"] = "ON"
			payload = "ON"
		} else {
			payloadMap["state"] = "OFF"
		}

		fmt.Printf("configTopic: %s \n\ndata:\n", configTopic)
		fmt.Println(string(data))

		dataSet, err := json.MarshalIndent(payload, "", "    ")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("setTopic:")
		fmt.Printf("%s \n\n%v\n", stateTopic, dataSet)

		if token := p.Client.Publish(stateTopic, 0, true, payload); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}
}

func (p *HabitPublisher) PublishNotes(notes []Note) {
	for _, note := range notes {

		// add publishing of config message to link to HA

		stateTopic := p.getNoteMqttTopic(note.Title, "state")
		//	setTopic := p.getNoteMqttTopic(note.Title, "set")
		deviceName, deviceId := p.getDeviceName("notes", "")

		configMessage := MQTTConfig{
			Name:       note.Title,
			StateTopic: stateTopic,
			//		DeviceClass:  "sensor",
			FriendlyName: note.Title,
			UniqueId:     note.Title,
			//		CommandTopic: setTopic,
			Device: Device{
				Identifiers: deviceId,
				Name:        deviceName,
			},
			Schema: "json",
		}

		configTopic := p.getNoteMqttTopic(note.Title, "config")
		fmt.Println("NOTESNOTESNOTESNOTESNOTESNOTESNOTES:")

		noteConfigJson, err := json.Marshal(configMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		data, err := json.MarshalIndent(noteConfigJson, "", "    ")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("============================\n\nPublishNotes:configTopic: for - %s\nnoteConfigJson:\n", configTopic)
		fmt.Println(string(data))

		if token := p.Client.Publish(configTopic, 0, true, noteConfigJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		for _, note := range notes {

			noteJson, err := json.Marshal(note)
			if err != nil {
				fmt.Println(err)
				continue
			}

			data, err := json.MarshalIndent(noteJson, "", "    ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Printf("========================\n\nPublishNotes:noteTopic: for - %s\n\npayload:\n", stateTopic)
			fmt.Println(string(data))

			if token := p.Client.Publish(stateTopic, 0, true, noteJson); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
			}
		}
	}
}

func (p *HabitPublisher) DeleteNote(note Note) {
	noteTopic := p.getNoteMqttTopic(note.Title, "note")
	if token := p.Client.Publish(noteTopic, 0, true, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

var notesMap = make(map[string]Note) // to store received notes

func (p *HabitPublisher) SubscribeToAllNotes() {
	handler := func(client mqtt.Client, msg mqtt.Message) {
		var note Note
		if err := json.Unmarshal(msg.Payload(), &note); err != nil {
			fmt.Println(err)
			return
		}
		notesMap[msg.Topic()] = note // Store the received note in the map
	}
	wildcardTopic := "notes/#" // assuming all your notes are published under "notes/"
	if token := p.Client.Subscribe(wildcardTopic, 0, handler); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

func (p *HabitPublisher) Disconnect() {
	p.Client.Disconnect(250)
}
