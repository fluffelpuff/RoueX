package static

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
