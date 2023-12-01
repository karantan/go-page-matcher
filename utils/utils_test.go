package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	assert := assert.New(t)

	assert.True(Exists(RootDir() + "/fixtures/www/foo.com/config.json"))
	assert.False(Exists(RootDir() + "/fixtures/www/foo.com/_config.json"))
	assert.False(Exists(RootDir() + "/fixtures/www/bar.com/config.json"))
}

func TestGetDomainFromURL(t *testing.T) {
	type args struct {
		websiteURL string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"url with https", args{"https://foo.com"}, "foo.com", false},
		{"url with https and www", args{"https://www.foo.com"}, "foo.com", false},
		{"no ssl", args{"http://foo.com"}, "foo.com", false},
		{"just domain", args{"foo.com"}, "foo.com", false},
		{"empty", args{""}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDomainFromURL(tt.args.websiteURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDomainFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDomainFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
