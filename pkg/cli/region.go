package cli

import (
	"github.com/manifoldco/promptui"
)

func promptRegion() (string, error) {
	l := "Enter Region"
	prompt := promptui.Prompt{
		Label: l,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}
