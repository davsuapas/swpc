/*
 *   Copyright (c) 2022 ELIPCERO
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

package hub_test

import (
	"errors"
	"testing"

	"github.com/swpoolcontroller/internal/hub"
	"github.com/swpoolcontroller/pkg/iot"
	"go.uber.org/zap"
)

var (
	errTrace = errors.New("trace error")
)

func TestTrace_Register(t *testing.T) {
	t.Parallel()

	h := hub.NewTrace(zap.NewExample())

	h.Register()

	h.Trace <- iot.Trace{
		Level:   iot.DebugLevel,
		Message: "msg",
	}
	h.Trace <- iot.Trace{
		Level:   iot.InfoLevel,
		Message: "msg",
	}
	h.Trace <- iot.Trace{
		Level:   iot.WarnLevel,
		Message: "msg",
	}
	h.Error <- errTrace
}
