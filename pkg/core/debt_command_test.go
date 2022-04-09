package core

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test_reDebtPayload(t *testing.T) {
	var match []string

	var (
		singleMention         = `-100 gel @test`
		expectedSingleMention = []string{singleMention, "-100", "gel", "@test", ""}
	)
	match = reDebtPayload.FindStringSubmatch(singleMention)
	assert.Equal(t, expectedSingleMention, match)

	var (
		multipleMention         = `100 gel @test1 @test2`
		expectedMultipleMention = []string{multipleMention, "100", "gel", "@test1 @test2", ""}
	)
	match = reDebtPayload.FindStringSubmatch(multipleMention)
	assert.Equal(t, expectedMultipleMention, match)

	var (
		comment         = `100 gel @test; comment text`
		expectedComment = []string{comment, "100", "gel", "@test", "comment text"}
	)
	match = reDebtPayload.FindStringSubmatch(comment)
	assert.Equal(t, expectedComment, match)
}
