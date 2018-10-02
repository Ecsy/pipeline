// Copyright © 2018 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	evbus "github.com/asaskevich/EventBus"
)

// eventBus interface separates the actual event handling implementation.
type EventBus interface {
	Publish(topic string, args ...interface{})
	Subscribe(topic string, fn interface{}) error
	SubscribeAsync(topic string, fn interface{}, transaction bool) error
}

// Bus is the global EventBus dispatcher object
var Bus EventBus = evbus.New()
