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
	ObjectId                string `json:"object_id"`
	Name              		string `json:"name"`
	CommandTopic            string `json:"command_topic"`
	FriendlyName			string `json:"friendly_name"`
	Payload					string `json:"payload"`
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

func (p *HabitPublisher) getMqttTopic(rawTopic string, posfix string)string{
	publishTopic := toSnakeCase(rawTopic)
	//publishTopic := strings.ToLower(thisTopic)


	fullTopic := "homeassistant/binary_sensor/"+  p.Topic + "/"+ publishTopic +"/" + posfix

	println(fullTopic)
	return fullTopic
}


func NewHabitPublisher(topic string) *HabitPublisher {
	broker := "192.168.1.5"
	port := 1883
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s:%d", broker, port))
	client := mqtt.NewClient(opts)
	return &HabitPublisher{Client: client, Topic: topic}
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
		setTopic := p.getMqttTopic(habit.Name, "set")


		deviceName := "HabitsV2"
		deviceId := "hab"
		if(len(habit.Group) > 0){
			deviceName = habit.Group
			deviceId = toSnakeCase(habit.Group)
		}

		configMessage := HabitConfig{
			Name: habit.Name,
			StateTopic: stateTopic,
			DeviceClass: "binary_sensor",
			UniqueId: habit.Name,
			CommandTopic: setTopic,
			Device: Device{
				Identifiers: deviceId,
				Name: deviceName,
			},
			Schema: "json",
		}

		configTopic := p.getMqttTopic(habit.Name, "config")


		fmt.Println("publishing:")
		fmt.Println(configTopic)


		habitConfigJson, err := json.Marshal(configMessage)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("habitConfigJson:")
		fmt.Print(configMessage)

		if token := p.Client.Publish(configTopic, 0, false, habitConfigJson); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}

		//habitMessageJson := HabitMessage{
		//	ObjectId: habit.Name,
		//	Name: habit.Name,
		//	FriendlyName: habit.Name,
		//	Payload: "on",
		//	Schema: "json",
		//}
//
		//// Convert the habit to a JSON object.
		//habitJson, err := json.Marshal(habitMessageJson)
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}

		payload := "OFF"
		if(habit.NeedsCompletion){
			payload = "ON"
		}

		if token := p.Client.Publish(stateTopic, 0, true, payload); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}
}


func (p *HabitPublisher) Disconnect() {
	p.Client.Disconnect(250)
}