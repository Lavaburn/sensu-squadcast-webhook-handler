package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := corev2.FixtureEvent("entity1", "check1")
	assert.Error(CheckArgs(event))
	plugin.APIURL = "InvalidURL"
	assert.Error(CheckArgs(event))
	plugin.APIURL = "http://sensu.example.com:3000"
	assert.NoError(CheckArgs(event))
}

func TestSendEventToSquadcast(t *testing.T) {
	testcases := []struct {
		sensuStatus  uint32
		webhookStatus string
	}{
		{0, "resolve"},
		{1, "trigger"},
		{2, "trigger"},
		{127, "trigger"},
	}

	plugin.Message = "{{.Entity.Name}}/{{.Check.Name}}"
	plugin.Description = "{{.Check.Output}}"
	
	checkOutput := "Check failed!"

	for _, tc := range testcases {
		assert := assert.New(t)
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Status = tc.sensuStatus
		event.Check.Output = checkOutput

		var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(err)
			
			msg := &SQEvent{}
			err = json.Unmarshal(body, msg)
			require.NoError(t, err)
			
			expectedMessage := "entity1/check1"
			assert.Equal(expectedMessage, msg.Message)
			
			expectedStatus := tc.webhookStatus
			assert.Equal(expectedStatus, msg.Status)	

			assert.Equal(checkOutput, msg.Description)
			
			w.WriteHeader(http.StatusOK)
		}))

		_, err := url.ParseRequestURI(test.URL)
		require.NoError(t, err)
		plugin.APIURL = test.URL
		assert.NoError(SendEventToSquadcast(event))
	}
}

func TestSendEventWithTemplateToSquadcast(t *testing.T) {
	testcases := []struct {
		sensuStatus  uint32
		webhookStatus string
	}{
		{0, "resolve"},
		{1, "trigger"},
	}

	plugin.Message = "{{.Entity.Name}}/{{.Check.Name}}"
	plugin.Template = "squadcast.tpl"
	
	checkOutput := "Check failed!"
	expectedDescription := "TODO:\nCheck failed!"

	for _, tc := range testcases {
		assert := assert.New(t)
		event := corev2.FixtureEvent("entity1", "check1")
		event.Check.Status = tc.sensuStatus
		event.Check.Output = checkOutput

		var test = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(err)
			
			msg := &SQEvent{}
			err = json.Unmarshal(body, msg)
			require.NoError(t, err)
			
			expectedMessage := "entity1/check1"
			assert.Equal(expectedMessage, msg.Message)
			
			expectedStatus := tc.webhookStatus
			assert.Equal(expectedStatus, msg.Status)	

			assert.Equal(expectedDescription, msg.Description)
			
			w.WriteHeader(http.StatusOK)
		}))

		_, err := url.ParseRequestURI(test.URL)
		require.NoError(t, err)
		plugin.APIURL = test.URL
		assert.NoError(SendEventToSquadcast(event))
	}
}
