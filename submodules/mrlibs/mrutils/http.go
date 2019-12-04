package mrutils

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mangarockhd.com/mrlibs/mrconstants"
)

var urlReplacer = strings.NewReplacer("%25", "%", "%2F", "/", "%3F", "?", "%3D", "=", "%26", "&", "%3A", ":", "+", "%20")

var validCountries = make(map[string]bool)

func init() {
	for code, country := range mrconstants.COUNTRY_NAME {
		validCountries[code] = true
		validCountries[country] = true
	}
}

//HandleCORSHeader Handle CORS header for request, return true if it can continue
func HandleCORSHeader(writer http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")

	if strings.Index(origin, "nabstudio.com") > 0 ||
		strings.Index(origin, "mangarock.com") > 0 ||
		strings.Index(origin, "localhost") > 0 ||
		strings.Index(origin, "mangarock.dev") > 0 {
		writer.Header().Set("Access-Control-Allow-Origin", origin)
		writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
		writer.Header().Set("Access-Control-Max-Age", "86400")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, If-None-Match, Authorization, Accept, QToken, Accept-Encoding")
	}
	if strings.ToUpper(r.Method) == "OPTIONS" {
		writer.WriteHeader(204) // send the headers with a 204 response code.
		return false
	}
	return true
}

//HandleCORSHeaderHandler Handle CORS header for request, return true if it can continue
func HandleCORSHeaderHandler(writer http.ResponseWriter, r *http.Request) {
	HandleCORSHeader(writer, r)
}

// CheckHTTPGetIntParam check http get param
func CheckHTTPGetIntParam(param string, r *http.Request) (int, bool) {
	_param := GetHTTPGetParam(param, r)
	if len(_param) == 0 {
		return 0, false
	}

	value, err := strconv.Atoi(_param)
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

// GetRemoteCountry get country name base on request
func GetRemoteCountry(r *http.Request) string {
	country := GetHTTPGetParam("country", r)
	if len(country) == 0 {
		country = GetRemoteIPCountry(r)
	}
	if validCountries[country] == true {
		return country
	}
	return "Unknown"
}

// GetRemoteCountryCode get country name base on request
func GetRemoteCountryCode(r *http.Request) string {
	country := GetHTTPGetParam("country", r)
	if len(country) == 0 {
		return GetRemoteIPCountryCode(r)
	}
	return GetCountryCodeByCountryName(country)
}

// GetHTTPGetParam get a "GET" param value
func GetHTTPGetParam(param string, r *http.Request) string {
	_get := r.URL.Query()
	if data, ok := _get[param]; ok && len(data) > 0 {
		// log.Printf("GetHTTPGetParam %v", data)
		return data[0]
	}
	return ""
}

// GetRemoteIPCountry trying to get country name base on Cloudflare header
func GetRemoteIPCountry(r *http.Request) string {
	countryCode := ""
	if len(r.Header.Get("Cf-Ipcountry")) > 0 {
		countryCode = r.Header.Get("Cf-Ipcountry")
	} else if len(r.Header.Get("cf-ipcountry")) > 0 {
		countryCode = r.Header.Get("cf-ipcountry")
	}
	countryCode = strings.ToUpper(countryCode)
	if countryName, ok := mrconstants.COUNTRY_NAME[countryCode]; ok {
		return countryName
	}
	return "Unknown"
}

// GetRemoteIPCountryCode trying to get country name base on Cloudflare header
func GetRemoteIPCountryCode(r *http.Request) string {
	countryCode := ""
	if len(r.Header.Get("Cf-Ipcountry")) > 0 {
		countryCode = r.Header.Get("Cf-Ipcountry")
	} else if len(r.Header.Get("cf-ipcountry")) > 0 {
		countryCode = r.Header.Get("cf-ipcountry")
	}
	countryCode = strings.ToUpper(countryCode)
	return countryCode
}

// GetCountryCodeByCountryName get country code by country name
func GetCountryCodeByCountryName(countryName string) string {
	if len(countryName) == 0 {
		return ""
	}

	name := strings.ToLower(countryName)
	if code, ok := reverseCountryNameToCode[name]; ok {
		return code
	}
	return countryName
}

func GetRemoteIp(r *http.Request) string {
	ip := r.RemoteAddr
	if len(r.Header.Get("CF-Connecting-IP")) > 0 {
		ip = r.Header.Get("CF-Connecting-IP")
	} else if len(r.Header.Get("X-Forwarded-For")) > 0 {
		ip = r.Header.Get("X-Forwarded-For")
	} else if len(r.Header.Get("x-forwarded-for")) > 0 {
		ip = r.Header.Get("x-forwarded-for")
	} else if len(r.Header.Get("X-FORWARDED-FOR")) > 0 {
		ip = r.Header.Get("X-FORWARDED-FOR")
	}
	return ip
}

func SafeUrl(urlStr string) string {
	// urlStr = strings.Replace(urlStr, "https://f01.mrcdn.info/", "http://reader.hoang.nabstudio.com:8092/", 1)

	if len(urlStr) < 5 {
		return urlStr
	}

	// immediate return if url not start with 'http' | '//'
	if urlStr[:4] != "http" && urlStr[:2] != "//" {
		return ""
	}

	if urlStr[0:2] == "//" {
		urlStr = "https:" + urlStr
	}

	x := -1
	x1 := strings.Index(urlStr, "://")
	if x1 > 0 {
		x = strings.Index(urlStr[(x1+3):], "/")
	}

	if x == -1 || x1 == -1 {
		var err error
		urlStr, err = url.QueryUnescape(urlStr)
		if err != nil {
			log.Printf("Cannot parse url %s", urlStr)
			return urlStr
		}
		x1 = strings.Index(urlStr, "://")
		if x1 > 0 {
			x = strings.Index(urlStr[(x1+3):], "/")
		}
		if x == -1 {
			log.Printf("Cannot parse url %s", urlStr)
			return urlStr
		}
	}
	lastPart := urlReplacer.Replace(url.QueryEscape(urlStr[x:]))
	return strings.Trim(urlStr[:x]+lastPart, " ")
}

// ValidateQTokenHeaderFromRequest validate QToken header
func ValidateQTokenHeaderFromRequest(r *http.Request, appChecksum []string) bool {
	clientToken := r.Header.Get("QToken")
	if clientToken == "" {
		return false
	}
	url := "https://" + r.Host + r.URL.Path + "?" + r.URL.RawQuery
	isValid := false
	totalChecksum := len(appChecksum)
	if totalChecksum == 0 {
		hash := Md5(Md5(Md5(url)+"mr") + "nabvn")
		isValid = (hash == clientToken)
		if !isValid {
			log.Printf("Failed url=%s clientToken=%s hash=%s r=%+v", url, clientToken, hash, r)
		}
	} else { //new qToken algorithm
		firstByte := clientToken[0]
		for i := 0; i < totalChecksum; i++ {
			checksum := appChecksum[i]
			if checksum[0] != firstByte {
				continue
			}
			hash := checksum[:1] + XXHASH64(url+":"+checksum)
			if hash == clientToken {
				return true
			}
		}
	}
	return isValid
}
