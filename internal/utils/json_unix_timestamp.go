// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"encoding/json"
	"time"
)

type UnixTimestamp struct {
	time.Time
}

func (t *UnixTimestamp) UnmarshalJSON(data []byte) error {
	var timestamp int64
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return err
	}

	t.Time = time.Unix(timestamp, 0)
	return nil
}
