package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"k8s.io/klog/v2"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomString(num int) string {
	b := make([]rune, num)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] // #nosec
	}

	return string(b)
}

func RunOSCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func RunBatchOSCommand(cmds [][]string) error {
	for _, cmd := range cmds {
		err := RunOSCommand(cmd[0], cmd[1:]...)
		if err != nil {
			klog.Error("exec command ", cmd, " error, ", err)
			return err
		}
	}

	return nil
}

func ReloadSystemService(service string) error {
	batchCmd := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "restart", service},
		{"systemctl", "enable", service},
		{"systemctl", "--no-pager", "status", service},
	}

	return RunBatchOSCommand(batchCmd)
}

func WritePropertiesFile(properties map[string]string, filename string) error {
	buf := ""
	for k, v := range properties {
		buf += fmt.Sprintf("%s=%s\n", k, v)
	}

	return os.WriteFile(filename, []byte(buf), 0644)
}
