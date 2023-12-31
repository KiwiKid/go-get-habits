package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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

func (p *HabitPublisher) publishMQTTMessage(topic string, message interface{}) error {
	jsonMessage, err := json.MarshalIndent(message, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	token := p.Client.Publish(topic, 0, true, jsonMessage)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("error publishing message to topic %s: %v", topic, token.Error())
	}

	return nil
}

func (p *HabitPublisher) getMqttTopic(rawTopic string, postfix string) string {
	publishTopic := lowerAndReplaceSpaces(rawTopic)
	//publishTopic := strings.ToLower(thisTopic)

	fullTopic := "homeassistant/binary_sensor/" + p.Topic + "/" + publishTopic + "/" + postfix

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
	configTopic := habit.getMQTTTopic(p.Topic, "config")
	if token := p.Client.Publish(configTopic, 0, false, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	stateTopic := habit.getMQTTTopic(p.Topic, "state")
	if token := p.Client.Publish(stateTopic, 0, false, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

func (p *HabitPublisher) PublishHabits(habits []Habit) {
	for _, habit := range habits {

		stateTopic := habit.getMQTTTopic(p.Topic, "state")

		deviceName, deviceId := habit.getDeviceName()

		configMessage := MQTTConfig{
			Name:         habit.Name,
			StateTopic:   stateTopic,
			DeviceClass:  "binary_sensor",
			FriendlyName: habit.Name,
			UniqueId:     habit.getUniqueId(),
			//		CommandTopic: setTopic,
			Device: Device{
				Identifiers: deviceId,
				Name:        deviceName,
			},
			Schema: "json",
		}

		configTopic := habit.getMQTTTopic(p.Topic, "config")

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
		// Construct and publish config message
		stateTopic := note.getMQTTTopic("state")
		deviceName, deviceId := note.getDeviceName()

		configMessage := MQTTConfig{
			Name:         note.Title,
			StateTopic:   stateTopic,
			FriendlyName: note.Title,
			UniqueId:     note.Title,
			Device: Device{
				Identifiers: deviceId,
				Name:        deviceName,
			},
			Schema: "json",
		}

		configTopic := note.getMQTTTopic("config")
		fmt.Println("Publishing MQTT config for note:", note.Title)

		if err := p.publishMQTTMessage(configTopic, configMessage); err != nil {
			fmt.Println(err)
			continue
		}
	}

	// Publish state messages for relevant notes
	currentDay := time.Now().Weekday().String()
	for _, note := range notes {
		if !strings.Contains(strings.ToLower(note.OnlyRelevantOnDay), strings.ToLower(currentDay)) {
			continue // Skip this note if not relevant today
		}

		stateTopic := note.getMQTTTopic("state")
		fmt.Println("Publishing MQTT state for note:", note.Title)

		if err := p.publishMQTTMessage(stateTopic, note); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func (p *HabitPublisher) DeleteNote(note Note) {
	noteTopic := note.getMQTTTopic("state")
	if token := p.Client.Publish(noteTopic, 0, true, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	} else {
		fmt.Printf("deleted: %s\n", noteTopic)
	}

	noteConfigTopic := note.getMQTTTopic("config")
	if token := p.Client.Publish(noteConfigTopic, 0, true, ""); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	} else {
		fmt.Printf("deleted: %s\n", noteConfigTopic)
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
