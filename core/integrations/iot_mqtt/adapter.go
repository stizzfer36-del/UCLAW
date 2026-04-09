// Package iot_mqtt wraps an MQTT broker for IoT device control.
// Uses the mosquitto_pub / mosquitto_sub CLI or a Go MQTT client.
//
// For production replace mosquitto_pub calls with:
//   github.com/eclipse/paho.mqtt.golang
package iot_mqtt

import (
	"fmt"
	"os/exec"
)

// Config holds MQTT broker connection details.
type Config struct {
	Broker string // e.g. localhost
	Port   int    // e.g. 1883
}

// Publish sends a payload to a topic (maps to iot_toggle_device / iot_read_sensor).
func (c *Config) Publish(topic, payload string) error {
	cmd := exec.Command(
		"mosquitto_pub",
		"-h", c.Broker,
		"-p", fmt.Sprintf("%d", c.Port),
		"-t", topic,
		"-m", payload,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("iot_mqtt: publish %s: %w: %s", topic, err, out)
	}
	return nil
}

// Subscribe reads one retained message from a topic.
func (c *Config) Subscribe(topic string) (string, error) {
	out, err := exec.Command(
		"mosquitto_sub",
		"-h", c.Broker,
		"-p", fmt.Sprintf("%d", c.Port),
		"-t", topic,
		"-C", "1", // receive exactly 1 message
	).Output()
	if err != nil {
		return "", fmt.Errorf("iot_mqtt: sub %s: %w", topic, err)
	}
	return string(out), nil
}
