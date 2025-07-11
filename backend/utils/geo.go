package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/faceair/clash-speedtest/logger"
)

// GeoLocation represents geographical location information
type GeoLocation struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
	Status      string  `json:"status"`
	Message     string  `json:"message,omitempty"`
}

// CountryFlags maps country codes to flag emojis
var CountryFlags = map[string]string{
	"AD": "🇦🇩", "AE": "🇦🇪", "AF": "🇦🇫", "AG": "🇦🇬", "AI": "🇦🇮", "AL": "🇦🇱", "AM": "🇦🇲",
	"AO": "🇦🇴", "AQ": "🇦🇶", "AR": "🇦🇷", "AS": "🇦🇸", "AT": "🇦🇹", "AU": "🇦🇺", "AW": "🇦🇼",
	"AX": "🇦🇽", "AZ": "🇦🇿", "BA": "🇧🇦", "BB": "🇧🇧", "BD": "🇧🇩", "BE": "🇧🇪", "BF": "🇧🇫",
	"BG": "🇧🇬", "BH": "🇧🇭", "BI": "🇧🇮", "BJ": "🇧🇯", "BL": "🇧🇱", "BM": "🇧🇲", "BN": "🇧🇳",
	"BO": "🇧🇴", "BQ": "🇧🇶", "BR": "🇧🇷", "BS": "🇧🇸", "BT": "🇧🇹", "BV": "🇧🇻", "BW": "🇧🇼",
	"BY": "🇧🇾", "BZ": "🇧🇿", "CA": "🇨🇦", "CC": "🇨🇨", "CD": "🇨🇩", "CF": "🇨🇫", "CG": "🇨🇬",
	"CH": "🇨🇭", "CI": "🇨🇮", "CK": "🇨🇰", "CL": "🇨🇱", "CM": "🇨🇲", "CN": "🇨🇳", "CO": "🇨🇴",
	"CR": "🇨🇷", "CU": "🇨🇺", "CV": "🇨🇻", "CW": "🇨🇼", "CX": "🇨🇽", "CY": "🇨🇾", "CZ": "🇨🇿",
	"DE": "🇩🇪", "DJ": "🇩🇯", "DK": "🇩🇰", "DM": "🇩🇲", "DO": "🇩🇴", "DZ": "🇩🇿", "EC": "🇪🇨",
	"EE": "🇪🇪", "EG": "🇪🇬", "EH": "🇪🇭", "ER": "🇪🇷", "ES": "🇪🇸", "ET": "🇪🇹", "FI": "🇫🇮",
	"FJ": "🇫🇯", "FK": "🇫🇰", "FM": "🇫🇲", "FO": "🇫🇴", "FR": "🇫🇷", "GA": "🇬🇦", "GB": "🇬🇧",
	"GD": "🇬🇩", "GE": "🇬🇪", "GF": "🇬🇫", "GG": "🇬🇬", "GH": "🇬🇭", "GI": "🇬🇮", "GL": "🇬🇱",
	"GM": "🇬🇲", "GN": "🇬🇳", "GP": "🇬🇵", "GQ": "🇬🇶", "GR": "🇬🇷", "GS": "🇬🇸", "GT": "🇬🇹",
	"GU": "🇬🇺", "GW": "🇬🇼", "GY": "🇬🇾", "HK": "🇭🇰", "HM": "🇭🇲", "HN": "🇭🇳", "HR": "🇭🇷",
	"HT": "🇭🇹", "HU": "🇭🇺", "ID": "🇮🇩", "IE": "🇮🇪", "IL": "🇮🇱", "IM": "🇮🇲", "IN": "🇮🇳",
	"IO": "🇮🇴", "IQ": "🇮🇶", "IR": "🇮🇷", "IS": "🇮🇸", "IT": "🇮🇹", "JE": "🇯🇪", "JM": "🇯🇲",
	"JO": "🇯🇴", "JP": "🇯🇵", "KE": "🇰🇪", "KG": "🇰🇬", "KH": "🇰🇭", "KI": "🇰🇮", "KM": "🇰🇲",
	"KN": "🇰🇳", "KP": "🇰🇵", "KR": "🇰🇷", "KW": "🇰🇼", "KY": "🇰🇾", "KZ": "🇰🇿", "LA": "🇱🇦",
	"LB": "🇱🇧", "LC": "🇱🇨", "LI": "🇱🇮", "LK": "🇱🇰", "LR": "🇱🇷", "LS": "🇱🇸", "LT": "🇱🇹",
	"LU": "🇱🇺", "LV": "🇱🇻", "LY": "🇱🇾", "MA": "🇲🇦", "MC": "🇲🇨", "MD": "🇲🇩", "ME": "🇲🇪",
	"MF": "🇲🇫", "MG": "🇲🇬", "MH": "🇲🇭", "MK": "🇲🇰", "ML": "🇲🇱", "MM": "🇲🇲", "MN": "🇲🇳",
	"MO": "🇲🇴", "MP": "🇲🇵", "MQ": "🇲🇶", "MR": "🇲🇷", "MS": "🇲🇸", "MT": "🇲🇹", "MU": "🇲🇺",
	"MV": "🇲🇻", "MW": "🇲🇼", "MX": "🇲🇽", "MY": "🇲🇾", "MZ": "🇲🇿", "NA": "🇳🇦", "NC": "🇳🇨",
	"NE": "🇳🇪", "NF": "🇳🇫", "NG": "🇳🇬", "NI": "🇳🇮", "NL": "🇳🇱", "NO": "🇳🇴", "NP": "🇳🇵",
	"NR": "🇳🇷", "NU": "🇳🇺", "NZ": "🇳🇿", "OM": "🇴🇲", "PA": "🇵🇦", "PE": "🇵🇪", "PF": "🇵🇫",
	"PG": "🇵🇬", "PH": "🇵🇭", "PK": "🇵🇰", "PL": "🇵🇱", "PM": "🇵🇲", "PN": "🇵🇳", "PR": "🇵🇷",
	"PS": "🇵🇸", "PT": "🇵🇹", "PW": "🇵🇼", "PY": "🇵🇾", "QA": "🇶🇦", "RE": "🇷🇪", "RO": "🇷🇴",
	"RS": "🇷🇸", "RU": "🇷🇺", "RW": "🇷🇼", "SA": "🇸🇦", "SB": "🇸🇧", "SC": "🇸🇨", "SD": "🇸🇩",
	"SE": "🇸🇪", "SG": "🇸🇬", "SH": "🇸🇭", "SI": "🇸🇮", "SJ": "🇸🇯", "SK": "🇸🇰", "SL": "🇸🇱",
	"SM": "🇸🇲", "SN": "🇸🇳", "SO": "🇸🇴", "SR": "🇸🇷", "SS": "🇸🇸", "ST": "🇸🇹", "SV": "🇸🇻",
	"SX": "🇸🇽", "SY": "🇸🇾", "SZ": "🇸🇿", "TC": "🇹🇨", "TD": "🇹🇩", "TF": "🇹🇫", "TG": "🇹🇬",
	"TH": "🇹🇭", "TJ": "🇹🇯", "TK": "🇹🇰", "TL": "🇹🇱", "TM": "🇹🇲", "TN": "🇹🇳", "TO": "🇹🇴",
	"TR": "🇹🇷", "TT": "🇹🇹", "TV": "🇹🇻", "TW": "🇹🇼", "TZ": "🇹🇿", "UA": "🇺🇦", "UG": "🇺🇬",
	"UM": "🇺🇲", "US": "🇺🇸", "UY": "🇺🇾", "UZ": "🇺🇿", "VA": "🇻🇦", "VC": "🇻🇨", "VE": "🇻🇪",
	"VG": "🇻🇬", "VI": "🇻🇮", "VN": "🇻🇳", "VU": "🇻🇺", "WF": "🇼🇫", "WS": "🇼🇸", "YE": "🇾🇪",
	"YT": "🇾🇹", "ZA": "🇿🇦", "ZM": "🇿🇲", "ZW": "🇿🇼",
}

// GeoService provides geographical location services
type GeoService struct {
	client  *http.Client
	baseURL string
}

// NewGeoService creates a new geographical location service
func NewGeoService() *GeoService {
	return &GeoService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "http://ip-api.com/json",
	}
}

// GetLocationByIP gets geographical location information for an IP address
func (g *GeoService) GetLocationByIP(ip string) (*GeoLocation, error) {
	// Validate IP address
	if net.ParseIP(ip) == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	url := fmt.Sprintf("%s/%s?fields=status,message,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query", g.baseURL, ip)

	logger.Logger.Debug("Fetching geo location",
		slog.String("ip", ip),
		slog.String("url", url),
	)

	resp, err := g.client.Get(url)
	if err != nil {
		logger.LogError("Failed to fetch geo location", err, slog.String("ip", ip))
		return nil, fmt.Errorf("failed to fetch geo location for %s: %w", ip, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geo API returned status %d for IP %s", resp.StatusCode, ip)
	}

	var location GeoLocation
	if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
		logger.LogError("Failed to decode geo location response", err, slog.String("ip", ip))
		return nil, fmt.Errorf("failed to decode geo location response for %s: %w", ip, err)
	}

	if location.Status != "success" {
		return nil, fmt.Errorf("geo API returned error for IP %s: %s", ip, location.Message)
	}

	logger.Logger.Debug("Geo location fetched successfully",
		slog.String("ip", ip),
		slog.String("country", location.Country),
		slog.String("city", location.City),
		slog.String("isp", location.ISP),
	)

	return &location, nil
}

// GetFlagEmoji returns the flag emoji for a country code
func GetFlagEmoji(countryCode string) string {
	if flag, ok := CountryFlags[strings.ToUpper(countryCode)]; ok {
		return flag
	}
	return "🌍" // Default globe emoji
}

// FormatLocationString formats a location into a readable string
func (g *GeoLocation) FormatLocationString() string {
	flag := GetFlagEmoji(g.CountryCode)
	if g.City != "" && g.Country != "" {
		return fmt.Sprintf("%s %s, %s", flag, g.City, g.Country)
	} else if g.Country != "" {
		return fmt.Sprintf("%s %s", flag, g.Country)
	}
	return fmt.Sprintf("%s Unknown", flag)
}

// FormatProxyName formats a proxy name with location and speed information
func FormatProxyNameWithLocation(originalName string, location *GeoLocation, downloadSpeed, uploadSpeed float64) string {
	if location == nil {
		return originalName
	}

	flag := GetFlagEmoji(location.CountryCode)
	locationStr := location.CountryCode
	if location.City != "" {
		locationStr = fmt.Sprintf("%s %s", location.CountryCode, location.City)
	}

	speedStr := ""
	if downloadSpeed > 0 {
		speedStr = fmt.Sprintf(" | ⬇️ %.2f MB/s", downloadSpeed/(1024*1024))
	}
	if uploadSpeed > 0 {
		if speedStr != "" {
			speedStr += fmt.Sprintf(" ⬆️ %.2f MB/s", uploadSpeed/(1024*1024))
		} else {
			speedStr = fmt.Sprintf(" | ⬆️ %.2f MB/s", uploadSpeed/(1024*1024))
		}
	}

	return fmt.Sprintf("%s %s | %s%s", flag, locationStr, originalName, speedStr)
}

// ExtractIPFromServer extracts IP address from server string (handles domain names)
func ExtractIPFromServer(server string) string {
	// If it's already an IP address, return as is
	if net.ParseIP(server) != nil {
		return server
	}

	// Try to resolve domain name to IP
	ips, err := net.LookupIP(server)
	if err != nil {
		logger.Logger.Debug("Failed to resolve domain to IP",
			slog.String("domain", server),
			slog.String("error", err.Error()),
		)
		return server // Return original if resolution fails
	}

	// Return the first IPv4 address found
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}

	// If no IPv4 found, return first IP
	if len(ips) > 0 {
		return ips[0].String()
	}

	return server
}
