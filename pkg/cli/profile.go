package cli

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/manifoldco/promptui"
	"gopkg.in/ini.v1"
)

func selectProfile() (string, error) {
	l := "Select profile"

	fname := config.DefaultSharedConfigFilename()
	profiles, err := getProfilesFromIni(fname)
	if err != nil {
		return "", err
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

func getProfilesFromIni(fname string) (profiles []string, err error) {
	f, err := ini.Load(fname)
	if err != nil {
		return profiles, err
	}

	for _, v := range f.Sections() {
		if len(v.Keys()) != 0 {
			profile := getProfileFromIniSection(v.Name())
			profiles = append(profiles, profile)
		}
	}
	return
}

func getProfileFromIniSection(section string) string {
	return strings.Split(section, " ")[1]
}
