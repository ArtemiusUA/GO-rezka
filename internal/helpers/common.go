package helpers

import (
	"github.com/spf13/viper"
	"reflect"
)

func InitConfig() {
	viper.SetEnvPrefix("GOREZKA")
	viper.SetDefault("DEBUG", false)
	viper.SetDefault("DATABASE_URL", "")
	viper.SetDefault("AUTH_TOKEN", "")
	viper.SetDefault("DOMAINS", []string{})
	viper.SetDefault("PORT", 8000)
	viper.SetDefault("HTTPS", false)
	viper.SetDefault("PROXIES", "")
	viper.AutomaticEnv()
}

func ReverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
