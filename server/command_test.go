package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
)

func Test_getAutoCompleteDesc(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]commandHandlerFunc
		want string
	}{
		{
			name: "empty",
			m:    map[string]commandHandlerFunc{},
			want: "",
		},
		{
			name: "one function",
			m: map[string]commandHandlerFunc{
				"test 1": nil,
			},
			want: "Available commands: test 1",
		},
		{
			name: "few functions",
			m: map[string]commandHandlerFunc{
				"test 1": nil,
				"test 2": nil,
				"test 3": nil,
			},
			want: "Available commands: test 1, test 2, test 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAutoCompleteDesc(tt.m); got != tt.want {
				t.Errorf("getAutoCompleteDesc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHint(t *testing.T) {
	type args struct {
		before rune
		after  rune
		keys   []string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty 1",
			args: args{
				before: '<',
				after:  '>',
			},
			want: "<>",
		},
		{
			name: "empty 2",
			args: args{
				before: '<',
				after:  '>',
				keys:   []string{},
			},
			want: "<>",
		},
		{
			name: "one hint",
			args: args{
				before: '[',
				after:  ']',
				keys:   []string{"test 1"},
			},
			want: "[test 1]",
		},
		{
			name: "few hints",
			args: args{
				before: '{',
				after:  '}',
				keys:   []string{"test 1", "test 2", "test 3"},
			},
			want: "{test 1|test 2|test 3}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHint(tt.args.before, tt.args.after, tt.args.keys...); got != tt.want {
				t.Errorf("getHint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func Test_parseCommandArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           *model.CommandArgs
		wantCommand    string
		wantAction     string
		wantParameters []string
	}{
		{
			name:           "empty",
			args:           &model.CommandArgs{},
			wantCommand:    "",
			wantAction:     "",
			wantParameters: nil,
		},
		{
			name: "command only",
			args: &model.CommandArgs{
				Command: "command",
			},
			wantCommand:    "command",
			wantAction:     "",
			wantParameters: nil,
		},
		{
			name: "command and action only",
			args: &model.CommandArgs{
				Command: "command action",
			},
			wantCommand:    "command",
			wantAction:     "action",
			wantParameters: nil,
		},
		{
			name: "command, action and one parameter",
			args: &model.CommandArgs{
				Command: "command action parameter-1",
			},
			wantCommand:    "command",
			wantAction:     "action",
			wantParameters: []string{"parameter-1"},
		},
		{
			name: "command, action and two parameters",
			args: &model.CommandArgs{
				Command: "command action parameter-1 parameter-2",
			},
			wantCommand:    "command",
			wantAction:     "action",
			wantParameters: []string{"parameter-1", "parameter-2"},
		},
		{
			name: "command, action and three parameters",
			args: &model.CommandArgs{
				Command: "command action parameter-1 parameter-2 parameter-3",
			},
			wantCommand:    "command",
			wantAction:     "action",
			wantParameters: []string{"parameter-1", "parameter-2", "parameter-3"},
		},
		{
			name: "command, action and four parameters",
			args: &model.CommandArgs{
				Command: "command action parameter-1 parameter-2 parameter-3 parameter-4",
			},
			wantCommand:    "command",
			wantAction:     "action",
			wantParameters: []string{"parameter-1", "parameter-2", "parameter-3 parameter-4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, action, parameters := parseCommandArgs(tt.args)
			if command != tt.wantCommand {
				t.Errorf("parseCommandArgs() command = %v, want %v", command, tt.wantCommand)
			}
			if action != tt.wantAction {
				t.Errorf("parseCommandArgs() action = %v, want %v", action, tt.wantAction)
			}
			if !stringsEqual(parameters, tt.wantParameters) {
				t.Errorf("parseCommandArgs() parameters = %v, want %v", parameters, tt.wantParameters)
			}
		})
	}
}
