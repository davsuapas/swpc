/*
 *   Copyright (c) 2022 CARISA
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package micro

import "github.com/swpoolcontroller/pkg/sockets"

// Hub manages the socket pool and distribute messages
type Hub interface {
	// Send sends message to the all clients into hub
	Send(message string)

	// Status request hub status via channel
	Status(resp chan sockets.Status)

	// Config sends the config to the hub
	Config(cnf sockets.Config)
}
