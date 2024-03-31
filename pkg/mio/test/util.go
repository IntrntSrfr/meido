package test

import (
	"github.com/intrntsrfr/meido/pkg/utils"
)

func NewTestConfig() *utils.Config {
	conf := utils.NewConfig()
	conf.Set("shards", 1)
	conf.Set("token", "asdf")
	return conf
}
