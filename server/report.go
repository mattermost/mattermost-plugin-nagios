package main

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-nagios/go-nagios/nagios"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
)

func (p *Plugin) sendMessages(channelID string, messages ...string) error {
	var firstID string

	for i, m := range messages {
		post := &model.Post{
			UserId:    p.botUserID,
			ChannelId: channelID,
			RootId:    firstID,
			Message:   m,
		}

		created, err := p.API.CreatePost(post)
		if err != nil {
			return fmt.Errorf("p.API.CreatePost: %w", err)
		}

		if i == 0 {
			firstID = created.Id
		}
	}

	return nil
}

func (p *Plugin) sendMonitoringReport(channelID string) error {
	hostCountReq := nagios.HostCountRequest{
		GeneralHostRequest: nagios.GeneralHostRequest{
			FormatOptions: nagios.FormatOptions{
				Enumerate: true,
			},
		}}

	var hostCount nagios.HostCount

	if err := p.client.Query(hostCountReq, &hostCount); err != nil {
		return fmt.Errorf("client.Query: %w", err)
	}

	hostListReq := nagios.HostListRequest{
		GeneralHostRequest: nagios.GeneralHostRequest{
			FormatOptions: nagios.FormatOptions{
				Enumerate: true,
			},
		}}

	var hostList nagios.HostList

	if err := p.client.Query(hostListReq, &hostList); err != nil {
		return fmt.Errorf("client.Query: %w", err)
	}

	serviceCountReq := nagios.ServiceCountRequest{
		GeneralServiceRequest: nagios.GeneralServiceRequest{
			FormatOptions: nagios.FormatOptions{
				Enumerate: true,
			},
		}}

	var serviceCount nagios.ServiceCount

	if err := p.client.Query(serviceCountReq, &serviceCount); err != nil {
		return fmt.Errorf("client.Query: %w", err)
	}

	serviceListReq := nagios.ServiceListRequest{
		GeneralServiceRequest: nagios.GeneralServiceRequest{
			FormatOptions: nagios.FormatOptions{
				Enumerate: true,
			},
		}}

	var serviceList nagios.ServiceList

	if err := p.client.Query(serviceListReq, &serviceList); err != nil {
		return fmt.Errorf("client.Query: %w", err)
	}

	if err := p.sendMessages(
		channelID,
		reportPreamble(time.Now()),
		formatHostCount(hostCount),
		formatHostList(hostList),
		formatServiceCount(serviceCount),
		formatServiceList(serviceList)); err != nil {
		return fmt.Errorf("p.sendMessages: %w", err)
	}

	return nil
}

func (p *Plugin) NextWaitMonitoringReportInterval(
	now time.Time,
	metadata cluster.JobMetadata) time.Duration {
	interval, err := getReportFrequency(p.API)
	if err != nil {
		p.API.LogError("getReportFrequency", logErrorKey, err)
		return time.Minute // Return a reasonable interval.
	}

	if since := now.Sub(metadata.LastFinished); interval > since {
		return interval - since
	}

	return 0
}

func (p *Plugin) monitoringReport() {
	c, err := getReportChannel(p.API)
	if err != nil {
		p.API.LogError("getReportChannel", logErrorKey, err)
		return
	}

	if c == "" { // fast path, there is no subscription.
		p.API.LogDebug("monitoringReport: no subscription")
		return
	}

	if err := p.sendMonitoringReport(c); err != nil {
		p.API.LogError("sendMonitoringReport", logErrorKey, err)
	}
}
