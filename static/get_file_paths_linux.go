//go:build linux && !arm

package static

// Gibt den PATH für eine bestimmte Datei aus
func GetFilePathFor(fp File) string {
	switch fp {
	case BASE_CONFIG:
		return DEBIAN_BASE_CONFIG_PATH
	case API_SOCKET:
		return DEBIAN_FIREWALL_TABLE_PATH
	case TRUSTED_RELAYS:
		return DEBIAN_TRUSTED_RELAYS_PATH
	case ROUTING_TABLE:
		return DEBIAN_ROUTING_TABLE_PATH
	case FIREWALL_TABLE:
		return DEBIAN_FIREWALL_TABLE_PATH
	case EXTERNAL_MODULES:
		return DEBIAN_EXTERNAL_MODULES
	case PRIVATE_KEY_FILE:
		return DEBIAN_RELAY_PRIVATE_KEY_FILE
	default:
		return ""
	}
}
