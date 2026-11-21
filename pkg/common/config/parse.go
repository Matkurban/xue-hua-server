package config

import (
	"os"
	"path/filepath"

	"server/v3/pkg/common/errs"
	"server/v3/pkg/common/field"
	"github.com/openimsdk/protocol/constant"
	"server/v3/pkg/msgprocessor"
)

// return absolude path join ../config/, this is k8s container config path.

// getProjectRoot returns the absolute path of the project root directory.
func GetProjectRoot() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", errs.Wrap(err)
	}
	projectRoot, err := field.OutDir(filepath.Join(filepath.Dir(executablePath), "../../../../.."))
	if err != nil {
		return "", errs.Wrap(err)
	}
	return projectRoot, nil
}

func GetOptionsByNotification(cfg NotificationConfig, sendMessage *bool) msgprocessor.Options {
	opts := msgprocessor.NewOptions()

	if sendMessage != nil {
		cfg.IsSendMsg = *sendMessage
	}
	if cfg.IsSendMsg {
		opts = msgprocessor.WithOptions(opts, msgprocessor.WithUnreadCount(true))
	}
	if cfg.OfflinePush.Enable {
		opts = msgprocessor.WithOptions(opts, msgprocessor.WithOfflinePush(true))
	}
	switch cfg.ReliabilityLevel {
	case constant.UnreliableNotification:
	case constant.ReliableNotificationNoMsg:
		opts = msgprocessor.WithOptions(opts, msgprocessor.WithHistory(true), msgprocessor.WithPersistent())
	}
	opts = msgprocessor.WithOptions(opts, msgprocessor.WithSendMsg(cfg.IsSendMsg))

	return opts
}
