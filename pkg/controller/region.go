package controller

import (
	"fmt"
	"strings"

	"github.com/Taiki130/ecsexec/pkg/constants"
	"github.com/manifoldco/promptui"
)

func SelectRegion() (string, error) {
	l := "Select region"
	prompt := promptui.Select{
		Label: l,
		Items: constants.AWS_VALID_REGIONS,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(constants.AWS_VALID_REGIONS[index]), strings.ToLower(input))
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select region: %w", err)
	}

	return result, nil
}
