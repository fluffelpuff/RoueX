package static

// Speichert alle unter OSX Verfügabren Pfade ab
const (
	// OSX Dateipfade
	OSX_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	OSX_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	OSX_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	OSX_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	OSX_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	OSX_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	OSX_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"
)

// Definiert den Datentypen für Dateieen
type File int

// Definiert alle Verfügabren Dateien
const (
	BASE_CONFIG      = File(0)
	API_SOCKET       = File(1)
	TRUSTED_RELAYS   = File(2)
	ROUTING_TABLE    = File(3)
	FIREWALL_TABLE   = File(4)
	EXTERNAL_MODULES = File(5)
	PRIVATE_KEY_FILE = File(6)
)

// Gibt den PATH für eine bestimmte Datei aus
func GetFilePathFor(fp File) string {
	switch fp {
	case BASE_CONFIG:
		return OSX_BASE_CONFIG_PATH
	case API_SOCKET:
		return OSX_NO_ROOT_API_SOCKET
	case TRUSTED_RELAYS:
		return OSX_TRUSTED_RELAYS_PATH
	case ROUTING_TABLE:
		return OSX_ROUTING_TABLE_PATH
	case FIREWALL_TABLE:
		return OSX_FIREWALL_TABLE_PATH
	case EXTERNAL_MODULES:
		return OSX_EXTERNAL_MODULES
	case PRIVATE_KEY_FILE:
		return OSX_RELAY_PRIVATE_KEY_FILE
	default:
		return ""
	}
}
