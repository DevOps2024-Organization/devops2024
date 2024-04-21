package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/bshuster-repo/logrus-logstash-hook"
	"net"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	conn, err := net.Dial("udp", "localhost:5228")
	if err != nil {
			Log.Fatal(err)
	}
	hook := logrustash.New(conn, logrustash.DefaultFormatter(logrus.Fields{"type": "myappName"}))

	Log.Hooks.Add(hook)
}
