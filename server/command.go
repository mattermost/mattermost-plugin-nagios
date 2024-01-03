package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func getAutoCompleteDesc(m map[string]commandHandlerFunc) string {
	if len(m) == 0 {
		return ""
	}

	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var b strings.Builder

	b.WriteString("Available commands: ")

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(k)
	}

	return b.String()
}

func getHint(before, after rune, keys ...string) string {
	var b strings.Builder

	b.WriteRune(before)

	for i, k := range keys {
		if i > 0 {
			b.WriteRune('|')
		}

		b.WriteString(k)
	}

	b.WriteRune(after)

	return b.String()
}

func getAlertsAutocompleteData() *model.AutocompleteData {
	alerts := model.NewAutocompleteData(
		alertsKey,
		getHint('[', ']', hostKey, serviceKey),
		"Allows you to get alerts.")

	alerts.AddCommand(
		model.NewAutocompleteData(
			hostKey,
			"<host>",
			"Allows you to get alerts from a specific host."))
	alerts.AddCommand(
		model.NewAutocompleteData(
			serviceKey,
			"<service>",
			"Allows you to get alerts from a specific service."))

	return alerts
}

func getNotificationsAutocompleteData() *model.AutocompleteData {
	notifications := model.NewAutocompleteData(
		notificationsKey,
		getHint('[', ']', hostKey, serviceKey),
		"Allows you to get notifications.")

	notifications.AddCommand(
		model.NewAutocompleteData(
			hostKey,
			"<host>",
			"Allows you to get notifications from a specific host."))
	notifications.AddCommand(
		model.NewAutocompleteData(
			serviceKey,
			"<service>",
			"Allows you to get notifications from a specific service."))

	return notifications
}

func getLogsAutocompleteData() *model.AutocompleteData {
	getLogs := model.NewAutocompleteData(
		getLogsKey,
		getHint('<', '>', alertsKey, notificationsKey),
		"Allows you to get alerts or notifications.")

	getLogs.AddCommand(getAlertsAutocompleteData())
	getLogs.AddCommand(getNotificationsAutocompleteData())

	return getLogs
}

func subscribeAutocompleteData() *model.AutocompleteData {
	subscribe := model.NewAutocompleteData(
		subscribeKey,
		getHint('<', '>', reportKey, configurationChangesKey),
		"Allows you to subscribe to system monitoring reports or configuration changes on the current channel.")

	subscribe.AddCommand(model.NewAutocompleteData(
		reportKey,
		"",
		"Allows you to subscribe to system monitoring reports on the current channel."))
	subscribe.AddCommand(model.NewAutocompleteData(
		configurationChangesKey,
		"",
		"Allows you to subscribe to configuration changes on the current channel."))

	return subscribe
}

func unsubscribeAutocompleteData() *model.AutocompleteData {
	unsubscribe := model.NewAutocompleteData(
		unsubscribeKey,
		getHint('<', '>', reportKey, configurationChangesKey),
		"Allows you to unsubscribe from system monitoring reports or configuration changes on the current channel.")

	unsubscribe.AddCommand(model.NewAutocompleteData(
		reportKey,
		"",
		"Allows you to unsubscribe from system monitoring reports on the current channel."))
	unsubscribe.AddCommand(model.NewAutocompleteData(
		configurationChangesKey,
		"",
		"Allows you to unsubscribe from configuration changes on the current channel."))

	return unsubscribe
}

func getAutocompleteData(desc string) *model.AutocompleteData {
	nagios := model.NewAutocompleteData("nagios", "[command]", desc)

	// Auto-complete for get-logs command.
	nagios.AddCommand(getLogsAutocompleteData())

	// Auto-complete for set-logs-limit command.
	nagios.AddCommand(
		model.NewAutocompleteData(
			setLogsLimitKey,
			"[positive integer]",
			"Allows you to limit the number of logs get-logs fetches."))

	// Auto-complete for set-logs-start-time command.
	nagios.AddCommand(
		model.NewAutocompleteData(
			setLogsStartTimeKey,
			"[seconds]",
			"Allows you to specify the age of the oldest log get-logs fetches."))

	// Auto-complete for set-report-frequency command.
	nagios.AddCommand(
		model.NewAutocompleteData(
			setReportFrequencyKey,
			"[minutes]",
			"Allows you to set the frequency of system monitoring reports."))

	// Auto-complete for subscribe command.
	nagios.AddCommand(subscribeAutocompleteData())

	// Auto-complete for unsubscribe command.
	nagios.AddCommand(unsubscribeAutocompleteData())

	return nagios
}

func (p *Plugin) getCommand(iconData string) *model.Command {
	desc := getAutoCompleteDesc(p.commandHandlers)

	return &model.Command{
		Trigger:              "nagios",
		AutoComplete:         true,
		AutoCompleteDesc:     desc,
		AutoCompleteHint:     "[command]",
		DisplayName:          "Nagios",
		Description:          "A Mattermost plugin to interact with Nagios",
		AutocompleteData:     getAutocompleteData(desc),
		AutocompleteIconData: iconData,
	}
}

func parseCommandArgs(args *model.CommandArgs) (
	command, action string,
	parameters []string) {
	fields := strings.SplitN(args.Command, " ", 5)

	if len(fields) > 0 {
		command = fields[0]
	}

	if len(fields) > 1 {
		action = fields[1]
	}

	if len(fields) > 2 {
		parameters = fields[2:]
	}

	return command, action, parameters
}

func (p *Plugin) sendResponse(
	args *model.CommandArgs,
	text string) *model.CommandResponse {
	p.API.SendEphemeralPost(args.UserId, &model.Post{
		UserId:    p.botUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	})

	return &model.CommandResponse{}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (
	*model.CommandResponse,
	*model.AppError) {
	command, action, parameters := parseCommandArgs(args)

	user, err := p.API.GetUser(args.UserId)
	if err != nil {
		return p.sendResponse(args, "User is not registered"), nil
	}

	if !user.IsSystemAdmin() && !p.API.HasPermissionToTeam(args.UserId, args.TeamId, model.PermissionManageTeam) {
		return p.sendResponse(args, "Nagios commands can only be run by System Admins and Team Admins"), nil
	}

	if command != "/nagios" {
		return &model.CommandResponse{}, nil
	}

	var msg string

	if f, ok := p.commandHandlers[action]; ok {
		msg = f(p, args.ChannelId, parameters)
	} else {
		msg = fmt.Sprintf("Unknown action (%s).", action)
	}

	return p.sendResponse(args, msg), nil
}
