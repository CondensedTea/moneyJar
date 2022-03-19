package core

import (
	"reflect"
	"testing"
)

func Test_parseMentions(t *testing.T) {
	type args struct {
		mentions string
	}
	tests := []struct {
		name          string
		args          args
		wantUsernames []string
		wantErr       bool
	}{
		{
			name:          "single mention",
			args:          args{mentions: "@test_user"},
			wantUsernames: []string{"test_user"},
		},
		{
			name:          "multiple mentions",
			args:          args{mentions: "@test_user, @test_user2"},
			wantUsernames: []string{"test_user", "test_user2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsernames, err := parseMentions(tt.args.mentions)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMentions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUsernames, tt.wantUsernames) {
				t.Errorf("parseMentions() gotUsernames = %v, want %v", gotUsernames, tt.wantUsernames)
			}
		})
	}
}
