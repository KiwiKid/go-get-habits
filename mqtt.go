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
		setTopic := p.getMqttTopic(habit.Name, "set")

		isDevStr := GetEnvWithDefault("IS_DEV", "false")
		isDev := false
		if isDevStr == "true" {
			isDev = true
		}
	

		deviceName := GetEnvWithDefault("HA_DEVICE_NAME", "")
		deviceId := "habV4"
		if(len(habit.Group) > 0){
			if(isDev){
				deviceName = "Habits " + habit.Group + " [dev]"
			}else{
				deviceName = "Habits "+ habit.Group
			}
			deviceId = toSnakeCase(deviceName)
		}

		configMessage := HabitConfig{
			Name: habit.Name,
			StateTopic: stateTopic,
			DeviceClass: "binary_sensor",
			FriendlyName: habit.Name,
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