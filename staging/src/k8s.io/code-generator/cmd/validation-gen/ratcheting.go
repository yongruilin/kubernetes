package main

import (
	"fmt"

	"k8s.io/gengo/v2"
)

const ratchetingTag = "k8s:ratcheting"

func extractRatchetingOptions(comments []string) (RatchetingOptions, error) {
	tags, err := gengo.ExtractFunctionStyleCommentTags("+", []string{ratchetingTag}, comments)
	if err != nil {
		return RatchetingOptions{}, err
	}

	if len(tags) == 0 {
		return RatchetingOptions{}, nil
	}

	for _, tag := range tags[ratchetingTag] {
		if tag.Name != ratchetingTag {
			continue
		}
		if tag.Value == "disabled" {
			return RatchetingOptions{NoRatcheting: true}, nil
		}
		if tag.Value == "1" {
			return RatchetingOptions{RatchOption: 1}, nil
		}
		if tag.Value == "2" {
			return RatchetingOptions{RatchOption: 2}, nil
		}
		return RatchetingOptions{}, fmt.Errorf("invalid ratcheting option: %s", tag.Value)
	}

	return RatchetingOptions{}, nil
}

type RatchetingOptions struct {
	NoRatcheting bool
	RatchOption  int
}
