package data

import (
	"github.com/spf13/cobra"
)

func getCmdDataPath(cmd *cobra.Command) string {
	dataPath, _ := cmd.Flags().GetString("data")

	// TODO: handle errors
	return dataPath
}

func LoadCmdData(cmd *cobra.Command) Data {
	return LoadData(getCmdDataPath((cmd)))
}

func SaveCmdData(cmd *cobra.Command, d Data) {
	// TODO: handle errors
	SaveData(getCmdDataPath(cmd), d)
}
