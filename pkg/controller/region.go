package controller

import (
	"github.com/Taiki130/ecsexec/pkg/constants"
	"github.com/manifoldco/promptui"
)

func SelectRegion() (string, error) {
	l := "Select region"
	prompt := promptui.Select{
		Label: l,
		Items: constants.AWS_VALID_REGIONS,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
