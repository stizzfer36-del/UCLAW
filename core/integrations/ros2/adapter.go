// Package ros2 wraps ROS2 CLI tools (ros2 service call, ros2 topic pub).
// ROS2: https://github.com/ros2/ros2
//
// Prerequisites: ROS2 environment sourced (source /opt/ros/humble/setup.bash)
// Topic/service whitelists enforced by UCLAW policy engine before calling.
package ros2

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CallService calls a ROS2 service: ros2 service call <service> <type> <args>.
func CallService(service, msgType, argsYAML string) (string, error) {
	cmd := exec.Command("ros2", "service", "call", service, msgType, argsYAML)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ros2: call %s: %w: %s", service, err, errBuf.String())
	}
	return strings.TrimSpace(out.String()), nil
}

// PublishTopic publishes a single message to a ROS2 topic.
func PublishTopic(topic, msgType, msgYAML string) error {
	cmd := exec.Command("ros2", "topic", "pub", "--once", topic, msgType, msgYAML)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ros2: pub %s: %w: %s", topic, err, errBuf.String())
	}
	return nil
}
