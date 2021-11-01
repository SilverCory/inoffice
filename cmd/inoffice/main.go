package main

import (
	"office"
	"office/inoffice"
)

func main() {
	var envCfg = office.GetEnv()

	var store inoffice.Store = inoffice.NewStore(envCfg)

	inoffice.StartServer(store, envCfg)
}
