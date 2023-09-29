package insider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBrowser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		fileName  string
		expectLen int
	}{
		{
			name:      "Should Parse Transactions",
			fileName:  "testdata/transactions.html",
			expectLen: 199,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileData, err := os.ReadFile(tt.fileName)
			if err != nil {
				t.Fatalf("Could not read file: %v", err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Write(fileData)
			}))

			browser := New(nil)

			transactions, err := browser.parse(server.URL)
			assert.NoError(t, err)
			assert.Len(t, transactions, tt.expectLen)
		})
	}
}

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
