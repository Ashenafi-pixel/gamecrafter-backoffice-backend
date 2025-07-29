package initiator

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func initConfig(name, path string, log *zap.Logger) {
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Sprintf("unable to start config %v ", err))
	}
}
