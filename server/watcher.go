package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-plugin-nagios/internal/watcher"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func formatChange(change watcher.Change) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("**%s** has been modified", change.Name))
	b.WriteString(" (-previous +current):\n\n")

	b.WriteString("```diff\n")

	if len(change.Diff)+b.Len()+3 > model.PostMessageMaxRunesV1*4 {
		b.WriteString("File has been changed, but the diff is too long.\n")
	} else {
		b.WriteString(change.Diff)
	}

	b.WriteString("```")

	return b.String()
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	token := p.getConfiguration().Token

	const notConfigured = "This functionality is not configured."

	if token == "" {
		http.Error(w, notConfigured, http.StatusNotImplemented)
		return
	}

	if token != r.Header.Get(watcher.TokenHeader) {
		p.API.LogWarn("Changes handler called, but authentication failed", "ctx", c)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	channelID, err := getChangesChannel(p.API)
	if err != nil {
		p.API.LogError("getChangesChannel", logErrorKey, err)
		return
	}

	if channelID == "" { // fast path, there is no subscription.
		p.API.LogWarn("Changes handler called, but there is no subscription")
		http.Error(w, notConfigured, http.StatusNotImplemented)

		return
	}

	var change watcher.Change

	if err := json.NewDecoder(r.Body).Decode(&change); err != nil {
		p.API.LogError("Decode", logErrorKey, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	post := &model.Post{
		UserId:    p.botUserID,
		ChannelId: channelID,
		Message:   formatChange(change),
	}

	if _, err := p.API.CreatePost(post); err != nil {
		p.API.LogError("CreatePost", logErrorKey, err)
	}
}
