package qust

import (
	"github.com/qlik-oss/qliksense-operator/pkg/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sigs.k8s.io/kustomize/api/types"
	"testing"
)

func TestCreateSupperSecretSelectivePatch(t *testing.T) {
	reader := setupCr(t)
	cfg, err := config.ReadCRConfigFromFile(reader)
	if err != nil {
		t.Fatalf("error reading config from file")
	}
	spMap := createSupperSecretSelectivePatch(cfg.Secrets)
	sp := spMap["qliksense"]
	if sp.ApiVersion != "qlik.com/v1" {
		t.Fail()
	}
	if sp.Kind != "SelectivePatch" {
		t.Fail()
	}
	if sp.Metadata["name"] != "qliksense-operator-secrets" {
		t.Fail()
	}
	if sp.Patches[0].Target.LabelSelector != "app=qliksense" || sp.Patches[0].Target.Kind != "SuperSecret" {
		t.Fail()
	}
	ss := &config.SupperSecret{
		ApiVersion: "qlik.com/v1",
		Kind:       "SuperSecret",
		Metadata: map[string]string{
			"name": "qliksense-secrets",
		},
		StringData: map[string]string{
			"mongoDbUri": "mongo://mongo:3307",
		},
	}
	ss2 := &config.SupperSecret{}
	yaml.Unmarshal([]byte(sp.Patches[0].Patch), ss2)

	if !reflect.DeepEqual(ss, ss2) {
		t.Fail()
		t.Log("expected selectivePatch: ", ss)
		t.Log("Actual SelectivePatch: ", ss2)
	}
}

func TestProcessCrSecrets(t *testing.T) {
	reader := setupCr(t)
	cfg, err := config.ReadCRConfigFromFile(reader)
	if err != nil {
		t.Fatalf("error reading config from file")
	}

	td, dir := createManifestsStructure(t)

	cfg.ManifestsRoot = dir
	ProcessCrSecrets(cfg)
	content, _ := ioutil.ReadFile(filepath.Join(dir, ".operator", "secrets", "qliksense.yaml"))

	sp := getSuperSecretSPTemplate("qliksense")
	scm := getSuperSecretTemplate("qliksense")
	scm.StringData = map[string]string{
		"mongoDbUri": "mongo://mongo:3307",
	}
	phb, _ := yaml.Marshal(scm)
	sp.Patches = []types.Patch{
		types.Patch{
			Patch:  string(phb),
			Target: getSelector("SuperSecret", "qliksense"),
		},
	}
	spOut := &config.SelectivePatch{}
	yaml.Unmarshal(content, spOut)
	if !reflect.DeepEqual(sp, spOut) {
		t.Fail()
	}

	td()
}
