package main

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type HabitPublisher struct {
	Client mqtt.Client
	Topic  string
}

type HabitMessage struct {
	ObjectId                string `json:"object_id"`
	Name              		string `json:"name"`
	CommandTopic            string `json:"command_topic"`
	FriendlyName			string `json:"friendly_name"`
	Schema                  string `json:"schema"`
}

type Device struct {
	Identifiers string `json:"identifiers"`
	Name        string `json:"name"`
}

type HabitConfig struct {
	Name					string `json:"name"`
	DeviceClass             string `json:"device_name"`
	UniqueId 				string `json:"unique_id"`
	StateTopic 				string `json:"state_topic"`
	CommandTopic            string `json:"command_topic"`
	FriendlyName			string `json:"friendly_name"`
	Device       			Device `json:"device"`
	Schema                  string `json:"schema"`
}



func NewHabitPublisher(broker string, port int, topic string) *HabitPublisher {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s:%d", broker, port))
	client := mqtt.NewClient(opts)
	return &HabitPublisher{Client: client, Topic: topic}
}

func (p *HabitPublisher) Connect() {
	if token := p.Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (p *HabitPublisher) PublishHabits(habits []Habit) {
	for _, habit := range habits {
		fmt.Println("publishing:")
		fmt.Println(habit)
		fmt.Println("to:")
		thisTopic := toSnakeCase(p.Topic)
		topic := thisTopic + "/" + habit.Name;
		publishTopic := strings.ToLower(topic)


		configMessage := HabitConfig{
			Name: habit.Name,
			StateTopic: publishTopic + "/state",
			DeviceClass: "binary_sensor",
			UniqueId: habit.Name,
			CommandTopic: publishTopic + "/set",
			Device: Device{
				Identifiers: "hab",
				Name: "Habits",
			},
			Schema: "json",
		}


		fullTopic := "homeassistant/binary_sensor/" + publishTopic + "/config"
		fmt.Println("fullTopic (config):")
		fmt.Println(thisTopic)
		fmt.Println(topic)
		fmt.Println(publishTopic)
		fmt.Println(fullTopic)

		habitConfigJson, err := json.Marshal(configMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if token := p.Client.Publish(fullTopic, 0, false, habitConfigJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		habitMessageJson := HabitMessage{
			ObjectId: habit.Name,
			Name: habit.Name,
			FriendlyName: habit.Name,
			Schema: "json",
		}

		// Convert the habit to a JSON object.
		habitJson, err := json.Marshal(habitMessageJson)
		if err != nil {
			fmt.Println(err)
			continue
		}


		fullConfigTopic := "homeassistant/binary_sensor/" + publishTopic + "/state"
		fmt.Println("fullTopic (state):")
		fmt.Println(fullConfigTopic)

		if token := p.Client.Publish(fullConfigTopic, 0, false, habitJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}
}


func (p *HabitPublisher) Disconnect() {
	p.Client.Disconnect(250)
}