package mrutils

import (
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"mangarockhd.com/mrlibs/mrconstants"
)

const primeRK = 16777619

var reverseCountryNameToCode map[string]string

func init() {
	reverseCountryNameToCode = make(map[string]string)
	for code, countryName := range mrconstants.COUNTRY_NAME {
		reverseCountryNameToCode[strings.ToLower(countryName)] = code
	}
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// hashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func hashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

//FindAllSubStrings find all positions of a strToSearch in source
func FindAllSubStrings(source *string, strToSearch string) []uint32 {
	// n := 0
	// special cases
	str := *source
	lenSource := len(str)
	lenNeedle := len(strToSearch)
	switch {
	case lenNeedle == 0:
		return []uint32{}
	case lenNeedle == 1:
		// special case worth making fast
		results := []uint32{}
		c := strToSearch[0]
		for i := 0; i < lenSource; i++ {
			if str[i] == c {
				results = append(results, uint32(i))
			}
		}
		return results
	case lenNeedle > lenSource:
		return []uint32{}
	case lenNeedle == lenSource:
		if strToSearch == str {
			return []uint32{0}
		}
		return []uint32{}
	}
	results := []uint32{}
	// Rabin-Karp search
	hashsep, pow := hashStr(strToSearch)
	h := uint32(0)
	for i := 0; i < lenNeedle; i++ {
		h = h*primeRK + uint32((str)[i])
	}
	lastmatch := 0
	if h == hashsep && (str)[:lenNeedle] == strToSearch {
		results = append(results, 0)
		lastmatch = lenNeedle
	}
	for i := lenNeedle; i < lenSource; {
		h *= primeRK
		h += uint32(str[i])
		h -= pow * uint32(str[i-lenNeedle])
		i++
		if h == hashsep && i-lenNeedle < lenSource &&
			lastmatch <= i-lenNeedle && str[i-lenNeedle:i] == strToSearch {
			results = append(results, uint32(i-lenNeedle))
			lastmatch = i
		}
	}
	return results
}

// FindAllLinesHavingString Return the start position of lines which contain a string
func FindAllLinesHavingString(source *string, strToSearch string, maxLine int) []uint32 {
	// n := 0
	// special cases
	str := *source
	lenSource := len(str)
	lenNeedle := len(strToSearch)

	switch {
	case lenNeedle == 0:
		return []uint32{}
	case lenNeedle == 1:
		// special case worth making fast
		results := []uint32{}
		c := strToSearch[0]
		findingEndOfLine := false
		for i := 0; i < lenSource; i++ {
			if findingEndOfLine && str[i] == '\n' {
				findingEndOfLine = false
			} else if !findingEndOfLine && str[i] == c {
				results = append(results, uint32(i))
				if len(results) >= maxLine {
					return results
				}
				findingEndOfLine = true
			}
		}
		return results
	case lenNeedle > lenSource:
		return []uint32{}
	case lenNeedle == lenSource:
		if strToSearch == str {
			return []uint32{0}
		}
		return []uint32{}
	}
	results := []uint32{}
	// Rabin-Karp search
	hashsep, pow := hashStr(strToSearch)
	h := uint32(0)
	findingEndOfLine := false
	for i := 0; i < lenNeedle; i++ {
		h = h*primeRK + uint32((str)[i])
	}
	lastmatch := 0
	if h == hashsep && (str)[:lenNeedle] == strToSearch {
		results = append(results, 0)
		findingEndOfLine = true
		lastmatch = lenNeedle
	}
	for i := lenNeedle; i < lenSource; {
		h *= primeRK
		h += uint32(str[i])
		h -= pow * uint32(str[i-lenNeedle])
		i++
		if !findingEndOfLine {
			if h == hashsep && lastmatch <= i-lenNeedle && str[i-lenNeedle:i] == strToSearch {
				results = append(results, uint32(i-lenNeedle))
				if len(results) >= maxLine {
					return results
				}
				findingEndOfLine = true
				lastmatch = i
			}
		} else if i == lenSource || str[i] == '\n' {
			findingEndOfLine = false
		}
	}
	return results
}

// Log do logging with code postion
func Log(format string, v ...interface{}) {
	stackTrace := string(debug.Stack())
	start := strings.Index(stackTrace, "mrutils.Log(0x")
	count := 0
	i := start + 10
	length := len(stackTrace)
	for ; i < length && count < 5; i++ {
		if stackTrace[i] == '\n' {
			count++
			if count == 3 {
				start = strings.Index(stackTrace[i:], "/src/") + i + 5
			}
		}
	}
	end := strings.Index(stackTrace[start:], " ") + start + 1
	pos := ""
	if start > 0 && end > 0 && start < length && end < length {
		pos = stackTrace[start:end]
	}
	log.Printf(pos+format, v...)
}
