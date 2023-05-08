package static

// Speichert alle unter OSX Verfügabren Pfade ab
const (
	// macOS Dateipfade
	OSX_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	OSX_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	OSX_NO_ROOT_CHANNEL_SOCKET = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	OSX_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	OSX_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	OSX_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	OSX_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	OSX_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// Linux Debian Dateipfade
	DEBIAN_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	DEBIAN_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	DEBIAN_NO_ROOT_CHANNEL_SOCKET = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	DEBIAN_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	DEBIAN_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	DEBIAN_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	DEBIAN_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	DEBIAN_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// FreeBSD Dateipfade
	FREEBSD_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	FREEBSD_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	FREEBSD_NO_ROOT_CHANNEL_SOCKET = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	FREEBSD_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	FREEBSD_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	FREEBSD_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	FREEBSD_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	FREEBSD_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// Windows Dateipfade
	WIN32_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	WIN32_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	WIN32_ROOT_CHANNEL_SOCKET    = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	WIN32_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	WIN32_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	WIN32_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	WIN32_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	WIN32_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// Raspi-Pi Dateipfade
	RASPI_PI_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	RASPI_PI_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	RASPI_PI_NO_ROOT_CHANNEL_SOCKET = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	RASPI_PI_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	RASPI_PI_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	RASPI_PI_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	RASPI_PI_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	RASPI_PI_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"

	// Windows Server Dateipfade
	WIN32_SERVER_BASE_CONFIG_PATH       = "/Users/fluffelbuff/Desktop/rouex.config"
	WIN32_SERVER_NO_ROOT_API_SOCKET     = "/Users/fluffelbuff/Desktop/rouex.socket"
	WIN32_SERVER_NO_ROOT_CHANNEL_SOCKET = "/Users/fluffelbuff/Desktop/rouex_channel.socket"
	WIN32_SERVER_TRUSTED_RELAYS_PATH    = "/Users/fluffelbuff/Desktop/trusted_relays.table"
	WIN32_SERVER_ROUTING_TABLE_PATH     = "/Users/fluffelbuff/Desktop/routing.table"
	WIN32_SERVER_FIREWALL_TABLE_PATH    = "/Users/fluffelbuff/Desktop/firewall.table"
	WIN32_SERVER_EXTERNAL_MODULES       = "/Users/fluffelbuff/Desktop/external_modules"
	WIN32_SERVER_RELAY_PRIVATE_KEY_FILE = "/Users/fluffelbuff/Desktop/relay.privkey.r"
)

// Speichert Namen, Version, etc ab
const (
	ADDRESS_PREFIX string = "rx"
	VERSION        uint64 = 1000000
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
	CHANNEL_PATH     = File(7)
)
