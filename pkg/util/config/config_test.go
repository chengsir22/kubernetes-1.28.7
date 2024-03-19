/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"context"
	"reflect"
	"testing"
)

func TestConfigurationChannels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := NewMux(nil)
	channelOne := mux.ChannelWithContext(ctx, "one")
	if channelOne != mux.ChannelWithContext(ctx, "one") {
		t.Error("Didn't get the same muxuration channel back with the same name")
	}
	channelTwo := mux.ChannelWithContext(ctx, "two")
	if channelOne == channelTwo {
		t.Error("Got back the same muxuration channel for different names")
	}
}

type MergeMock struct {
	source string
	update interface{}
	t      *testing.T
}

func (m MergeMock) Merge(source string, update interface{}) error {
	if m.source != source {
		m.t.Errorf("Expected %s, Got %s", m.source, source)
	}
	if !reflect.DeepEqual(m.update, update) {
		m.t.Errorf("Expected %s, Got %s", m.update, update)
	}
	return nil
}

func TestMergeInvoked(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	merger := MergeMock{"one", "test", t}
	mux := NewMux(&merger)
	mux.ChannelWithContext(ctx, "one") <- "test"
}

func TestMergeFuncInvoked(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan bool)
	mux := NewMux(MergeFunc(func(source string, update interface{}) error {
		if source != "one" {
			t.Errorf("Expected %s, Got %s", "one", source)
		}
		if update.(string) != "test" {
			t.Errorf("Expected %s, Got %s", "test", update)
		}
		ch <- true
		return nil
	}))
	mux.ChannelWithContext(ctx, "one") <- "test"
	<-ch
}

func TestSimultaneousMerge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan bool, 2)
	mux := NewMux(MergeFunc(func(source string, update interface{}) error {
		switch source {
		case "one":
			if update.(string) != "test" {
				t.Errorf("Expected %s, Got %s", "test", update)
			}
		case "two":
			if update.(string) != "test2" {
				t.Errorf("Expected %s, Got %s", "test2", update)
			}
		default:
			t.Errorf("Unexpected source, Got %s", update)
		}
		ch <- true
		return nil
	}))
	source := mux.ChannelWithContext(ctx, "one")
	source2 := mux.ChannelWithContext(ctx, "two")
	source <- "test"
	source2 <- "test2"
	<-ch
	<-ch
}

func TestBroadcaster(t *testing.T) {
	b := NewBroadcaster()
	b.Notify(struct{}{})

	ch := make(chan bool, 2)
	b.Add(ListenerFunc(func(object interface{}) {
		if object != "test" {
			t.Errorf("Expected %s, Got %s", "test", object)
		}
		ch <- true
	}))
	b.Add(ListenerFunc(func(object interface{}) {
		if object != "test" {
			t.Errorf("Expected %s, Got %s", "test", object)
		}
		ch <- true
	}))
	b.Notify("test")
	<-ch
	<-ch
}
