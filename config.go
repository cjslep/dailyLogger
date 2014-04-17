package dailyLogger

import (
	"io/ioutil"
	"os"
	"gopkg.in/v1/yaml"
)

type LoggingConfig struct {
	LogFileName string
	DirectoryLogPath string
	FilePermissions os.FileMode
	FolderPermissions os.FileMode
}

func LoadLoggingConfig(filename string) (*LoggingConfig, error) {
	cont, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := LoggingConfig{}
	err = yaml.Unmarshal(cont, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
