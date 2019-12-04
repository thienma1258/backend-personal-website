package mrmarket


type MangaRockMarket struct {
	Code           string
	Name           string
	Platform       string
	AppIdentifiers map[string]bool
}

var Markets = map[string]*MangaRockMarket{
	"app-store": &MangaRockMarket{
		Code:     "app-store",
		Name:     "iOS",
		Platform: "ios",
		AppIdentifiers: map[string]bool{
			"com.notabasement.mr2": true,
		},
	},
	"google-play": &MangaRockMarket{
		Code:     "google-play",
		Name:     "Google Play Store",
		Platform: "android",
		AppIdentifiers: map[string]bool{
			"com.notabasement.mangarock.android.titan": true,
			"com.mangarockapp.us":                      true,
		},
	},
	"offline": &MangaRockMarket{
		Code:     "offline",
		Name:     "Android Definitive",
		Platform: "android",
		AppIdentifiers: map[string]bool{
			"com.notabasement.mangarock.android.lotus": true,
		},
	},
	"website": &MangaRockMarket{
		Code:     "website",
		Name:     "Website",
		Platform: "website",
		AppIdentifiers: map[string]bool{
			"com.notabasement.mangarock.website": true,
		},
	},
}

var mapAppIdentifiersToMarket map[string]*MangaRockMarket
var mapPlatformToMarket map[string]*MangaRockMarket

func init() {
	mapAppIdentifiersToMarket = make(map[string]*MangaRockMarket)
	mapPlatformToMarket = make(map[string]*MangaRockMarket)
	for _, market := range Markets {
		for identifier := range market.AppIdentifiers {
			mapAppIdentifiersToMarket[identifier] = market
			mapPlatformToMarket[market.Platform] = market
		}
	}
}

func GetMarketFromAppIdentifier(appIdentifier string) *MangaRockMarket {
	if market, ok := mapAppIdentifiersToMarket[appIdentifier]; ok {
		return market
	}
	return nil
}

func GetMarket(platform string, appIdentifier string) *MangaRockMarket {
	if appIdentifier != "" {
		return GetMarketFromAppIdentifier(appIdentifier)
	}
	if market, ok := mapPlatformToMarket[platform]; ok {
		return market
	}
	return nil
}
