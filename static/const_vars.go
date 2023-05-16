package static

// Speichert alle unter OSX Verfügabren Pfade ab
const (
	// macOS Dateipfade
	OSX_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	OSX_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	OSX_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	OSX_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	OSX_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	OSX_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	OSX_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// Linux Dateipfade
	DEBIAN_BASE_CONFIG_PATH       = "/home/fluffelbuff/Schreibtisch/rouex.config"
	DEBIAN_NO_ROOT_API_SOCKET     = "/home/fluffelbuff/Schreibtisch/rouex.socket"
	DEBIAN_TRUSTED_RELAYS_PATH    = "/home/fluffelbuff/Schreibtisch/trusted_relays.table"
	DEBIAN_ROUTING_TABLE_PATH     = "/home/fluffelbuff/Schreibtisch/routing.table"
	DEBIAN_FIREWALL_TABLE_PATH    = "/home/fluffelbuff/Schreibtisch/firewall.table"
	DEBIAN_EXTERNAL_MODULES       = "/home/fluffelbuff/Schreibtisch/external_modules/"
	DEBIAN_RELAY_PRIVATE_KEY_FILE = "/home/fluffelbuff/Schreibtisch/relay.privkey.r"

	// Windows Dateipfade
	WIN32_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	WIN32_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	WIN32_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	WIN32_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	WIN32_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	WIN32_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	WIN32_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"
)

// Speichert Namen, Version, etc ab
const (
	// Gibt den Adressprefix an
	ADDRESS_PREFIX string = "rx"

	// Gibt die Version an
	VERSION uint64 = 1000000
)

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
