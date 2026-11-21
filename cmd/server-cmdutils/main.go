package main

import (
	"server/v3/pkg/common/program"
	"server/v3/pkg/common/cmd"
)

func main() {
	msgUtilsCmd := cmd.NewMsgUtilsCmd("serverCmdUtils", "server cmd utils", nil)
	getCmd := cmd.NewGetCmd()
	fixCmd := cmd.NewFixCmd()
	clearCmd := cmd.NewClearCmd()
	seqCmd := cmd.NewSeqCmd()
	msgCmd := cmd.NewMsgCmd()
	getCmd.AddCommand(seqCmd.GetSeqCmd(), msgCmd.GetMsgCmd())
	getCmd.AddSuperGroupIDFlag()
	getCmd.AddUserIDFlag()
	getCmd.AddConfigDirFlag()
	getCmd.AddIndexFlag()
	getCmd.AddBeginSeqFlag()
	getCmd.AddLimitFlag()
	// server get seq --userID=xxx
	// server get seq --superGroupID=xxx
	// server get msg --userID=xxx --beginSeq=100 --limit=10
	// server get msg --superGroupID=xxx --beginSeq=100 --limit=10

	fixCmd.AddCommand(seqCmd.FixSeqCmd())
	fixCmd.AddSuperGroupIDFlag()
	fixCmd.AddUserIDFlag()
	fixCmd.AddConfigDirFlag()
	fixCmd.AddIndexFlag()
	fixCmd.AddFixAllFlag()
	// server fix seq --userID=xxx
	// server fix seq --superGroupID=xxx
	// server fix seq --fixAll

	clearCmd.AddCommand(msgCmd.ClearMsgCmd())
	clearCmd.AddSuperGroupIDFlag()
	clearCmd.AddUserIDFlag()
	clearCmd.AddConfigDirFlag()
	clearCmd.AddIndexFlag()
	clearCmd.AddClearAllFlag()
	clearCmd.AddBeginSeqFlag()
	clearCmd.AddLimitFlag()
	// server clear msg --userID=xxx --beginSeq=100 --limit=10
	// server clear msg --superGroupID=xxx --beginSeq=100 --limit=10
	// server clear msg --clearAll
	msgUtilsCmd.AddCommand(&getCmd.Command, &fixCmd.Command, &clearCmd.Command)
	if err := msgUtilsCmd.Execute(); err != nil {
		program.ExitWithError(err)
	}
}
