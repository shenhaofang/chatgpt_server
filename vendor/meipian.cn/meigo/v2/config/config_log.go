package config

type LogConfig struct {
	Path   string
	Format string
	Daily  bool
	Level  string
}

func LoadLogConfig() (logger *LogConfig) {
	return &LogConfig{
		Path: GetStr("log.path"),
		Format: GetDft("log.format", "text"),
		Daily: GetBool("log.daily", false),
		Level: GetDft("log.level", "info"),
	}
}
