package controller

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

func Select(field string, items []string) (string, error) {
	l := fmt.Sprintf("Select %s", field)

	prompt := promptui.Select{
		Label: l,
		Items: items,
		Searcher: func(input string, index int) bool {
			return strings.Contains(strings.ToLower(items[index]), strings.ToLower(input))
		},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to select %s: %w", field, err)
	}
	return result, err
}
