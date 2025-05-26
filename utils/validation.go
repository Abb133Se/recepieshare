package utils

import (
	"errors"
	"strconv"
)

var validSortOptions = map[string]bool{
	"like_desc": true,
	"like_asc":  true,
	"":          true,
}

func ValidateEntityID(id string) (uint, error) {
	parsedID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return 0, errors.New("invalid ID format")
	}
	return uint(parsedID), nil
}

func ValidateOffLimit(limit, offset string) (int, int, error) {
	var resultedLimit, resultedOffset int

	if limit == "" && offset == "" {
		return 0, 0, errors.New("invalid limit and offset values")
	}
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			if l < 0 {
				return 0, 0, errors.New("negative limit, must be a non-negative integer")
			}

			resultedLimit = l
		} else {
			return 0, 0, errors.New("invalid limit value")
		}
	}
	if offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			if o < 0 {
				return 0, 0, errors.New("negative offset, must be a non-negative integer")
			}
			resultedOffset = o
		} else {
			return 0, 0, errors.New("invalid offset value")
		}
	}
	return resultedLimit, resultedOffset, nil
}

func ValidateSortOption(sort string) (string, error) {
	if _, ok := validSortOptions[sort]; !ok {
		return "", errors.New("invalid sort option, allowed values are: like_desc, like_asc")
	}
	return sort, nil
}
