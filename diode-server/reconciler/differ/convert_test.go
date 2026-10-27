package differ_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/differ"
)

// The plugin's tags_resolved survives both plan conversions, so consumers
// of the change set can name the tags an object holds.
func TestConvert_KeepsTagsResolved(t *testing.T) {
	const tags = `[{"ref":7,"name":"prod","slug":"prod"},{"ref":"new_object:extras.tag:u1","name":"lab","slug":"lab"}]`
	var cs netboxdiodeplugin.ChangeSet
	require.NoError(t, json.Unmarshal([]byte(`{"id":"cs-1","changes":[
		{"id":"c1","change_type":"update","object_type":"dcim.device","object_id":3,"before":{"tags":[7]},"data":{"tags":[7,"new_object:extras.tag:u1"]},"tags_resolved":`+tags+`},
		{"id":"c2","change_type":"noop","object_type":"dcim.platform","object_id":4,"before":{}}
	]}`), &cs))

	planned, err := differ.ConvertBulkPlanResult(netboxdiodeplugin.BulkPlanResult{ID: "1", ChangeSet: &cs}, "dcim.device")
	require.NoError(t, err)
	applied, planErr, applyErr := differ.ConvertBulkPlanApplyResult(netboxdiodeplugin.BulkPlanApplyResult{ID: "1", ChangeSet: &cs}, "dcim.device")
	require.NoError(t, planErr)
	require.NoError(t, applyErr)

	for _, got := range [][]json.RawMessage{
		{planned.Changes[0].TagsResolved, planned.Changes[1].TagsResolved},
		{applied.Changes[0].TagsResolved, applied.Changes[1].TagsResolved},
	} {
		require.JSONEq(t, tags, string(got[0]))
		require.Empty(t, got[1], "a plugin that sends no tags_resolved leaves it empty")
	}
}
