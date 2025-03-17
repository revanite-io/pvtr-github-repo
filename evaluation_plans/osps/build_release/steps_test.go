package build_release

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/rhysd/actionlint"
)


var encoded string = "bmFtZTogT1NQUyBCYXNlbGluZSBTY2FuCgpvbjogW3dvcmtmbG93X2Rpc3Bh\ndGNoXQoKam9iczoKICBzY2FuOgogICAgcnVucy1vbjogdWJ1bnR1LWxhdGVz\ndAoKICAgIHN0ZXBzOgogICAgICAtIG5hbWU6IENoZWNrb3V0IHJlcG9zaXRv\ncnkKICAgICAgICB1c2VzOiBhY3Rpb25zL2NoZWNrb3V0QHY0CgogICAgICAt\nIG5hbWU6IFB1bGwgdGhlIHB2dHItZ2l0aHViLXJlcG8gaW1hZ2UKICAgICAg\nICBydW46IGRvY2tlciBwdWxsIGVkZGlla25pZ2h0L3B2dHItZ2l0aHViLXJl\ncG86bGF0ZXN0CgogICAgICAtIG5hbWU6IEFkZCBHaXRIdWIgU2VjcmV0IHRv\nIGNvbmZpZyBmaWxlIHNvIGl0IGlzIHByb3RlY3RlZCBpbiBvdXRwdXRzCiAg\nICAgICAgcnVuOiB8CiAgICAgICAgICBzZWQgLWkgJ3Mve3sgVE9LRU4gfX0v\nJHt7IHNlY3JldHMuVE9LRU4gfX0vZycgJHt7IGdpdGh1Yi53b3Jrc3BhY2Ug\nfX0vLmdpdGh1Yi9wdnRyLWNvbmZpZy55bWwKCiAgICAgIC0gbmFtZTogU2Nh\nbiBhbGwgcmVwb3Mgc3BlY2lmaWVkIGluIC5naXRodWIvcHZ0ci1jb25maWcu\neW1sCiAgICAgICAgcnVuOiB8CiAgICAgICAgICBkb2NrZXIgcnVuIC0tcm0g\nXAogICAgICAgICAgICAtdiAke3sgZ2l0aHViLndvcmtzcGFjZSB9fS8uZ2l0\naHViL3B2dHItY29uZmlnLnltbDovLnByaXZhdGVlci9jb25maWcueW1sIFwK\nICAgICAgICAgICAgLXYgJHt7IGdpdGh1Yi53b3Jrc3BhY2UgfX0vZG9ja2Vy\nX291dHB1dDovZXZhbHVhdGlvbl9yZXN1bHRzIFwKICAgICAgICAgICAgZWRk\naWVrbmlnaHQvcHZ0ci1naXRodWItcmVwbzpsYXRlc3QK\n"



func TestParse( t *testing.T) {

	decoded, _ := base64.StdEncoding.DecodeString(encoded)
	fmt.Printf("Decoded String: %s", decoded)

	workflow, err := actionlint.Parse(decoded)

	if err != nil {
		t.Errorf("Error parsing workflow: %v", err)
	}

	assert.Equal(t, workflow.Name.Value, "OSPS Baseline Scan", "workflow parsing failed")

}

func TestCicdSanitizedInputParameters(t *testing.T) {

	//Need to mock test payloads
	//payloads with various env inputs
	workflow, error := actionlint.Parse([]byte(`github.event.issue.title`))

	if error != nil {
		t.Errorf("Error parsing workflow: %v", error)
	}

	fmt.Printf("Workflow: %v", workflow)
		
}