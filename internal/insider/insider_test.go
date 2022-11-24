// Package insider should runs every day

// on 8:00 AM MSK (00:00 in finviz.com)

// to get the latest data for sell and buy transactions.

package insider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactions_LastDay(t *testing.T) {
	tests := []struct {
		name string
		tr   Transactions
		want int
	}{
		{
			name: "success",
			tr: Transactions{
				{
					SEC: SEC{
						NotificationDate: time.Now().Add(-time.Hour * 24),
					},
				},
				{
					SEC: SEC{
						NotificationDate: time.Now(),
					},
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tr.lastDay()
			assert.Equal(t, len(got), tt.want)
		})
	}
}
