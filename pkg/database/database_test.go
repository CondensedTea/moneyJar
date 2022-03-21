package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mergeDuplicateAccounts(t *testing.T) {
	type args struct {
		accounts []Account
	}
	tests := []struct {
		name               string
		args               args
		wantResultAccounts []Account
	}{
		{
			name: "default",
			args: args{
				accounts: []Account{
					{FromUser: 1, FromUserName: "user1", ToUser: 2, ToUserName: "user2", Balance: 100},
					{FromUser: 2, FromUserName: "user2", ToUser: 1, ToUserName: "user1", Balance: 50, IsFlipped: true},
				},
			},
			wantResultAccounts: []Account{
				{FromUser: 1, FromUserName: "user1", ToUser: 2, ToUserName: "user2", Balance: 50},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResultAccounts := mergeDuplicateAccounts(tt.args.accounts); !assert.EqualValues(t, tt.wantResultAccounts, gotResultAccounts) {
				t.Errorf("mergeDuplicateAccounts() = %v, want %v", gotResultAccounts, tt.wantResultAccounts)
			}
		})
	}
}
