package main

import "k8s.io/gengo/v2"

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
	}

	return RatchetingOptions{}, nil
}

type RatchetingOptions struct {
	NoRatcheting bool
}
