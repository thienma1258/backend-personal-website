package utils

import (
	"net/http"
	"os"
	"strings"
)

var COUNTRY_NAME = map[string]string{
	"AF": "Afghanistan",
	"AX": "Aland Islands",
	"AL": "Albania",
	"DZ": "Algeria",
	"AS": "American Samoa",
	"AD": "Andorra",
	"AO": "Angola",
	"AI": "Anguilla",
	"AQ": "Antarctica",
	"AG": "Antigua and Barbuda",
	"AR": "Argentina",
	"AM": "Armenia",
	"AW": "Aruba",
	"AU": "Australia",
	"AT": "Austria",
	"AZ": "Azerbaijan",
	"BS": "Bahamas",
	"BH": "Bahrain",
	"BD": "Bangladesh",
	"BB": "Barbados",
	"BY": "Belarus",
	"BE": "Belgium",
	"BZ": "Belize",
	"BJ": "Benin",
	"BM": "Bermuda",
	"BT": "Bhutan",
	"BO": "Bolivia",
	"BA": "Bosnia and Herzegovina",
	"BW": "Botswana",
	"BV": "Bouvet Island",
	"BR": "Brazil",
	"IO": "British Indian Ocean Territory",
	"VG": "British Virgin Islands",
	"BN": "Brunei Darussalam",
	"BG": "Bulgaria",
	"BF": "Burkina Faso",
	"MM": "Myanmar",
	"BI": "Burundi",
	"KH": "Cambodia",
	"CM": "Cameroon",
	"CA": "Canada",
	"CV": "Cape Verde",
	"KY": "Cayman Islands",
	"CF": "Central African Republic",
	"TD": "Chad",
	"CL": "Chile",
	"CN": "China",
	"CX": "Christmas Island",
	"CC": "Cocos (Keeling) Islands",
	"CO": "Colombia",
	"CI": "Cote D'Ivoire",
	"KM": "Comoros",
	"CK": "Cook Islands",
	"CR": "Costa Rica",
	"HR": "Croatia",
	"CU": "Cuba",
	"CY": "Cyprus",
	"CZ": "Czech Republic",
	"CD": "Democratic Republic of the Congo",
	"DK": "Denmark",
	"DJ": "Djibouti",
	"DM": "Dominica",
	"DO": "Dominican Republic",
	"EC": "Ecuador",
	"EG": "Egypt",
	"SV": "El Salvador",
	"GQ": "Equatorial Guinea",
	"ER": "Eritrea",
	"EE": "Estonia",
	"ET": "Ethiopia",
	"FK": "Falkland Islands (Malvinas)",
	"FO": "Faroe Islands",
	"FJ": "Fiji",
	"FI": "Finland",
	"FR": "France",
	"PF": "French Polynesia",
	"GA": "Gabon",
	"GM": "Gambia",
	"GE": "Georgia",
	"DE": "Germany",
	"GH": "Ghana",
	"GI": "Gibraltar",
	"GR": "Greece",
	"GL": "Greenland",
	"GD": "Grenada",
	"GU": "Guam",
	"GT": "Guatemala",
	"GN": "Guinea",
	"GW": "Guinea-Bissau",
	"GY": "Guyana",
	"HT": "Haiti",
	"HN": "Honduras",
	"HK": "Hong Kong",
	"HU": "Hungary",
	"IS": "Iceland",
	"IN": "India",
	"ID": "Indonesia",
	"IQ": "Iraq",
	"IE": "Ireland",
	"IM": "Isle of Man",
	"IL": "Israel",
	"IT": "Italy",
	"JM": "Jamaica",
	"JP": "Japan",
	"JE": "Jersey",
	"JO": "Jordan",
	"KZ": "Kazakhstan",
	"KE": "Kenya",
	"KI": "Kiribati",
	"KW": "Kuwait",
	"KG": "Kyrgyzstan",
	"LA": "Laos",
	"LV": "Latvia",
	"LB": "Lebanon",
	"LS": "Lesotho",
	"LR": "Liberia",
	"LY": "Libya",
	"LI": "Liechtenstein",
	"LT": "Lithuania",
	"LU": "Luxembourg",
	"MO": "Macao",
	"MK": "Macedonia",
	"MG": "Madagascar",
	"MW": "Malawi",
	"MY": "Malaysia",
	"MV": "Maldives",
	"ML": "Mali",
	"MT": "Malta",
	"MH": "Marshall Islands",
	"MR": "Mauritania",
	"MU": "Mauritius",
	"YT": "Mayotte",
	"MX": "Mexico",
	"MD": "Moldova",
	"MC": "Monaco",
	"MN": "Mongolia",
	"ME": "Montenegro",
	"MS": "Montserrat",
	"MA": "Morocco",
	"MZ": "Mozambique",
	"NA": "Namibia",
	"NR": "Nauru",
	"NP": "Nepal",
	"NL": "Netherlands",
	"AN": "Netherlands Antilles",
	"NC": "New Caledonia",
	"NZ": "New Zealand",
	"NI": "Nicaragua",
	"NE": "Niger",
	"NG": "Nigeria",
	"NU": "Niue",
	"KP": "North Korea",
	"MP": "Northern Mariana Islands",
	"NO": "Norway",
	"OM": "Oman",
	"PK": "Pakistan",
	"PW": "Palau",
	"PA": "Panama",
	"PG": "Papua New Guinea",
	"PY": "Paraguay",
	"PE": "Peru",
	"PH": "Philippines",
	"PN": "Pitcairn Islands",
	"PL": "Poland",
	"PT": "Portugal",
	"PR": "Puerto Rico",
	"QA": "Qatar",
	"CG": "Republic of the Congo",
	"RO": "Romania",
	"RU": "Russian Federation",
	"RW": "Rwanda",
	"BL": "Saint Barthelemy",
	"SH": "Saint Helena",
	"KN": "Saint Kitts and Nevis",
	"LC": "Saint Lucia",
	"MF": "Saint Martin",
	"PM": "Saint Pierre and Miquelon",
	"VC": "Saint Vincent And Grenadines",
	"WS": "Samoa",
	"SM": "San Marino",
	"ST": "Sao Tome and Principe",
	"SA": "Saudi Arabia",
	"SN": "Senegal",
	"RS": "Serbia",
	"SC": "Seychelles",
	"SL": "Sierra Leone",
	"SG": "Singapore",
	"SK": "Slovakia",
	"SI": "Slovenia",
	"SB": "Solomon Islands",
	"SO": "Somalia",
	"ZA": "South Africa",
	"KR": "Republic of Korea",
	"ES": "Spain",
	"LK": "Sri Lanka",
	"SD": "Sudan",
	"SR": "Suriname",
	"SJ": "Svalbard And Jan Mayen",
	"SZ": "Swaziland",
	"SE": "Sweden",
	"CH": "Switzerland",
	"TW": "Taiwan",
	"TJ": "Tajikistan",
	"TZ": "Tanzania",
	"TH": "Thailand",
	"TL": "Timor-Leste",
	"TG": "Togo",
	"TK": "Tokelau",
	"TO": "Tonga",
	"TT": "Trinidad and Tobago",
	"TN": "Tunisia",
	"TR": "Turkey",
	"TM": "Turkmenistan",
	"TC": "Turks and Caicos Islands",
	"TV": "Tuvalu",
	"UG": "Uganda",
	"UA": "Ukraine",
	"AE": "United Arab Emirates",
	"GB": "United Kingdom",
	"US": "United States",
	"UY": "Uruguay",
	"VI": "US Virgin Islands",
	"UZ": "Uzbekistan",
	"VU": "Vanuatu",
	"VE": "Venezuela",
	"VN": "Vietnam",
	"WF": "Wallis and Futuna",
	"EH": "Western Sahara",
	"YE": "Yemen",
	"ZM": "Zambia",
	"ZW": "Zimbabwe",
	"GF": "French Guiana",
	"TF": "French Southern Territories",
	"GP": "Guadeloupe",
	"GG": "Guernsey",
	"VA": "Holy See (Vatican City State)",
	"IR": "Iran, Islamic Republic Of",
	"MQ": "Martinique",
	"FM": "Micronesia, Federated States Of",
	"NF": "Norfolk Island",
	"PS": "Palestinian Territory, Occupied",
	"RE": "Reunion",
	"GS": "South Georgia And Sandwich Isl.",
	"SY": "Syrian Arab Republic",
}

var reverseCountryNameToCode map[string]string

// Verbose verbose
var verbose = os.Getenv("VERBOSE") == "1"

func init() {
	reverseCountryNameToCode = make(map[string]string)
	for code, countryName := range COUNTRY_NAME {
		reverseCountryNameToCode[strings.ToLower(countryName)] = code
	}
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

// GetRemoteIPCountry trying to get country name base on Cloudflare header
func GetRemoteIPCountry(r *http.Request) string {
	countryCode := ""
	if len(r.Header.Get("Cf-Ipcountry")) > 0 {
		countryCode = r.Header.Get("Cf-Ipcountry")
	} else if len(r.Header.Get("cf-ipcountry")) > 0 {
		countryCode = r.Header.Get("cf-ipcountry")
	}
	countryCode = strings.ToUpper(countryCode)
	if countryName, ok := COUNTRY_NAME[countryCode]; ok {
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
