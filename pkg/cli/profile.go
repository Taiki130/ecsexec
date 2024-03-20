package cli

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

func SelectProfile() (string, error) {
	l := "Select Profile"

	fname := config.DefaultSharedConfigFilename()
	f, err := ini.Load(fname)
	if err != nil {
		return "", err
	}

	var profiles []string
	for _, v := range f.Sections() {
		if len(v.Keys()) != 0 {
			profile := getProfileFromIniSection(v.Name())
			profiles = append(profiles, profile)
		}
	}

	prompt := promptui.Select{
		Label: l,
		Items: profiles,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}

func getProfileFromIniSection(section string) string {
	return strings.Split(section, " ")[1]
}
