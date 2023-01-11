// Copyright 2022 iLogtail Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protocol

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/ilogtail/pkg/models"
)

var streamPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	}}

func (c *Converter) ConvertToRawStream(groupEvents *models.PipelineGroupEvents, targetFields []string) (stream [][]byte, values []map[string]string, err error) {
	if len(groupEvents.Events) == 0 {
		return nil, nil, nil
	}

	byteStream := streamPool.Get().(*bytes.Buffer)
	defer streamPool.Put(byteStream)
	defer byteStream.Reset()
	for idx, event := range groupEvents.Events {
		eventType := event.GetType()
		if eventType != models.EventTypeByteArray {
			return nil, nil, fmt.Errorf("unsupported event type %v", eventType)
		}
		if idx != 0 {
			byteStream.Write([]byte{'\n'})
		}
		byteStream.Write(event.(models.ByteArray))
	}

	var targetValues map[string]string
	if len(targetFields) > 0 {
		targetValues = findTargetFieldsInGroup(targetFields, groupEvents.Group)
	}
	return [][]byte{byteStream.Bytes()}, []map[string]string{targetValues}, nil
}

func findTargetFieldsInGroup(targetFields []string, group *models.GroupInfo) map[string]string {
	if len(targetFields) == 0 {
		return nil
	}

	targetKVs := make(map[string]string, len(targetFields))

	for _, field := range targetFields {
		var tagName, tagValue string
		if strings.HasPrefix(field, targetGroupMetadataPrefix) {
			tagName = field[len(targetGroupMetadataPrefix):]
			tagValue = group.GetMetadata().Get(tagName)
		} else if strings.HasPrefix(field, targetGroupTagsPrefix) {
			tagName = field[len(targetGroupTagsPrefix):]
			tagValue = group.GetTags().Get(tagName)
		}
		targetKVs[field] = tagValue
	}

	return targetKVs
}