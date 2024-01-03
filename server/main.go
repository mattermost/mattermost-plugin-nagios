package main

import (
	"github.com/mattermost/mattermost/server/public/plugin"
)

func main() {
	plugin.ClientMain(&Plugin{
		commandHandlers: map[string]commandHandlerFunc{
			setLogsLimitKey:       setLogsLimit,
			setLogsStartTimeKey:   setLogsStartTime,
			getLogsKey:            getLogs,
			setReportFrequencyKey: setReportFrequency,
			subscribeKey:          subscribe,
			unsubscribeKey:        unsubscribe,
		},
	})
}
