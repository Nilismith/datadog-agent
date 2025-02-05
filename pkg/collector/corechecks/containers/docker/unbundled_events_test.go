// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build docker

package docker

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/comp/core/tagger/taggerimpl"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/util/containers"
	"github.com/DataDog/datadog-agent/pkg/util/docker"
	"github.com/stretchr/testify/assert"
)

func TestUnbundledEventsTransform(t *testing.T) {
	ts := time.Now()
	containerID := "foobar"
	containerName := "foo"
	containerTags := []string{"image_name:foo", "image_tag:latest"}
	imageName := "foo:latest"
	hostname := "test-host"

	fakeTagger := taggerimpl.SetupFakeTagger(t)
	defer fakeTagger.ResetTagger()

	fakeTagger.SetTags(
		containers.BuildTaggerEntityName(containerID), "-",
		containerTags, []string{}, []string{}, []string{},
	)

	tests := []struct {
		name     string
		event    *docker.ContainerEvent
		expected []event.Event
	}{
		{
			name: "event is filtered out",
			event: &docker.ContainerEvent{
				ContainerID:   containerID,
				ContainerName: containerName,
				ImageName:     imageName,
				Action:        "create",
				Timestamp:     ts,
			},
			expected: nil,
		},
		{
			name: "event is filtered out",
			event: &docker.ContainerEvent{
				ContainerID:   containerID,
				ContainerName: containerName,
				ImageName:     imageName,
				Action:        "oom",
				Timestamp:     ts,
			},
			expected: []event.Event{
				{
					Title:          "Container foobar: oom",
					Text:           "Container foobar (running image \"foo:latest\"): oom",
					AlertType:      event.EventAlertTypeError,
					AggregationKey: "docker:foobar",
					Ts:             ts.Unix(),
					Host:           hostname,
					SourceTypeName: "docker",
					EventType:      "docker",
					Priority:       event.EventPriorityNormal,
					Tags: []string{
						"image_name:foo",
						"image_tag:latest",
						"event_type:oom",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := newUnbundledTransformer(hostname, []string{"oom", "kill"})

			events, errors := transformer.Transform([]*docker.ContainerEvent{tt.event})

			assert.Empty(t, errors)
			assert.Equal(t, tt.expected, events)
		})
	}
}
