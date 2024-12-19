package db_config

type DBConfig struct {
	Driver              string `mapstructure:"driver"`
	Host                string `mapstructure:"host"`
	Port                string `mapstructure:"port"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	Database            string `mapstructure:"database"`
	SchemaPath          string `mapstructure:"schema_path"`
	IsDefaultConnection bool
}
