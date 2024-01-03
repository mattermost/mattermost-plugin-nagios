package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mattermost/mattermost-plugin-nagios/go-nagios/nagios"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/command"
)

// Plugin implements the interface expected by the Mattermost server to
// communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult
	// getConfiguration and setConfiguration for usage.
	configuration *configuration

	client *nagios.Client

	botUserID string

	commandHandlers map[string]commandHandlerFunc

	monitoringReportJob *cluster.Job
}

func (p *Plugin) storeInitialKV() error {
	initials := map[string]int{
		setLogsLimitKey:       p.getConfiguration().InitialLogsLimit,
		setLogsStartTimeKey:   p.getConfiguration().InitialLogsStartTime,
		setReportFrequencyKey: p.getConfiguration().InitialReportFrequency,
	}

	for key, val := range initials {
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Errorf("json.Marshal: %w", err)
		}

		if err := p.API.KVSet(key, b); err != nil {
			return fmt.Errorf("p.API.KVSet: %w", err)
		}
	}

	return nil
}

func (p *Plugin) getProfileImage() ([]byte, error) {
	path, err := p.API.GetBundlePath()
	if err != nil {
		return nil, fmt.Errorf("p.API.GetBundlePath: %w", err)
	}

	// NOTICE: We don't use any of the Nagios logos to avoid legal issues.
	// Instead, we use an image resembling a part of the Nagios Core logo.
	file, err := os.ReadFile(filepath.Join(path, "assets", "orbit-467260.png"))
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile: %w", err)
	}

	return file, nil
}

func (p *Plugin) OnActivate() error {
	client := pluginapi.NewClient(p.API, p.Driver)

	botUserID, err := client.Bot.EnsureBot(&model.Bot{
		Username:    "nagios",
		DisplayName: "Nagios",
		Description: "Created by the Nagios Plugin.",
	})
	if err != nil {
		return fmt.Errorf("p.Helpers.EnsureBot: %w", err)
	}

	p.botUserID = botUserID

	img, err := p.getProfileImage()
	if err != nil {
		return fmt.Errorf("p.getProfileImage: %w", err)
	}

	if err := p.API.SetProfileImage(botUserID, img); err != nil {
		return fmt.Errorf("p.API.SetProfileImage: %w", err)
	}

	ico, err := command.GetIconData(p.API, "assets/orbit-467260.svg")
	if err != nil {
		return fmt.Errorf("command.GetIconData: %w", err)
	}

	if err = p.API.RegisterCommand(p.getCommand(ico)); err != nil {
		return fmt.Errorf("p.API.RegisterCommand: %w", err)
	}

	j, err := cluster.Schedule(p.API, "monitoring-report", p.NextWaitMonitoringReportInterval, p.monitoringReport)
	if err != nil {
		return fmt.Errorf("cluster.Schedule: %w", err)
	}

	p.monitoringReportJob = j

	return nil
}

func (p *Plugin) OnDeactivate() error {
	if err := p.monitoringReportJob.Close(); err != nil {
		return fmt.Errorf("p.monitoringReportJob.Close: %w", err)
	}

	return nil
}
