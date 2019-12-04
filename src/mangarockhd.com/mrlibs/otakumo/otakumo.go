package otakumo

import (
	"fmt"
	"strconv"
)

// GetUserIDFromOtakumoID Get User Id From Otakumo ID
func GetUserIDFromOtakumoID(userOtakumoID string) string {
	// if len(userOtakumoID) != 18 {
	if len(userOtakumoID) < 18 {
		return ""
	}
	return string(userOtakumoID[8:])
}

// GetMRSourceEntityIdFromOtakumoID get entityID from otakumoID
func GetMRSourceEntityIdFromOtakumoID(otakumoID string) int {
	st := getNumericEntityIDFromOtakumoID(otakumoID)
	eid, err := strconv.Atoi(st)
	if err != nil {
		return 0
	}
	return eid
}

// GetMangaRockSourceOtakumoIDFromEntityID generate Otakumo id for Manga Rock Source Entit ID
func GetMangaRockSourceOtakumoIDFromEntityID(entityType string, id int) string {
	return fmt.Sprintf("mrs-%s-%d", entityType, id)
}

func getNumericEntityIDFromOtakumoID(otakumoID string) string {
	i := len(otakumoID) - 1
	for otakumoID[i] >= '0' && otakumoID[i] <= '9' {
		i--
	}
	if i == len(otakumoID) {
		return ""
	}
	return otakumoID[i+1:]
}
