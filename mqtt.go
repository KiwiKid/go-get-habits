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
	ObjectId                    string `json:"object_id"`
	Name              		string `json:"name"`
	StateTopic 				string `json:"command_topic"`
	CommandTopic            string `json:"command_topic"`
	FriendlyName			string `json:friendly_name`
	Device					string `json:device`
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
		topic := toSnakeCase(p.Topic) + "/" + habit.Name;
		publishTopic := strings.ToLower(topic)
		fmt.Println("to:")
		fmt.Println(topic)


		habitMessage := HabitMessage{
			ObjectId: habit.Name,
			Name: habit.Name,
			StateTopic: publishTopic + "/state",
			CommandTopic: publishTopic + "/set",
			FriendlyName: habit.Name,
			Device: "habits",
			Schema: "json",
		}

		// Convert the habit to a JSON object.
		habitJson, err := json.Marshal(habitMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if token := p.Client.Publish(publishTopic, 0, false, habitJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}
}


func (p *HabitPublisher) Disconnect() {
	p.Client.Disconnect(250)
}