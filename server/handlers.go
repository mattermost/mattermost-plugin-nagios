package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost-plugin-nagios/go-nagios/nagios"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type commandHandlerFunc func(p *Plugin, channelID string, parameters []string) string

// TODO(DanielSz50): implement get-current-limits command

const logErrorKey = "error"

const setLogsLimitKey = "set-logs-limit"

func getLogsLimit(api plugin.API) (int, error) {
	b, err := api.KVGet(setLogsLimitKey)
	if err != nil {
		return 0, fmt.Errorf("api.KVGet: %w", err)
	}

	var limit int

	if err := json.Unmarshal(b, &limit); err != nil {
		return 0, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return limit, nil
}

func (p *Plugin) setLogsLimit(parameters []string) string {
	if len(parameters) != 1 {
		return "You must supply exactly one parameter (integer value)."
	}

	const settingLogsLimitUnsuccessful = "Setting logs limit unsuccessful."

	i, err := strconv.Atoi(parameters[0])
	if err != nil {
		p.API.LogError("Atoi", logErrorKey, err)
		return settingLogsLimitUnsuccessful
	}

	if i <= 0 {
		return "Invalid argument - logs limit must be a positive integer."
	}

	b, err := json.Marshal(i)
	if err != nil {
		p.API.LogError("Marshal", logErrorKey, err)
		return settingLogsLimitUnsuccessful
	}

	if err := p.API.KVSet(setLogsLimitKey, b); err != nil {
		p.API.LogError("KVSet", logErrorKey, err)
		return settingLogsLimitUnsuccessful
	}

	return "Limit set successfully."
}

func setLogsLimit(p *Plugin, channelID string, parameters []string) string {
	return p.setLogsLimit(parameters)
}

const setLogsStartTimeKey = "set-logs-start-time"

func getLogsStartTime(api plugin.API) (time.Duration, error) {
	b, err := api.KVGet(setLogsStartTimeKey)
	if err != nil {
		return 0, fmt.Errorf("api.KVGet: %w", err)
	}

	var seconds int64

	if err := json.Unmarshal(b, &seconds); err != nil {
		return 0, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return time.Duration(seconds) * time.Second, nil
}

func (p *Plugin) setLogsStartTime(parameters []string) string {
	if len(parameters) != 1 {
		return "You must supply exactly one parameter (number of seconds)."
	}

	const settingLogsStartTimeUnsuccessful = "Setting logs start time unsuccessful."

	i, err := strconv.ParseInt(parameters[0], 10, 64)
	if err != nil {
		p.API.LogError("ParseInt", logErrorKey, err)
		return settingLogsStartTimeUnsuccessful
	}

	if i <= 0 {
		return "Invalid argument - start time must be a positive integer."
	}

	b, err := json.Marshal(i)
	if err != nil {
		p.API.LogError("Marshal", logErrorKey, err)
		return settingLogsStartTimeUnsuccessful
	}

	if err := p.API.KVSet(setLogsStartTimeKey, b); err != nil {
		p.API.LogError("KVSet", logErrorKey, err)
		return settingLogsStartTimeUnsuccessful
	}

	return "Start time set successfully."
}

func setLogsStartTime(p *Plugin, channelID string, parameters []string) string {
	return p.setLogsStartTime(parameters)
}

// formatNagiosTimestamp formats the timestamp from Nagios Core JSON CGIs
// output. These CGIs return the number of milliseconds since the Unix Epoch
// (hence division by 1000). This is contrary to what these CGIs consume, which
// is the _number of seconds_ since the Unix Epoch.
func formatNagiosTimestamp(t int64) string {
	return time.Unix(t/1e3, 0).String()
}

func formatHostName(name, alt string) string {
	if len(name) == 0 {
		return alt
	}

	return name
}

const (
	gettingLogsUnsuccessful = "Getting logs unsuccessful"
	resultTypeTextSuccess   = "Success"
	hostKey                 = "host"
	serviceKey              = "service"
	alertsKey               = "alerts"
	notificationsKey        = "notifications"
	getLogsKey              = "get-logs"
)

func gettingLogsUnsuccessfulMessage(message string) string {
	return fmt.Sprintf("%s: %s", gettingLogsUnsuccessful, message)
}

func unknownParameterMessage(parameter string) string {
	return fmt.Sprintf("Unknown parameter (%s).", parameter)
}

func formatAlertListEntry(e nagios.AlertListEntry) string {
	return fmt.Sprintf("%s [%s] %s: %s | %s | %s | %s | %s",
		emoji(e.State),
		formatNagiosTimestamp(e.Timestamp),
		e.ObjectType,
		formatHostName(e.HostName, e.Name),
		e.Description,
		e.StateType,
		e.State,
		e.PluginOutput)
}

func formatAlerts(alerts nagios.AlertList) string {
	if alerts.Result.TypeText != resultTypeTextSuccess {
		return gettingLogsUnsuccessfulMessage(alerts.Result.TypeText)
	}

	if len(alerts.Data.AlertList) == 0 {
		return "No alerts."
	}

	var b strings.Builder

	for i, v := range alerts.Data.AlertList {
		if i > 0 {
			b.WriteRune('\n')
		}

		b.WriteString(formatAlertListEntry(v))
	}

	return b.String()
}

func formatNotificationListEntry(e nagios.NotificationListEntry) string {
	return fmt.Sprintf("[%s] %s: %s | %s | %s | %s | %s | %s",
		formatNagiosTimestamp(e.Timestamp),
		e.ObjectType,
		formatHostName(e.HostName, e.Name),
		e.Description,
		e.Contact,
		e.NotificationType,
		e.Method,
		e.Message)
}

func formatNotifications(notifications nagios.NotificationList) string {
	if notifications.Result.TypeText != resultTypeTextSuccess {
		return gettingLogsUnsuccessfulMessage(notifications.Result.TypeText)
	}

	if len(notifications.Data.NotificationList) == 0 {
		return "No notifications."
	}

	var b strings.Builder

	for i, v := range notifications.Data.NotificationList {
		if i > 0 {
			b.WriteRune('\n')
		}

		b.WriteString(formatNotificationListEntry(v))
	}

	return b.String()
}

// Cheat sheet:
//
// [command] [action]      [parameters...]
// get-log   alerts        <host>    <URL>
// get-log   alerts        <service> <SVC>
// get-log   notifications <host>    <URL>
// get-log   notifications <service> <SVC>

func getLogsSpecific(parameters []string) (hostName, serviceDescription, message string, ok bool) {
	if len(parameters) == 0 {
		return "", "", "", true
	}

	switch parameters[0] {
	case hostKey:
		if len(parameters) < 2 {
			return "", "", "You must supply host name.", false
		}

		return parameters[1], "", "", true
	case serviceKey:
		if len(parameters) < 2 {
			return "", "", "You must supply service description.", false
		}

		return "", parameters[1], "", true
	default:
		return "", "", unknownParameterMessage(parameters[0]), false
	}
}

func (p *Plugin) getLogs(parameters []string) string {
	if len(parameters) == 0 {
		return "You must supply at least one parameter (alerts|notifications)."
	}

	c, err := getLogsLimit(p.API)
	if err != nil {
		p.API.LogError("getLogsLimit", logErrorKey, err)
		return gettingLogsUnsuccessfulMessage(err.Error())
	}

	hostName, serviceDescription, message, ok := getLogsSpecific(parameters[1:])
	if !ok {
		return message
	}

	d, err := getLogsStartTime(p.API)
	if err != nil {
		p.API.LogError("getLogsStartTime", logErrorKey, err)
		return gettingLogsUnsuccessfulMessage(err.Error())
	}

	now := time.Now()
	then := now.Add(-d)

	switch parameters[0] {
	case alertsKey:
		q := nagios.AlertListRequest{
			GeneralAlertRequest: nagios.GeneralAlertRequest{
				FormatOptions: nagios.FormatOptions{
					Enumerate: true,
				},
				Count:              c,
				HostName:           hostName,
				ServiceDescription: serviceDescription,
				StartTime:          then.Unix(),
				EndTime:            now.Unix(),
			},
		}

		var alerts nagios.AlertList

		if err := p.client.Query(q, &alerts); err != nil {
			p.API.LogError("Query", logErrorKey, err)
			return gettingLogsUnsuccessfulMessage(err.Error())
		}

		return formatAlerts(alerts)
	case notificationsKey:
		q := nagios.NotificationListRequest{
			GeneralNotificationRequest: nagios.GeneralNotificationRequest{
				FormatOptions: nagios.FormatOptions{
					Enumerate: true,
				},
				Count:              c,
				HostName:           hostName,
				ServiceDescription: serviceDescription,
				StartTime:          then.Unix(),
				EndTime:            now.Unix(),
			},
		}

		var notifications nagios.NotificationList

		if err := p.client.Query(q, &notifications); err != nil {
			p.API.LogError("Query", logErrorKey, err)
			return gettingLogsUnsuccessfulMessage(err.Error())
		}

		return formatNotifications(notifications)
	default:
		return unknownParameterMessage(parameters[0])
	}
}

func getLogs(p *Plugin, channelID string, parameters []string) string {
	return p.getLogs(parameters)
}

const setReportFrequencyKey = "set-report-frequency"

func getReportFrequency(api plugin.API) (time.Duration, error) {
	b, err := api.KVGet(setReportFrequencyKey)
	if err != nil {
		return 0, fmt.Errorf("api.KVGet: %w", err)
	}

	var minutes int

	if err := json.Unmarshal(b, &minutes); err != nil {
		return 0, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return time.Duration(minutes) * time.Minute, nil
}

func (p *Plugin) setReportFrequency(parameters []string) string {
	if len(parameters) != 1 {
		return "You must supply exactly one parameter (number of minutes)."
	}

	const settingReportFrequencyUnsuccessful = "Setting report frequency unsuccessful."

	i, err := strconv.Atoi(parameters[0])
	if err != nil {
		p.API.LogError("Atoi", logErrorKey, err)
		return settingReportFrequencyUnsuccessful
	}

	if i <= 0 {
		return "Invalid argument - report frequency must be a positive integer."
	}

	b, err := json.Marshal(i)
	if err != nil {
		p.API.LogError("Marshal", logErrorKey, err)
		return settingReportFrequencyUnsuccessful
	}

	if err := p.API.KVSet(setReportFrequencyKey, b); err != nil {
		p.API.LogError("KVSet", logErrorKey, err)
		return settingReportFrequencyUnsuccessful
	}

	return "Report frequency set successfully."
}

func setReportFrequency(p *Plugin, channelID string, parameters []string) string {
	return p.setReportFrequency(parameters)
}

const (
	reportKey               = "report"
	configurationChangesKey = "configuration-changes"
	subscribeKey            = "subscribe"
	unsubscribeKey          = "unsubscribe"
)

func getReportChannel(api plugin.API) (string, error) {
	b, err := api.KVGet(reportKey)
	if err != nil {
		return "", fmt.Errorf("api.KVGet: %w", err)
	}

	if b == nil {
		return "", nil
	}

	var channel string

	if err := json.Unmarshal(b, &channel); err != nil {
		return "", fmt.Errorf("json.Unmarshal: %w", err)
	}

	return channel, nil
}

func setReportChannel(api plugin.API, channelID string) string {
	const settingReportChannelUnsuccessful = "Setting system monitoring report channel unsuccessful."

	b, err := json.Marshal(channelID)
	if err != nil {
		api.LogError("Marshal", logErrorKey, err)
		return settingReportChannelUnsuccessful
	}

	if err := api.KVSet(reportKey, b); err != nil {
		api.LogError("KVSet", logErrorKey, err)
		return settingReportChannelUnsuccessful
	}

	return "Subscribed to system monitoring report successfully."
}

func getChangesChannel(api plugin.API) (string, error) {
	b, err := api.KVGet(configurationChangesKey)
	if err != nil {
		return "", fmt.Errorf("api.KVGet: %w", err)
	}

	if b == nil {
		return "", nil
	}

	var channel string

	if err := json.Unmarshal(b, &channel); err != nil {
		return "", fmt.Errorf("json.Unmarshal: %w", err)
	}

	return channel, nil
}

func setChangesChannel(api plugin.API, channelID string) string {
	const settingChangesChannelUnsuccessful = "Setting configuration changes channel unsuccessful."

	b, err := json.Marshal(channelID)
	if err != nil {
		api.LogError("Marshal", logErrorKey, err)
		return settingChangesChannelUnsuccessful
	}

	if err := api.KVSet(configurationChangesKey, b); err != nil {
		api.LogError("KVSet", logErrorKey, err)
		return settingChangesChannelUnsuccessful
	}

	return "Subscribed to configuration changes successfully."
}

func (p *Plugin) subscribe(channelID string, parameters []string) string {
	if len(parameters) != 1 {
		return "You must supply exactly one parameter (report|configuration-changes)."
	}

	switch parameters[0] {
	case reportKey:
		return setReportChannel(p.API, channelID)
	case configurationChangesKey:
		return setChangesChannel(p.API, channelID)
	default:
		return unknownParameterMessage(parameters[0])
	}
}

func subscribe(p *Plugin, channelID string, parameters []string) string {
	return p.subscribe(channelID, parameters)
}

func (p *Plugin) unsubscribe(parameters []string) string {
	if len(parameters) != 1 {
		return "You must supply exactly one parameter (report|configuration-changes)."
	}

	const unsubscribingUnsuccessful = "Unsubscribing unsuccessful."

	switch parameters[0] {
	case reportKey:
		if err := p.API.KVDelete(reportKey); err != nil {
			p.API.LogError("KVDelete", logErrorKey, err)
			return unsubscribingUnsuccessful
		}
	case configurationChangesKey:
		if err := p.API.KVDelete(configurationChangesKey); err != nil {
			p.API.LogError("KVDelete", logErrorKey, err)
			return unsubscribingUnsuccessful
		}
	default:
		return unknownParameterMessage(parameters[0])
	}

	return "Unsubscribed successfully."
}

func unsubscribe(p *Plugin, channelID string, parameters []string) string {
	return p.unsubscribe(parameters)
}
