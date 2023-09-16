package main

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type HabitPublisher struct {
	Client mqtt.Client
	Topic  string
}

type HabitMessage struct {
	Name                    string `json:"name"`
	StateTopic              string `json:"state_topic"`
	CommandTopic            string `json:"command_topic"`
	BrightnessStateTopic    string `json:"brightness_state_topic"`
	BrightnessCommandTopic  string `json:"brightness_command_topic"`
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

		habitMessage := HabitMessage{
			Name: habit.Name,
			StateTopic: p.Topic + "/" + habit.Name + "/state",
			CommandTopic: p.Topic + "/" + habit.Name + "/set",
			Schema: "json",
		}

		// Convert the habit to a JSON object.
		habitJson, err := json.Marshal(habitMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if token := p.Client.Publish(p.Topic, 0, false, habitJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}
}


func (p *HabitPublisher) Disconnect() {
	p.Client.Disconnect(250)
}