package controller

import "github.com/manifoldco/promptui"

func PromptRegion() (string, error) {
	l := "Enter region"
	prompt := promptui.Prompt{
		Label: l,
	}
	result, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return result, nil
}
